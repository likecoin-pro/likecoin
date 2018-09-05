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
	chainID uint64
	db      *goldb.Storage
	Mempool *mempool.Storage

	// blocks
	mxW          sync.Mutex
	mxR          sync.RWMutex
	lastBlock    *blockchain.Block //
	stat         *Statistic        //
	cacheHeaders *gosync.Cache     // blockNum => *BlockHeader
	cacheTxs     *gosync.Cache     // blockNum => []*Transaction
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
	dbIdxTxID         = 0x20 // (txID)                        => txNum
	dbIdxAsset        = 0x21 // (asset, txNum)                => sateValue
	dbIdxAssetAddr    = 0x22 // (asset, addr, txNum)          => sateValue
	dbIdxAssetAddrTag = 0x23 // (asset, addr, addrTag, txNum) => sateValue
	dbIdxUsers        = 0x24 // (userID) => txUID
	dbIdxSourceTx     = 0x25 // (providerID, sourceID, txUID) => nil
	dbIdxSourceAddr   = 0x26 // (providerID, sourceID, addr)  => total supply by addr
	//dbIdxInvites      = 0x24 // (userID, txNum)               => invitedUserID
)

var (
	ErrBlockNotFound         = errors.New("block not found")
	errTxHasBeenRegistered   = errors.New("tx has been registered")
	errTxNotFound            = errors.New("tx not found")
	errUserHasBeenRegistered = errors.New("user has been registered")
	errUserNotFound          = errors.New("user not found")
	errAddrNotFound          = errors.New("address not found")
	errIncorrectAddress      = errors.New("incorrect address")
	errIncorrectAssetVal     = errors.New("incorrect asset value")
	errIncorrectTxState      = errors.New("incorrect tx state")
	errIncorrectChainRoot    = errors.New("incorrect chain root")
	errIncorrectStateRoot    = errors.New("incorrect state root")
)

