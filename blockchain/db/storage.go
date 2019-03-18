package db

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/denisskin/goldb"
	"github.com/denisskin/gosync"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/commons/bignum"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/likecoin-pro/likecoin/crypto/patricia"
	"github.com/likecoin-pro/likecoin/object"
	"github.com/likecoin-pro/likecoin/services/mempool"
)

type BlockchainStorage struct {
	Cfg     *blockchain.Config
	db      *goldb.Storage
	Mempool *mempool.Storage

	// blocks
	mxW          sync.Mutex
	mxR          sync.RWMutex
	lastBlock    *blockchain.Block //
	stat         *Statistic        //
	cacheHeaders *gosync.Cache     // blockNum => *BlockHeader
	cacheTxs     *gosync.Cache     // blockNum => []*Transaction
	cacheIdxTx   *gosync.Cache     // idxKey => *Transaction
	middleware   []Middleware      //
}

type Middleware func(*goldb.Transaction, *blockchain.Block)

const (
	// tables
	dbTabHeaders   = 0x01 // (blockNum) => BlockHeader
	dbTabTxs       = 0x02 // (blockNum, txIdx) => Transaction
	dbTabChainTree = 0x03 //
	dbTabStateTree = 0x04 // (asset, addr) => sateValue
	dbTabStat      = 0x05 // (ts) => Statistic

	// indexes
	dbIdxTxID          = 0x20 // (txID)                        => txNum
	dbIdxAsset         = 0x21 // (asset, txNum)                => sateValue
	dbIdxAssetAddr     = 0x22 // (asset, addr, txNum)          => sateValue
	dbIdxAssetAddrMemo = 0x23 // (asset, addr, addrTag, txNum) => sateValue
	dbIdxUsers         = 0x24 // (userID) => txUID
	dbIdxSourceTx      = 0x25 // (providerID, sourceID, txUID) => nil
	dbIdxSourceAddr    = 0x26 // (providerID, sourceID, addr)  => total supply by addr
	dbIdxInvites       = 0x27 // (userID, txNum)               => invitedUserID
	dbIdxSrcInvites    = 0x28 // (userID, txNum)               => invitedUserID
	dbIdxBalances      = 0x29 // (asset, addr)                 => balance
)

var (
	ErrBlockNotFound         = errors.New("block not found")
	errTxHasBeenRegistered   = errors.New("tx has been registered")
	errTxNotFound            = errors.New("tx not found")
	errUserHasBeenRegistered = errors.New("user has been registered")
	errUserNotFound          = errors.New("user not found")
	ErrAddrNotFound          = errors.New("address not found")
	errIncorrectAddress      = errors.New("incorrect address")
	errIncorrectAssetVal     = errors.New("incorrect asset value")
	errIncorrectTxState      = errors.New("incorrect tx state")
	errIncorrectChainRoot    = errors.New("incorrect chain root")
	errIncorrectStateRoot    = errors.New("incorrect state root")
)

func NewBlockchainStorage(cfg *blockchain.Config) (s *BlockchainStorage) {
	s = &BlockchainStorage{
		Cfg:          cfg,
		db:           goldb.NewStorage(cfg.DataDir, nil),
		cacheHeaders: gosync.NewCache(100000),
		cacheTxs:     gosync.NewCache(100000),
		cacheIdxTx:   gosync.NewCache(30000),
		Mempool:      mempool.NewStorage(),
	}

	if cfg.VacuumDB {
		s.db.Vacuum()
	}

	// query last block
	if b, err := s.queryLastBlock(); err != nil {
		panic(err)
	} else {
		s.lastBlock = b
	}
	// query actual totals
	s.stat = &Statistic{}
	if err := s.db.QueryValue(goldb.NewQuery(dbTabStat).Last(), &s.stat); err != nil {
		panic(err)
	}

	return
}

func (s *BlockchainStorage) Close() (err error) {
	return s.db.Close()
}

func (s *BlockchainStorage) Drop() (err error) {
	s.db.Close()
	return s.db.Drop()
}

func (s *BlockchainStorage) VacuumDB() error {
	return s.db.Vacuum()
}

func (s *BlockchainStorage) DBStorage() *goldb.Storage {
	return s.db
}

