package db

import (
	"bytes"
	"errors"
	"math/big"
	"sync"

	"github.com/denisskin/goldb"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/blockchain/transaction"
	"github.com/likecoin-pro/likecoin/commons/hex"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/likecoin-pro/likecoin/object"
)

type BlockchainStorage struct {
	chainID   uint64
	db        *goldb.Storage
	lastBlock *blockchain.Block

	cacheBlock struct {
		block *blockchain.Block
		mx    sync.RWMutex
	}
}

const (
	// tables
	dbTabBlock     = 0x01 // (blockNum) => blockData
	dbTabUsers     = 0x02 // (userID) => txNum
	dbTabStateTree = 0x03 // (asset, addr) => sateValue
	dbTabChainTree = 0x04 // (asset, addr) => sateValue

	// indexes
	dbIdxTxID         = 0x10 // (txID)                        => txNum
	dbIdxAsset        = 0x11 // (asset, txNum)                => sateValue
	dbIdxAssetAddr    = 0x12 // (asset, addr, txNum)          => sateValue
	dbIdxAssetAddrTag = 0x13 // (asset, addr, addrTag, txNum) => sateValue
	dbIdxInvites      = 0x14 // (userID, txNum)               => invitedUserID
)

var (
	errBlockNotFound         = errors.New("block not found")
	errTxHasBeenRegistered   = errors.New("tx has been registered")
	errTxNotFound            = errors.New("tx not found")
	errUserHasBeenRegistered = errors.New("user has been registered")
	errUserNotFound          = errors.New("user not found")
	errIncorrectTxParams     = errors.New("incorrect tx params")
	errIncorrectTxState      = errors.New("incorrect tx state")
	errIncorrectChainRoot    = errors.New("incorrect chain root")
	errIncorrectStateRoot    = errors.New("incorrect state root")
)