func NewBlockchainStorage(chainID uint64, dir string) (s *BlockchainStorage) {
	s = &BlockchainStorage{
		chainID:      chainID,
		db:           goldb.NewStorage(dir, nil),
		cacheHeaders: gosync.NewCache(10000),
		cacheTxs:     gosync.NewCache(1000),
		Mempool:      mempool.NewStorage(),
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
	return state.NewState(s.chainID, func(a assets.Asset, addr crypto.Address) (v bignum.Int) {
		if err := s.db.QueryValue(goldb.NewQuery(dbIdxAssetAddr, a, addr).Last(), &v); err != nil {
			panic(err)
		}
		return
	})
}

//----------------- put block --------------------------
func (s *BlockchainStorage) PutBlock(block *blockchain.Block, fVerifyTransactions bool) error {
	return s.PutBlocks([]*blockchain.Block{block}, fVerifyTransactions)
}

// open db.transaction; verify block; save block and index-records
func (s *BlockchainStorage) PutBlocks(blocks []*blockchain.Block, fVerifyTransactions bool) error {
	if len(blocks) == 0 {
		return nil
	}
	// lock tx exec
	s.mxW.Lock()
	defer s.mxW.Unlock()

	// verify blocks
	lastBlockHeader := s.lastBlock.BlockHeader
	for _, block := range blocks {
		if err := block.Verify(lastBlockHeader); err != nil {
			return err
		}
		lastBlockHeader = block.BlockHeader
	}

	var blockStat = s.stat

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

				// check transaction by txID
				if id, _ := tr.GetID(goldb.Key(dbIdxTxID, txID)); id != 0 {
					tr.Fail(errTxHasBeenRegistered)
				}

				if fVerifyTransactions {

					//-- verify sender signature
					if err := tx.Verify(); err != nil {
						tr.Fail(err)
					}

					//-- verify transaction state
					// make state by dbTransaction
					st := state.NewState(s.chainID, func(a assets.Asset, addr crypto.Address) (v bignum.Int) {
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

				// handle user registration
				switch tx.Type {

				case object.TxTypeEmission:
					obj, _ := tx.Object()
					if emission, ok := obj.(*object.Emission); ok {
						for _, out := range emission.Outs {
							// set last tx by source
							tr.Put(goldb.Key(dbIdxSourceTx, emission.Asset, out.SourceID, txUID), nil)

							// increment last tx by source
							var srcAddrTotal bignum.Int
							tr.GetVar(goldb.Key(dbIdxSourceAddr, emission.Asset, out.SourceID, out.Address), &srcAddrTotal)
							srcAddrTotal = srcAddrTotal.Add(emission.Amount(out.Delta))
							tr.PutVar(goldb.Key(dbIdxSourceAddr, emission.Asset, out.SourceID, out.Address), srcAddrTotal)
						}

						blockStat.IncSupplyStat(emission) // refresh totals statistic
					}

				case object.TxTypeTransfer:
					obj, _ := tx.Object()
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

					blockStat.Users++ // increment users counter
				}

				// put transaction data
				tr.PutVar(goldb.Key(dbTabTxs, block.Num, txIdx), tx)

				// put index transaction by txID
				tr.PutID(goldb.Key(dbIdxTxID, txID), txUID)

				// save state to db-storage
				for stIdx, v := range tx.StateUpdates {
					if v.ChainID == s.chainID {
						stateTree.Put(v.StateKey(), v.Balance.Bytes())

						if v.Asset.IsName() {
							tr.PutVar(goldb.Key(dbIdxAsset, v.Asset, txUID, stIdx, v.Address), v.Balance)
						}

						tr.PutVar(goldb.Key(dbIdxAssetAddr, v.Asset, v.Address, txUID, stIdx), v.Balance)

						if v.Tag != 0 { // change state with tag
							tr.PutVar(goldb.Key(dbIdxAssetAddrTag, v.Asset, v.Address, v.Tag, txUID, stIdx), v.Balance)
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
		block = blockchain.NewBlock(blockchain.GenesisBlockHeader(), nil)
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

func (s *BlockchainStorage) BlockHeader(num uint64) (h *blockchain.BlockHeader, err error) {
	if num == 0 {
		return blockchain.GenesisBlockHeader(), nil
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
		tx.SetBlockInfo(blockNum, txIdx, block.Timestamp)
	}
	return
}

func (s *BlockchainStorage) GetTransaction(blockNum uint64, txIdx int) (tx *blockchain.Transaction, err error) {
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
	blockNum, txIdx := decodeTxUID(txUID)
	return s.GetTransaction(blockNum, txIdx)
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

func (s *BlockchainStorage) transactionByIdxKey(idxKey []byte) (*blockchain.Transaction, error) {
	if txUID, err := s.db.GetID(idxKey); err != nil {
		return nil, err
	} else {
		return s.transactionByUID(txUID)
	}
}

func (s *BlockchainStorage) FetchTransactions(
	asset assets.Asset,
	addr crypto.Address,
	tag uint64,
	offset uint64,
	limit int64,
	orderDesc bool,
	fn func(tx *blockchain.Transaction, val bignum.Int) error,
) error {
	var q *goldb.Query
	if tag == 0 { // fetch transactions by address
		q = goldb.NewQuery(dbIdxAssetAddr, asset, addr)
	} else { // fetch transactions by address+tag
		q = goldb.NewQuery(dbIdxAssetAddrTag, asset, addr, tag)
	}
	if offset > 0 {
		q.Offset(offset)
	}
	if limit <= 0 {
		limit = 1000
	}
	q.Limit(limit)
	q.Order(orderDesc)

	var txUID uint64
	return s.db.Fetch(q, func(rec goldb.Record) error {
		var _tag, _txUID uint64
		if tag == 0 {
			rec.MustDecodeKey(&asset, &addr, &_txUID)
		} else {
			rec.MustDecodeKey(&asset, &addr, &_tag, &_txUID)
		}
		if txUID == _txUID { // exclude multiple records with the same txUID
			return nil
		}
		txUID = _txUID
		tx, err := s.transactionByUID(txUID)
		if err != nil {
			return err
		}
		var v bignum.Int
		rec.MustDecode(&v)
		return fn(tx, v)
	})
}

func (s *BlockchainStorage) QueryTransaction(
	asset assets.Asset,
	addr crypto.Address,
	tag uint64,
	offset uint64,
	orderDesc bool,
) (tx *blockchain.Transaction, val bignum.Int, err error) {
	err = s.FetchTransactions(asset, addr, tag, offset, 1, orderDesc, func(t *blockchain.Transaction, v bignum.Int) error {
		tx, val = t, v
		return goldb.Break
	})
	return
}

func (s *BlockchainStorage) QueryTransactions(
	asset assets.Asset,
	addr crypto.Address,
	tag uint64,
	offset uint64,
	limitBlocks int64,
	orderDesc bool,
) (txs []*blockchain.Transaction, err error) {
	err = s.FetchTransactions(asset, addr, tag, offset, limitBlocks, orderDesc, func(tx *blockchain.Transaction, _ bignum.Int) error {
		txs = append(txs, tx)
		return nil
	})
	return
}

// AddressByStr returns address by nickname "@nick", "0x<hexUserID>" or by address "LikeXXXXXXXXXXXX"
func (s *BlockchainStorage) AddressByStr(str string) (addr crypto.Address, tag uint64, err error) {
	if str == "" {
		err = errAddrNotFound
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
		} else if tx, _, err := s.UserByID(userID); err != nil {
			return crypto.NilAddress, 0, err
		} else {
			return tx.SenderAddress(), 0, nil
		}
	}
	addr, tag, err = crypto.ParseAddress(str)
	return
}

func (s *BlockchainStorage) UserByID(userID uint64) (tx *blockchain.Transaction, u *object.User, err error) {
	if tx, err = s.transactionByIdxKey(goldb.Key(dbIdxUsers, userID)); err != nil {
		return
	}
	obj, err := tx.Object()
	if err != nil {
		return
	}
	u, ok := obj.(*object.User)
	if !ok {
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
	return s.UserByID(addr.ID())
}

func (s *BlockchainStorage) UserByNick(name string) (tx *blockchain.Transaction, u *object.User, err error) {
	name = strings.TrimPrefix(name, "@")
	addr, _, err := s.NameAddress(name)
	if err != nil {
		return
	}
	return s.UserByID(addr.ID())
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
		err = errAddrNotFound
	} else if val.Sign() <= 0 { // state value have to be > 0
		err = errIncorrectAssetVal
	}
	return
}

func (s *BlockchainStorage) GetBalance(addr crypto.Address, asset assets.Asset) (balance bignum.Int, lastTx *blockchain.Transaction, err error) {
	lastTx, balance, err = s.QueryTransaction(asset, addr, 0, 0, true)
	return
}

func (s *BlockchainStorage) LastTx(addr crypto.Address, tag uint64, asset assets.Asset) (lastTx *blockchain.Transaction, err error) {
	lastTx, _, err = s.QueryTransaction(asset, addr, tag, 0, true)
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