func (s *BlockchainStorage) AddMiddleware(fn Middleware) {
	s.middleware = append(s.middleware, fn)
}

func (s *BlockchainStorage) ChainTree() *patricia.Tree {
	return patricia.NewTree(patricia.NewMemoryStorage(patricia.NewSubStorage(s.db, goldb.Key(dbTabChainTree))))
}

func (s *BlockchainStorage) StateTree() *patricia.Tree {
	return patricia.NewTree(patricia.NewMemoryStorage(patricia.NewSubStorage(s.db, goldb.Key(dbTabStateTree))))
}

// State returns state struct from db
func (s *BlockchainStorage) State() *state.State {
	return state.NewState(s.Cfg.ChainID, func(a assets.Asset, addr crypto.Address) (v bignum.Int) {
		if err := s.db.QueryValue(goldb.NewQuery(dbIdxAssetAddr, a, addr).Last(), &v); err != nil {
			panic(err)
		}
		return
	})
}

//----------------- put block --------------------------
// open db.transaction; verify block; save block and index-records
func (s *BlockchainStorage) PutBlock(blocks ...*blockchain.Block) error {
	if len(blocks) == 0 {
		return nil
	}
	// lock tx exec
	s.mxW.Lock()
	defer s.mxW.Unlock()

	// verify blocks
	lastBlockHeader := s.lastBlock.BlockHeader
	for _, block := range blocks {
		if err := block.Verify(lastBlockHeader, s.Cfg); err != nil {
			return err
		}
		lastBlockHeader = block.BlockHeader
	}

	var blockStat = s.stat
	var txsIDs []uint64

	// open db transaction
	err := s.db.Exec(func(tr *goldb.Transaction) {

		stateTree := patricia.NewSubTree(tr, goldb.Key(dbTabStateTree))
		chainTree := patricia.NewSubTree(tr, goldb.Key(dbTabChainTree))

		for _, block := range blocks {

			// init new block statistic
			blockStat = blockStat.New(block.Num, len(block.Txs))

			// add index on transactions
			for txIdx, tx := range block.Txs {

				txID := tx.ID()
				txUID := encodeTxUID(block.Num, txIdx)
				txsIDs = append(txsIDs, txID)

				// check transaction by txID
				if id, _ := tr.GetID(goldb.Key(dbIdxTxID, txID)); id != 0 {
					tr.Fail(errTxHasBeenRegistered)
				}

				if s.Cfg.VerifyTxsLevel >= blockchain.VerifyTxLevel1 {

					//-- verify sender signature
					if err := tx.Verify(s.Cfg); err != nil {
						tr.Fail(err)
					}

					//-- verify transaction state
					// make state by dbTransaction
					st := state.NewState(s.Cfg.ChainID, func(a assets.Asset, addr crypto.Address) (v bignum.Int) {
						// get state from db
						tr.QueryValue(goldb.NewQuery(dbIdxAssetAddr, a, addr).Last(), &v)
						return
					})

					// execute transaction
					stateUpdates, err := tx.Execute(st)
					if err != nil {
						tr.Fail(err)
					}

					// compare result state
					if !tx.StateUpdates.Equal(stateUpdates) {
						tr.Fail(errIncorrectTxState)
					}
				}

				obj := tx.TxObject()

				switch tx.Type {

				case object.TxTypeEmission:
					if emission, ok := obj.(*object.Emission); ok {
						if emission.IsPrimaryEmission() {
							for _, out := range emission.Outs {
								// set last tx by source
								tr.Put(goldb.Key(dbIdxSourceTx, emission.Asset, out.SourceID, txUID), nil)

								// increment last tx by source
								if out.Delta > 0 {
									delta := emission.Amount(out.Delta)
									tr.IncrementBig(goldb.Key(dbIdxSourceAddr, emission.Asset, out.SourceID, out.Address), delta.BigInt())
								}
							}
						}
						// else if emission.IsReferralReward() {
						//	for _, out := range emission.Outs {
						//		if out.Delta > 0 {
						//			delta := emission.Amount(out.Delta)
						//			tr.IncrementBig(goldb.Key(dbIdxSrcInvites, emission.Asset, out.SourceID, out.Address), delta.BigInt())
						//		}
						//	}

						blockStat.IncSupplyStat(emission) // refresh totals statistic
					}

				case object.TxTypeTransfer:
					if tr, ok := obj.(*object.Transfer); ok {
						blockStat.IncVolumeStat(tr) // refresh statistic of total transfers
					}

				case object.TxTypeUser:
					userID := tx.Sender.ID()

					// get user by userID
					if usrTxUID, _ := tr.GetID(goldb.Key(dbIdxUsers, userID)); usrTxUID != 0 {
						tr.Fail(errUserHasBeenRegistered)
					}
					tr.PutID(goldb.Key(dbIdxUsers, userID), txUID)

					if usr, ok := obj.(*object.User); ok && usr.ReferrerID != 0 {
						tr.PutID(goldb.Key(dbIdxInvites, usr.ReferrerID, txUID), txUID)
					}

					blockStat.Users++ // increment users counter
				}

				// put transaction data
				tr.PutVar(goldb.Key(dbTabTxs, block.Num, txIdx), tx)

				// put index transaction by txID
				tr.PutID(goldb.Key(dbIdxTxID, txID), txUID)

				// save state to db-storage
				for stIdx, v := range tx.StateUpdates {
					if v.ChainID == s.Cfg.ChainID {
						stateTree.Put(v.StateKey(), v.Balance.Bytes())

						if v.Asset.IsName() {
							tr.PutVar(goldb.Key(dbIdxAsset, v.Asset, txUID, stIdx, v.Address), v.Balance)
						}

						tr.PutVar(goldb.Key(dbIdxAssetAddr, v.Asset, v.Address, txUID, stIdx), v.Balance)

						if !v.Balance.IsZero() {
							tr.PutVar(goldb.Key(dbIdxBalances, v.Asset, v.Address), v.Balance)
						} else {
							tr.Delete(goldb.Key(dbIdxBalances, v.Asset, v.Address))
						}
						if v.Memo != 0 { // change state with memo
							tr.PutVar(goldb.Key(dbIdxAssetAddrMemo, v.Asset, v.Address, v.Memo, txUID, stIdx), v.Balance)
						}
					}
				}
			}

			// verify state root
			if stateRoot, _ := stateTree.Root(); !bytes.Equal(block.StateRoot, stateRoot) {
				tr.Fail(errIncorrectStateRoot)
			}

			// verify chain root
			chainTree.PutVar(block.Num, block.Hash())
			if chainRoot, _ := chainTree.Root(); !bytes.Equal(block.ChainRoot, chainRoot) {
				tr.Fail(errIncorrectChainRoot)
			}

			// put block
			tr.PutVar(goldb.Key(dbTabHeaders, block.Num), block.BlockHeader)

			// save totals
			tr.PutVar(goldb.Key(dbTabStat, block.Timestamp, block.Num), blockStat)

			// middleware for each block
			for _, fn := range s.middleware {
				fn(tr, block)
			}
		}
	})

	if err != nil {
		return err
	}

	//--- success block commit ------

	// refresh last block and totals info
	s.mxR.Lock()
	s.lastBlock = blocks[len(blocks)-1]
	s.stat = blockStat
	s.mxR.Unlock()

	for _, block := range blocks {
		s.cacheHeaders.Set(block.Num, block.BlockHeader)
	}

	// remove txs from Mempool
	s.Mempool.RemoveTxs(txsIDs)

	return nil
}

