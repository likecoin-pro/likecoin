package db

import (
	"bytes"
	"errors"
	"math/big"
	"strings"

	"github.com/denisskin/goldb"
	"github.com/denisskin/gosync"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
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
	lastBlock   *blockchain.Block //
	cacheBlocks *gosync.Cache     // blockNum => *Block
}

const (
	// tables
	dbTabBlock     = 0x01 // (blockNum) => Block
	dbTabTxs       = 0x02 // (blockNum, txIdx) => BlockTx
	dbTabChainTree = 0x03 //
	dbTabStateTree = 0x04 // (asset, addr) => sateValue

	// indexes
	dbIdxTxID         = 0x10 // (txID)                        => txNum
	dbIdxAsset        = 0x11 // (asset, txNum)                => sateValue
	dbIdxAssetAddr    = 0x12 // (asset, addr, txNum)          => sateValue
	dbIdxAssetAddrTag = 0x13 // (asset, addr, addrTag, txNum) => sateValue
	dbIdxUsers        = 0x14 // (userID) => txUID
	//dbIdxInvites      = 0x14 // (userID, txNum)               => invitedUserID
)

var (
	errBlockNotFound         = errors.New("block not found")
	errTxHasBeenRegistered   = errors.New("tx has been registered")
	errTxNotFound            = errors.New("tx not found")
	errUserHasBeenRegistered = errors.New("user has been registered")
	errUserNotFound          = errors.New("user not found")
	errAddrNotFound          = errors.New("address not found")
	errIncorrectTxState      = errors.New("incorrect tx state")
	errIncorrectChainRoot    = errors.New("incorrect chain root")
	errIncorrectStateRoot    = errors.New("incorrect state root")
)

func NewBlockchainStorage(chainID uint64, dir string) (s *BlockchainStorage) {
	s = &BlockchainStorage{
		chainID:     chainID,
		db:          goldb.NewStorage(dir, nil),
		cacheBlocks: gosync.NewCache(10000),
		Mempool:     mempool.NewStorage(),
	}

	b, err := s.getLastBlock()
	if err != nil {
		panic(err)
	}
	s.lastBlock = b

	return
}

func (s *BlockchainStorage) Close() (err error) {
	return s.db.Close()
}

func (s *BlockchainStorage) Drop() (err error) {
	s.db.Close()
	return s.db.Drop()
}

func newPatriciaTree(db patricia.Storage, tab goldb.Entity) *patricia.Tree {
	_, isTx := db.(*goldb.Transaction)
	return patricia.NewTree(db, goldb.Key(tab), !isTx)
}

func (s *BlockchainStorage) ChainTree() *patricia.Tree {
	return newPatriciaTree(s.db, dbTabChainTree)
}

func (s *BlockchainStorage) StateTree() *patricia.Tree {
	return newPatriciaTree(s.db, dbTabStateTree)
}

// State returns state struct from db
func (s *BlockchainStorage) State() *state.State {
	return state.NewState(s.chainID, func(a assets.Asset, addr crypto.Address) state.Number {
		var v *big.Int
		if err := s.db.QueryValue(goldb.NewQuery(dbIdxAssetAddr, a, addr).Last(), &v); err != nil {
			panic(err)
		}
		return v
	})
}