func NewBlockchainStorage(chainID uint64, dir string) (s *BlockchainStorage) {
	s = &BlockchainStorage{
		chainID: chainID,
		db:      goldb.NewStorage(dir, nil),
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

// open db.transaction; verify block; save block and index-records
func (s *BlockchainStorage) PutBlock(block *blockchain.Block, fVerifyTransactions bool) (err error) {
	return s.db.Exec(func(tr *goldb.Transaction) {

		// verify block headers
		if err := block.VerifyHeader(s.lastBlock); err != nil {
			tr.Fail(err) // dbTransaction fail
		}

		stateTree := newPatriciaTree(tr, dbTabStateTree)

		// add index on transactions
		for _, it := range block.Items {

			tx := it.Tx
			txID := it.TxID()
			txUID := it.UID()

			// check tx-chain info
			if h := tx.GetHeader(); h.ChainID != s.chainID {
				tr.Fail(errIncorrectTxParams)
			}

			// check transaction by txID
			if id, _ := tr.GetInt(goldb.Key(dbIdxTxID, txID)); id != 0 {
				tr.Fail(errTxHasBeenRegistered)
			}

			if fVerifyTransactions {
				//-- verify transaction data. (todo: ?? parallel verify)
				if err := tx.Verify(); err != nil {
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
				newState, err := st.Execute(tx)
				if err != nil {
					tr.Fail(err)
				}

				// verify state
				if !it.State.Equal(newState.Values()) {
					tr.Fail(errIncorrectTxState)
				}
			}

			// handle user registration
			if user, ok := tx.(*object.User); ok {
				userID := user.UserID()

				// get user by userID
				if usrTxUID, _ := tr.GetID(goldb.Key(dbTabUsers, userID)); usrTxUID != 0 {
					tr.Fail(errUserHasBeenRegistered)
				}
				tr.PutID(goldb.Key(dbTabUsers, userID), txUID) //

				// todo: referrerID
			}

			// put index transaction by txID
			tr.PutVar(goldb.Key(dbIdxTxID, txID), txUID)

			// save state to db-storage
			for seq, v := range it.State {

				stateTree.Put(v.Hash())

				if v.ChainID == s.chainID {
					tr.PutVar(goldb.Key(dbIdxAsset, v.Asset, txUID, seq, v.Address), v.Balance)

					tr.PutVar(goldb.Key(dbIdxAssetAddr, v.Asset, v.Address, txUID, seq), v.Balance)

					if v.Tag != 0 { // change state with tag
						tr.PutVar(goldb.Key(dbIdxAssetAddrTag, v.Asset, v.Address, v.Tag, txUID, seq), v.Balance)
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
		chainTree.Put(block.Hash())
		if chainRoot, _ := chainTree.Root(); !bytes.Equal(block.ChainRoot, chainRoot) {
			tr.Fail(errIncorrectChainRoot)
		}

		// put block
		tr.PutVar(goldb.Key(dbTabBlock, block.Num), block)

		s.lastBlock = block // success; commit transaction; add block to blockchain
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
	// get from block cache
	s.cacheBlock.mx.RLock()
	if s.cacheBlock.block != nil && s.cacheBlock.block.Num == num {
		block = s.cacheBlock.block
	}
	s.cacheBlock.mx.RUnlock()
	if block != nil {
		return
	}

	// get block from db-storage
	block = new(blockchain.Block)
	if ok, err := s.db.GetVar(goldb.Key(dbTabBlock, num), block); err != nil {
		return nil, err
	} else if !ok {
		return nil, errBlockNotFound
	}

	// set block-cache
	s.cacheBlock.mx.Lock()
	s.cacheBlock.block = block
	s.cacheBlock.mx.Unlock()

	return block, nil
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
		rec.Decode(block)
		return fn(block)
	})
}

func (s *BlockchainStorage) TransactionByHash(txHash []byte) (*blockchain.BlockItem, error) {
	it, err := s.TransactionByID(transaction.TxIDByHash(txHash))
	if err == nil && it != nil && !bytes.Equal(txHash, it.TxHash()) { // collision
		return nil, nil
	}
	return it, err
}

func (s *BlockchainStorage) TransactionByID(txID uint64) (*blockchain.BlockItem, error) {
	if txUID, err := s.db.GetID(goldb.Key(dbIdxTxID, txID)); err != nil {
		return nil, err
	} else {
		return s.TransactionByUID(txUID)
	}
}

func (s *BlockchainStorage) TransactionByUID(txUID uint64) (it *blockchain.BlockItem, err error) {
	blockNum, txIdx := blockchain.DecodeTxUID(txUID)
	block, err := s.GetBlock(blockNum)
	if err != nil {
		return
	}
	if txIdx >= len(block.Items) || txIdx < 0 {
		err = errTxNotFound
		return
	}
	return block.Items[txIdx], nil
}

func (s *BlockchainStorage) fetchTxUID(
	asset assets.Asset,
	addr crypto.Address,
	filterTag uint64,
	offsetBlock uint64,
	limitBlocks int64,
	orderDesc bool,
	fn func(txUID uint64, val state.Number) error,
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
			rec.DecodeKey(&asset, &txUID)
		case dbIdxAssetAddr:
			rec.DecodeKey(&asset, &addr, &txUID)
		case dbIdxAssetAddrTag:
			rec.DecodeKey(&asset, &addr, &filterTag, &txUID)
		}
		if blockNum := txUID >> 32; blockNum != lastBlock {
			lastBlock = blockNum
			nBlocks++
			if limitBlocks > 0 && nBlocks > limitBlocks {
				return goldb.Break
			}
		}
		return fn(txUID, rec.ValueBigInt())
	})
}

func (s *BlockchainStorage) FetchTransactions(
	asset assets.Asset,
	addr crypto.Address,
	filterTag uint64,
	offsetBlock uint64,
	limitBlocks int64,
	orderDesc bool,
	fn func(tx *blockchain.BlockItem, val state.Number) error,
) error {
	return s.fetchTxUID(asset, addr, filterTag, offsetBlock, limitBlocks, orderDesc, func(txUID uint64, val state.Number) error {
		if tx, err := s.TransactionByUID(txUID); err != nil {
			return err
		} else {
			return fn(tx, val)
		}
	})
}

func (s *BlockchainStorage) QueryTransaction(
	asset assets.Asset,
	addr crypto.Address,
	filterTag uint64,
	offsetBlock uint64,
	orderDesc bool,
) (tx *blockchain.BlockItem, val state.Number, err error) {
	val = state.Int(0)
	err = s.fetchTxUID(asset, addr, filterTag, offsetBlock, 1, orderDesc, func(txUID uint64, v state.Number) error {
		val = v
		tx, err = s.TransactionByUID(txUID)
		if err != nil {
			return err
		}
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
) (txs []*blockchain.BlockItem, err error) {
	err = s.FetchTransactions(asset, addr, filterTag, offsetBlock, limitBlocks, orderDesc, func(tx *blockchain.BlockItem, _ state.Number) error {
		txs = append(txs, tx)
		return nil
	})
	return
}

func (s *BlockchainStorage) UserByID(userID uint64) (tx *blockchain.BlockItem, u *object.User, err error) {
	txUID, err := s.db.GetID(goldb.Key(dbTabUsers, userID))
	if err != nil {
		return
	}
	if tx, err = s.TransactionByUID(txUID); err != nil {
		return
	}
	u, ok := tx.Tx.(*object.User)
	if !ok {
		err = errUserNotFound
	}
	return
}

func (s *BlockchainStorage) UserByNick(name string) (tx *blockchain.BlockItem, u *object.User, err error) {
	addr, _, err := s.NameAddress(name)
	if err != nil {
		return
	}
	return s.UserByID(addr.ID())
}

func (s *BlockchainStorage) LastAssetAddress(asset assets.Asset) (addr crypto.Address, txUID uint64, val state.Number, err error) {
	q := goldb.NewQuery(dbIdxAsset, asset).Last()
	err = s.db.Fetch(q, func(rec goldb.Record) error {
		rec.DecodeKey(&asset, &txUID, new(int), &addr)
		val = rec.ValueBigInt()
		return nil
	})
	return
}

func (s *BlockchainStorage) NameAddress(name string) (addr crypto.Address, txUID uint64, err error) {
	asset := assets.NewName(name)
	addr, txUID, val, err := s.LastAssetAddress(asset)
	if val.Sign() <= 0 { // state value have to be > 0
		err = errors.New("incorrect asset value")
	}
	return
}

func (s *BlockchainStorage) GetBalance(addr crypto.Address, asset assets.Asset) (balance state.Number, lastTx *blockchain.BlockItem, err error) {
	lastTx, balance, err = s.QueryTransaction(asset, addr, 0, 0, true)
	return
}

func (s *BlockchainStorage) LastTx(addr crypto.Address, tag uint64, asset assets.Asset) (lastTx *blockchain.BlockItem, err error) {
	lastTx, _, err = s.QueryTransaction(asset, addr, tag, 0, true)
	return
}

type AddressInfo struct {
	Address       string       `json:"address"`     // original address
	Tag           uint64       `json:"tag"`         //
	TaggedAddress string       `json:"address+tag"` // address+tag
	Balance       state.Number `json:"balance"`     // balance on address (not tagged address)
	Asset         assets.Asset `json:"asset"`       //
	LastTx        hex.Bytes    `json:"last_tx"`     // last tx of tagged_address
}

func (s *BlockchainStorage) AddressInfo(addr crypto.Address, tag uint64, asset assets.Asset) (inf AddressInfo, err error) {
	inf.TaggedAddress = addr.TaggedString(tag)
	inf.Address = addr.String()
	inf.Tag = tag
	inf.Asset = asset
	bal, tx, err := s.GetBalance(addr, asset)
	if err != nil {
		return
	}
	inf.Balance = bal
	if tag != 0 {
		if tx, err = s.LastTx(addr, tag, asset); err != nil {
			return
		}
	}
	if tx != nil {
		inf.LastTx = tx.TxHash()
	}
	return
}