func (s *BlockchainStorage) LastBlock() *blockchain.Block {
	s.mxR.RLock()
	defer s.mxR.RUnlock()
	return s.lastBlock
}

func (s *BlockchainStorage) Totals() *Statistic {
	s.mxR.RLock()
	defer s.mxR.RUnlock()
	return s.stat.Clone()
}

func (s *BlockchainStorage) CountBlocks() uint64 {
	s.mxR.RLock()
	defer s.mxR.RUnlock()
	return s.stat.Blocks
}

func (s *BlockchainStorage) CountTxs() int64 {
	s.mxR.RLock()
	defer s.mxR.RUnlock()
	return s.stat.Txs
}

func (s *BlockchainStorage) TotalsAt(t time.Time) (totals *Statistic, err error) {
	q := goldb.NewQuery(dbTabStat).Offset(t.UnixNano() / 1e3).OrderDesc().Limit(1)
	err = s.db.QueryValue(q, &totals)
	return
}

func (s *BlockchainStorage) TotalSupply(asset assets.Asset) bignum.Int {
	return s.Totals().CoinStat(asset).Supply
}

func (s *BlockchainStorage) CurrentRate(asset assets.Asset) bignum.Int {
	return s.Totals().CoinStat(asset).Rate
}

func (s *BlockchainStorage) queryLastBlock() (block *blockchain.Block, err error) {
	err = s.FetchBlocks(0, 1, true, func(b *blockchain.Block) error {
		block = b
		return nil
	})
	if err == nil && block == nil {
		block = blockchain.NewBlock(blockchain.GenesisBlockHeader(s.Cfg), nil)
	}
	return
}