//----------------- put block --------------------------
// open db.transaction; verify block; save block and index-records
func (s *BlockchainStorage) PutBlock(
	block *blockchain.Block,
	txs []*blockchain.BlockTx,
	fVerifyTransactions bool,
) (err error) {
	return s.db.Exec(func(tr *goldb.Transaction) {

		// verify block header
		if err := block.Verify(s.lastBlock); err != nil {
			tr.Fail(err)
		}
		// verify block txs
		if err := block.VerifyTxs(txs); err != nil {
			tr.Fail(err)
		}

		stateTree := newPatriciaTree(tr, dbTabStateTree)

		// add index on transactions
		for txIdx, bTx := range txs {

			tx := bTx.Tx
			txID := tx.ID()
			txUID := encodeTxUID(block.Num, txIdx)

			// check transaction by txID
			if id, _ := tr.GetID(goldb.Key(dbIdxTxID, txID)); id != 0 {
				tr.Fail(errTxHasBeenRegistered)
			}

			if fVerifyTransactions {

				//-- verify sender signature
				if err := bTx.Verify(); err != nil {
					tr.Fail(err)
				}

				//-- verify transaction state
				// make state by dbTransaction
				st := state.NewState(s.chainID, func(a assets.Asset, addr crypto.Address) state.Number { // get state from db
					var v *big.Int
					tr.QueryValue(goldb.NewQuery(dbIdxAssetAddr, a, addr).Last(), &v)
					return v
				})

				// execute transaction
				stateUpdates, err := tx.Execute(st)
				if err != nil {
					tr.Fail(err)
				}

				// compare state
				if !bTx.StateUpdates.Equal(stateUpdates) {
					tr.Fail(errIncorrectTxState)
				}
			}

			// handle user registration
			if tx.Type == object.TxTypeUser {
				userID := tx.Sender.ID()

				// get user by userID
				if usrTxUID, _ := tr.GetID(goldb.Key(dbIdxUsers, userID)); usrTxUID != 0 {
					tr.Fail(errUserHasBeenRegistered)
				}
				tr.PutID(goldb.Key(dbIdxUsers, userID), txUID) //
			}

			// put transaction data
			tr.PutVar(goldb.Key(dbTabTxs, block.Num, txIdx), bTx)

			// put index transaction by txID
			tr.PutID(goldb.Key(dbIdxTxID, txID), txUID)

			// save state to db-storage
			for stIdx, v := range bTx.StateUpdates {
				if v.ChainID == s.chainID {
					stateTree.Put(v.StateKey(), v.Balance.Bytes())

					tr.PutVar(goldb.Key(dbIdxAsset, v.Asset, txUID, stIdx, v.Address), v.Balance)

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
		chainTree := newPatriciaTree(tr, dbTabChainTree)
		chainTree.PutVar(block.Num, block.Hash())
		if chainRoot, _ := chainTree.Root(); !bytes.Equal(block.ChainRoot, chainRoot) {
			tr.Fail(errIncorrectChainRoot)
		}

		// put block
		tr.PutVar(goldb.Key(dbTabBlock, block.Num), block)

		// success; commit transaction; add block to blockchain
		s.lastBlock = block
		s.cacheBlocks.Set(block.Num, block)
	})
}

func (s *BlockchainStorage) LastBlock() *blockchain.Block {
	return s.lastBlock
}

func (s *BlockchainStorage) LastBlockNum() uint64 {
	return s.lastBlock.Num
}

func (s *BlockchainStorage) getLastBlock() (block *blockchain.Block, err error) {
	err = s.FetchBlocks(0, 1, true, func(b *blockchain.Block) error {
		block = b
		return nil
	})
	if err == nil && block == nil {
		block = blockchain.GenesisBlock()
	}
	return
}

func (s *BlockchainStorage) GetBlock(num uint64) (block *blockchain.Block, err error) {
	if num == 0 {
		return blockchain.GenesisBlock(), nil
	}
	if block, _ = s.cacheBlocks.Get(num).(*blockchain.Block); block != nil {
		return
	}

	// get block from db-storage
	block = new(blockchain.Block)
	if ok, err := s.db.GetVar(goldb.Key(dbTabBlock, num), block); err != nil {
		return nil, err
	} else if !ok {
		return nil, errBlockNotFound
	}

	s.cacheBlocks.Set(num, block)
	return block, nil
}

func (s *BlockchainStorage) BlockSize(num uint64) (sz int64) {
	if block, err := s.GetBlock(num); err == nil {
		sz += block.Size()
	}
	txs, _ := s.BlockTxs(num)
	for _, tx := range txs {
		sz += int64(len(tx.Encode()))
	}
	return
}

func (s *BlockchainStorage) FetchBlocks(offset uint64, limit int64, desc bool, fn func(block *blockchain.Block) error) error {
	q := goldb.NewQuery(dbTabBlock)
	if offset > 0 {
		q.Offset(offset)
	}
	q.Order(desc)
	if limit > 0 {
		q.Limit(limit)
	}
	return s.db.Fetch(q, func(rec goldb.Record) error {
		var block = new(blockchain.Block)
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

func (s *BlockchainStorage) addBlockInfoToTx(bTx *blockchain.BlockTx, blockNum uint64, txIdx int) (err error) {
	block, err := s.GetBlock(blockNum)
	if err == nil {
		bTx.TxSeq = encodeTxUID(blockNum, txIdx)
		bTx.BlockTs = block.Timestamp
	}
	return
}

func (s *BlockchainStorage) GetTransaction(blockNum uint64, txIdx int) (bTx *blockchain.BlockTx, err error) {
	ok, err := s.db.GetVar(goldb.Key(dbTabTxs, blockNum, txIdx), &bTx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errTxNotFound
	}
	err = s.addBlockInfoToTx(bTx, blockNum, txIdx)
	return
}

func (s *BlockchainStorage) transactionByUID(txUID uint64) (*blockchain.BlockTx, error) {
	blockNum, txIdx := decodeTxUID(txUID)
	return s.GetTransaction(blockNum, txIdx)
}

func (s *BlockchainStorage) BlockTxs(blockNum uint64) (txs []*blockchain.BlockTx, err error) {
	var bNum uint64
	var txIdx int
	err = s.db.Fetch(goldb.NewQuery(dbTabTxs, blockNum), func(rec goldb.Record) error {
		var tx *blockchain.BlockTx
		rec.MustDecode(&tx)
		rec.MustDecodeKey(&bNum, &txIdx)
		txs = append(txs, tx)
		return s.addBlockInfoToTx(tx, bNum, txIdx)
	})
	return
}

func (s *BlockchainStorage) TransactionByHash(txHash []byte) (*blockchain.BlockTx, error) {
	it, err := s.TransactionByID(blockchain.TxIDByHash(txHash))
	if err == nil && it != nil && !bytes.Equal(txHash, it.Tx.Hash()) { // collision
		return nil, nil
	}
	return it, err
}

func (s *BlockchainStorage) TransactionByID(txID uint64) (*blockchain.BlockTx, error) {
	return s.transactionByIdxKey(goldb.Key(dbIdxTxID, txID))
}

func (s *BlockchainStorage) transactionByIdxKey(idxKey []byte) (*blockchain.BlockTx, error) {
	if txUID, err := s.db.GetID(idxKey); err != nil {
		return nil, err
	} else {
		return s.transactionByUID(txUID)
	}
}

func (s *BlockchainStorage) FetchTransactions(
	asset assets.Asset,
	addr crypto.Address,
	filterTag uint64,
	offsetBlock uint64,
	limitBlocks int64,
	orderDesc bool,
	fn func(tx *blockchain.BlockTx, val state.Number) error,
) error {

	typ, q := dbIdxAsset, goldb.NewQuery(dbIdxAsset, asset)
	if filterTag != 0 { // fetch transactions by address tag
		typ, q = dbIdxAssetAddrTag, goldb.NewQuery(dbIdxAssetAddrTag, asset, addr, filterTag)
	} else if !addr.IsNil() { // get address history
		typ, q = dbIdxAssetAddr, goldb.NewQuery(dbIdxAssetAddr, asset, addr)
	}
	if offsetBlock > 0 {
		q.Offset(offsetBlock << 32)
	}
	q.Order(orderDesc)
	var txUID, lastBlock uint64
	var nBlocks int64
	return s.db.Fetch(q, func(rec goldb.Record) error {
		switch typ { // get txUID from record-key
		case dbIdxAsset:
			rec.MustDecodeKey(&asset, &txUID)
		case dbIdxAssetAddr:
			rec.MustDecodeKey(&asset, &addr, &txUID)
		case dbIdxAssetAddrTag:
			rec.MustDecodeKey(&asset, &addr, &filterTag, &txUID)
		}
		if blockNum := txUID >> 32; blockNum != lastBlock {
			lastBlock = blockNum
			nBlocks++
			if limitBlocks > 0 && nBlocks > limitBlocks {
				return goldb.Break
			}
		}
		tx, err := s.transactionByUID(txUID)
		if err != nil {
			return err
		}
		return fn(tx, rec.ValueBigInt())
	})
}

func (s *BlockchainStorage) QueryTransaction(
	asset assets.Asset,
	addr crypto.Address,
	filterTag uint64,
	offsetBlock uint64,
	orderDesc bool,
) (tx *blockchain.BlockTx, val state.Number, err error) {
	val = state.Int(0)
	err = s.FetchTransactions(asset, addr, filterTag, offsetBlock, 1, orderDesc, func(t *blockchain.BlockTx, v state.Number) error {
		tx, val = t, v
		return goldb.Break
	})
	return
}

func (s *BlockchainStorage) QueryTransactions(
	asset assets.Asset,
	addr crypto.Address,
	filterTag uint64,
	offsetBlock uint64,
	limitBlocks int64,
	orderDesc bool,
) (txs []*blockchain.BlockTx, err error) {
	err = s.FetchTransactions(asset, addr, filterTag, offsetBlock, limitBlocks, orderDesc, func(tx *blockchain.BlockTx, _ state.Number) error {
		txs = append(txs, tx)
		return nil
	})
	return
}

// AddrByStr returns address by nickname "@nick" or by address "LikeXXXXXXXXXXXX"
func (s *BlockchainStorage) AddrByStr(nameOrAddr string) (addr crypto.Address, err error) {
	if nameOrAddr == "" {
		err = errAddrNotFound
		return
	}
	if nameOrAddr[0] == '@' {
		if tx, _, err := s.UserByNick(nameOrAddr); err != nil || tx == nil {
			return crypto.NilAddress, err
		} else {
			return tx.TxAddress(), err
		}
	}
	addr, _, err = crypto.ParseAddress(nameOrAddr)
	return
}

func (s *BlockchainStorage) UserByID(userID uint64) (bTx *blockchain.BlockTx, u *object.User, err error) {
	if bTx, err = s.transactionByIdxKey(goldb.Key(dbIdxUsers, userID)); err != nil {
		return
	}
	obj, err := bTx.Tx.Object()
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
func (s *BlockchainStorage) UserByStr(nameOrAddr string) (tx *blockchain.BlockTx, u *object.User, err error) {

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

func (s *BlockchainStorage) UserByNick(name string) (tx *blockchain.BlockTx, u *object.User, err error) {
	name = strings.TrimPrefix(name, "@")
	addr, _, err := s.NameAddress(name)
	if err != nil {
		return
	}
	return s.UserByID(addr.ID())
}

func (s *BlockchainStorage) LastAssetAddress(asset assets.Asset) (addr crypto.Address, txUID uint64, val state.Number, err error) {
	q := goldb.NewQuery(dbIdxAsset, asset).Last()
	err = s.db.Fetch(q, func(rec goldb.Record) error {
		rec.MustDecodeKey(&asset, &txUID, new(int), &addr)
		val = rec.ValueBigInt()
		return nil
	})
	return
}

func (s *BlockchainStorage) NameAddress(name string) (addr crypto.Address, txUID uint64, err error) {
	asset := assets.NewName(name)
	addr, txUID, val, err := s.LastAssetAddress(asset)
	if err != nil {
		return
	}
	if txUID == 0 {
		err = errAddrNotFound
	} else if val.Sign() <= 0 { // state value have to be > 0
		err = errors.New("incorrect asset value")
	}
	return
}

func (s *BlockchainStorage) GetBalance(addr crypto.Address, asset assets.Asset) (balance state.Number, lastTx *blockchain.BlockTx, err error) {
	lastTx, balance, err = s.QueryTransaction(asset, addr, 0, 0, true)
	return
}

func (s *BlockchainStorage) LastTx(addr crypto.Address, tag uint64, asset assets.Asset) (lastTx *blockchain.BlockTx, err error) {
	lastTx, _, err = s.QueryTransaction(asset, addr, tag, 0, true)
	return
}