func (s *BlockchainStorage) GetBlock(num uint64) (block *blockchain.Block, err error) {
	h, err := s.BlockHeader(num)
	if err != nil {
		return
	}
	txs, err := s.BlockTxs(num)
	if err != nil {
		return
	}
	return blockchain.NewBlock(h, txs), nil
}

func (s *BlockchainStorage) GetBlocks(offset uint64, limit int64, desc bool) (blocks []*blockchain.Block, err error) {
	err = s.FetchBlocks(offset, limit, desc, func(block *blockchain.Block) error {
		blocks = append(blocks, block)
		return nil
	})
	return
}

func (s *BlockchainStorage) BlockHeader(num uint64) (h *blockchain.BlockHeader, err error) {
	if num == 0 {
		return blockchain.GenesisBlockHeader(s.Cfg), nil
	}
	if h, _ = s.cacheHeaders.Get(num).(*blockchain.BlockHeader); h != nil {
		return
	}

	// get block from db-storage
	h = new(blockchain.BlockHeader)
	if ok, err := s.db.GetVar(goldb.Key(dbTabHeaders, num), h); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrBlockNotFound
	}

	s.cacheHeaders.Set(num, h)
	return h, nil
}

func (s *BlockchainStorage) FetchBlocks(offset uint64, limit int64, desc bool, fn func(block *blockchain.Block) error) error {
	return s.FetchBlockHeaders(offset, limit, desc, func(h *blockchain.BlockHeader) error {
		if txs, err := s.BlockTxs(h.Num); err != nil {
			return err
		} else {
			return fn(blockchain.NewBlock(h, txs))
		}
	})
}

func (s *BlockchainStorage) FetchBlockHeaders(offset uint64, limit int64, desc bool, fn func(block *blockchain.BlockHeader) error) error {
	q := goldb.NewQuery(dbTabHeaders)
	if offset > 0 {
		q.Offset(offset)
	}
	q.Order(desc)
	if limit > 0 {
		q.Limit(limit)
	}
	return s.db.Fetch(q, func(rec goldb.Record) error {
		var block = new(blockchain.BlockHeader)
		rec.MustDecode(block)
		return fn(block)
	})
}

//------------ txs --------------------
func encodeTxUID(blockNum uint64, txIdx int) uint64 {
	return (blockNum << 32) | uint64(txIdx)
}

func decodeTxUID(txUID uint64) (blockNum uint64, txIdx int) {
	return txUID >> 32, int(txUID & 0xffffffff)
}

func (s *BlockchainStorage) addBlockInfoToTx(tx *blockchain.Transaction, blockNum uint64, txIdx int) (err error) {
	block, err := s.BlockHeader(blockNum)
	if err == nil {
		tx.SetBlockInfo(s, blockNum, txIdx, block.Timestamp)
	}
	return
}

func (s *BlockchainStorage) GetTransaction(blockNum uint64, txIdx int) (tx *blockchain.Transaction, err error) {
	if blockNum == 0 {
		return
	}
	ok, err := s.db.GetVar(goldb.Key(dbTabTxs, blockNum, txIdx), &tx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errTxNotFound
	}
	err = s.addBlockInfoToTx(tx, blockNum, txIdx)
	return
}

func (s *BlockchainStorage) transactionByUID(txUID uint64) (*blockchain.Transaction, error) {
	if txUID == 0 {
		return nil, nil
	}
	return s.GetTransaction(decodeTxUID(txUID))
}

func (s *BlockchainStorage) BlockTxs(blockNum uint64) (txs []*blockchain.Transaction, err error) {
	if tt, ok := s.cacheTxs.Get(blockNum).([]*blockchain.Transaction); ok {
		return tt, nil
	}
	var bNum uint64
	var txIdx int
	err = s.db.Fetch(goldb.NewQuery(dbTabTxs, blockNum), func(rec goldb.Record) error {
		var tx *blockchain.Transaction
		rec.MustDecode(&tx)
		rec.MustDecodeKey(&bNum, &txIdx)
		txs = append(txs, tx)
		return s.addBlockInfoToTx(tx, bNum, txIdx)
	})
	if err == nil && len(txs) > 0 {
		s.cacheTxs.Set(blockNum, txs)
	}
	return
}

func (s *BlockchainStorage) BlockTxsCount(blockNum uint64) (count int, err error) {
	num, err := s.db.GetNumRows(goldb.NewQuery(dbTabTxs, blockNum))
	return int(num), err
}

func (s *BlockchainStorage) TransactionByHash(txHash []byte) (*blockchain.Transaction, error) {
	tx, err := s.TransactionByID(blockchain.TxIDByHash(txHash))
	if err == nil && tx != nil && !bytes.Equal(txHash, tx.Hash()) { // collision
		return nil, nil
	}
	return tx, err
}

func (s *BlockchainStorage) TransactionByID(txID uint64) (*blockchain.Transaction, error) {
	return s.transactionByIdxKey(goldb.Key(dbIdxTxID, txID))
}

func (s *BlockchainStorage) transactionByIdxKey(idxKey []byte) (tx *blockchain.Transaction, err error) {
	if tx, _ = s.cacheIdxTx.Get(idxKey).(*blockchain.Transaction); tx != nil {
		return
	}
	txUID, err := s.db.GetID(idxKey)
	if err != nil {
		return
	}
	tx, err = s.transactionByUID(txUID)
	if tx != nil {
		s.cacheIdxTx.Set(idxKey, tx)
	}
	return
}

func (s *BlockchainStorage) fetchTransactionsByIndex(q *goldb.Query, fn func(tx *blockchain.Transaction) error) error {
	return s.db.Fetch(q, func(rec goldb.Record) (err error) {
		var txUID uint64
		rec.MustDecode(&txUID)
		tx, err := s.transactionByUID(txUID)
		if tx != nil && err == nil {
			return fn(tx)
		}
		return
	})
}

func (s *BlockchainStorage) FetchTransactions(
	offset uint64,
	limit int64,
	orderDesc bool,
	fn func(tx *blockchain.Transaction) error,
) error {
	q := goldb.NewQuery(dbTabTxs)
	if offset > 0 {
		q.Offset(offset>>32, int(offset&0xffffffff)) // blockNum, txIdx
	}
	if limit > 0 {
		q.Limit(limit)
	}
	q.Order(orderDesc)
	return s.db.Fetch(q, func(rec goldb.Record) error {
		var blockNum uint64
		var txIdx int
		var tx *blockchain.Transaction
		rec.MustDecodeKey(&blockNum, &txIdx)
		rec.MustDecode(&tx)
		if err := s.addBlockInfoToTx(tx, blockNum, txIdx); err != nil {
			return err
		}
		return fn(tx)
	})
}

func (s *BlockchainStorage) FetchTransactionsByAddr(
	asset assets.Asset,
	addr crypto.Address,
	memo uint64,
	offset uint64,
	limit int64,
	orderDesc bool,
	txType int,
	fn func(tx *blockchain.Transaction, val bignum.Int) error,
) error {
	var q *goldb.Query
	if memo == 0 { // fetch transactions by address
		q = goldb.NewQuery(dbIdxAssetAddr, asset, addr)
	} else { // fetch transactions by address+memo
		q = goldb.NewQuery(dbIdxAssetAddrMemo, asset, addr, memo)
	}
	if offset > 0 {
		q.Offset(offset)
	}
	if limit <= 0 {
		limit = 1000
	}
	q.Order(orderDesc)

	var txUID uint64
	return s.db.Fetch(q, func(rec goldb.Record) error {
		if limit <= 0 {
			return goldb.Break
		}
		var _memo, _txUID uint64
		if memo == 0 {
			rec.MustDecodeKey(&asset, &addr, &_txUID)
		} else {
			rec.MustDecodeKey(&asset, &addr, &_memo, &_txUID)
		}
		if txUID == _txUID { // exclude multiple records with the same txUID
			return nil
		}
		txUID = _txUID
		tx, err := s.transactionByUID(txUID)
		if err != nil {
			return err
		}
		if txType >= 0 && int(tx.Type) != txType {
			return nil
		}
		var v bignum.Int
		rec.MustDecode(&v)
		limit--
		return fn(tx, v)
	})
}

func (s *BlockchainStorage) QueryTransaction(
	asset assets.Asset,
	addr crypto.Address,
	memo uint64,
	offset uint64,
	orderDesc bool,
) (tx *blockchain.Transaction, val bignum.Int, err error) {
	err = s.FetchTransactionsByAddr(asset, addr, memo, offset, 1, orderDesc, -1, func(t *blockchain.Transaction, v bignum.Int) error {
		tx, val = t, v
		return goldb.Break
	})
	return
}

func (s *BlockchainStorage) QueryTransactions(
	asset assets.Asset,
	addr crypto.Address,
	memo uint64,
	offset uint64,
	limitBlocks int64,
	orderDesc bool,
	txType int,
) (txs []*blockchain.Transaction, err error) {
	err = s.FetchTransactionsByAddr(asset, addr, memo, offset, limitBlocks, orderDesc, txType, func(tx *blockchain.Transaction, _ bignum.Int) error {
		txs = append(txs, tx)
		return nil
	})
	return
}

// AddressByStr returns address by nickname "@nick", "0x<hexUserID>" or by address "LikeXXXXXXXXXXXX"
func (s *BlockchainStorage) AddressByStr(str string) (addr crypto.Address, memo uint64, err error) {
	if str == "" {
		err = ErrAddrNotFound
		return
	}
	if str[0] == '@' { // address by nickname "@<nickname>"
		if tx, _, err := s.UserByNick(str); err != nil || tx == nil {
			return crypto.NilAddress, 0, err
		} else {
			return tx.SenderAddress(), 0, err
		}
	}
	if len(str) == 18 && str[:2] == "0x" { // address by userID "0x<userID:hex>"
		if userID, err := strconv.ParseUint(str[2:], 16, 64); err != nil {
			return crypto.NilAddress, 0, errIncorrectAddress
		} else if tx, _, err := s.UserByID(userID); err != nil || tx == nil {
			return crypto.NilAddress, 0, err
		} else {
			return tx.SenderAddress(), 0, nil
		}
	}
	addr, memo, err = crypto.ParseAddress(str)
	return
}

func (s *BlockchainStorage) UsernameByID(userID uint64) (nick string, err error) {
	// todo: use cache

	_, u, err := s.UserByID(userID)
	if u != nil {
		nick = u.Nick
	}
	return
}

func (s *BlockchainStorage) UserByID(userID uint64) (tx *blockchain.Transaction, u *object.User, err error) {
	if userID == 0 {
		return
	}
	if tx, err = s.transactionByIdxKey(goldb.Key(dbIdxUsers, userID)); err != nil || tx == nil {
		return
	}
	obj, err := tx.Object()
	if err != nil {
		return
	}
	u, ok := obj.(*object.User)
	if !ok || u == nil {
		err = errUserNotFound
	}
	return
}

// UserByStr returns user-info by nickname "@nick" or by address "LikeXXXXXXXXXXXX"
func (s *BlockchainStorage) UserByStr(nameOrAddr string) (tx *blockchain.Transaction, u *object.User, err error) {
	if len(nameOrAddr) == 0 {
		return
	}
	if nameOrAddr[0] == '@' {
		return s.UserByNick(nameOrAddr)
	}
	addr, _, err := crypto.ParseAddress(nameOrAddr)
	if err != nil {
		return
	}
	return s.UserByAddress(addr)
}

func (s *BlockchainStorage) UserByAddress(addr crypto.Address) (*blockchain.Transaction, *object.User, error) {
	if tx, u, err := s.UserByID(addr.ID()); err == nil && tx != nil && addr.Equal(tx.SenderAddress()) {
		return tx, u, nil
	} else {
		return nil, nil, err
	}
}

func (s *BlockchainStorage) UserByNick(name string) (tx *blockchain.Transaction, u *object.User, err error) {
	name = strings.TrimPrefix(name, "@")
	addr, _, err := s.NameAddress(name)
	if err != nil {
		return
	}
	return s.UserByID(addr.ID())
}

func (s *BlockchainStorage) FetchInvitedUsers(
	userID uint64,
	offset uint64,
	limit int64,
	orderDesc bool,
	fn func(tx *blockchain.Transaction, u *object.User) error,
) error {
	q := goldb.NewQuery(dbIdxInvites, userID)
	if offset > 0 {
		q.Offset(offset)
	}
	q.Order(orderDesc).Limit(limit)
	return s.fetchTransactionsByIndex(q, func(tx *blockchain.Transaction) error {
		if user, ok := tx.TxObject().(*object.User); ok && user != nil {
			return fn(tx, user)
		}
		return nil
	})
}

func (s *BlockchainStorage) QueryInvitedUsers(userID, offset uint64, limit int64) (users []*object.User, err error) {
	err = s.FetchInvitedUsers(userID, offset, limit, false, func(tx *blockchain.Transaction, u *object.User) error {
		users = append(users, u)
		return nil
	})
	return
}

func (s *BlockchainStorage) LastAssetTx(asset assets.Asset) (addr crypto.Address, txUID uint64, val bignum.Int, err error) {
	q := goldb.NewQuery(dbIdxAsset, asset).Last()
	err = s.db.Fetch(q, func(rec goldb.Record) error {
		rec.MustDecodeKey(&asset, &txUID, new(int), &addr)
		rec.MustDecode(&val)
		return nil
	})
	return
}

func (s *BlockchainStorage) NameAddress(name string) (addr crypto.Address, txUID uint64, err error) {
	asset := assets.NewName(name)
	addr, txUID, val, err := s.LastAssetTx(asset)
	if err != nil {
		return
	}
	if txUID == 0 {
		err = ErrAddrNotFound
	} else if val.Sign() <= 0 { // state value have to be > 0
		err = errIncorrectAssetVal
	}
	return
}

func (s *BlockchainStorage) GetBalance(addr crypto.Address, asset assets.Asset) (balance bignum.Int, lastTx *blockchain.Transaction, err error) {
	lastTx, balance, err = s.QueryTransaction(asset, addr, 0, 0, true)
	return
}

func (s *BlockchainStorage) LastTx(addr crypto.Address, memo uint64, asset assets.Asset) (lastTx *blockchain.Transaction, err error) {
	lastTx, _, err = s.QueryTransaction(asset, addr, memo, 0, true)
	return
}

func (s *BlockchainStorage) SourceTotalSupply(asset assets.Asset, addr crypto.Address, sourceID string) (total bignum.Int, err error) {
	_, err = s.db.GetVar(goldb.Key(dbIdxSourceAddr, asset, sourceID, addr), &total)
	return
}

func (s *BlockchainStorage) LastSourceData(asset assets.Asset, sourceID string) (curLikes int64, curAddr crypto.Address, err error) {
	// todo: ? use other index, remove lastSourceTx. (get curAddr, curLikes by asset and sourceID)

	_, out, err := s.lastSourceTx(asset, sourceID)
	if err == nil && out != nil {
		curLikes = out.SourceValue
		curAddr = out.Address
	}
	return
}

func (s *BlockchainStorage) lastSourceTx(asset assets.Asset, sourceID string) (tx *blockchain.Transaction, txOut *object.EmissionOut, err error) {
	err = s.db.Fetch(goldb.NewQuery(dbIdxSourceTx, asset, sourceID).Last(), func(rec goldb.Record) (err error) {
		var txUID uint64
		rec.MustDecodeKey(&asset, new(string), &txUID)
		//rec.MustDecode(&total)
		if tx, err = s.transactionByUID(txUID); tx != nil {
			txObj, _ := tx.Object()
			if emission, ok := txObj.(*object.Emission); ok && emission != nil {
				txOut = emission.OutBySrc(sourceID)
			}
		}
		return
	})
	return
}
