package db

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/denisskin/goldb"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/likecoin-pro/likecoin/object"
)

type BlockchainStorage struct {
	storage   *goldb.Storage
	lastBlock *blockchain.Block
}

const (
	// tables
	dbTabBlock = 0x01 // (blockNum) => blockData
	dbTabUsers = 0x02 // (userID) => txNum
	//dbTabState = 0x03 // (asset, addr) => sateValue

	// indexes
	dbIdxTxID         = 0x10 // (txID)                        => txNum
	dbIdxAsset        = 0x11 // (asset, txNum)                => sateValue
	dbIdxAssetAddr    = 0x12 // (asset, addr, txNum)          => sateValue
	dbIdxAssetAddrTag = 0x13 // (asset, addr, addrTag, txNum) => sateValue
)

var (
	errBlockNotFound         = errors.New("block not found")
	errTxHasBeenRegistered   = errors.New("tx has been registered")
	errTxNotFound            = errors.New("tx not found")
	errUserHasBeenRegistered = errors.New("user has been registered")
	errUserNotFound          = errors.New("user not found")
	errIncorrectTxState      = errors.New("incorrect tx state")
)

func NewBlockchainStorage(dir string) (s *BlockchainStorage) {
	s = &BlockchainStorage{
		storage: goldb.NewStorage(dir, nil),
	}

	b, err := s.getLastBlock()
	if err != nil {
		panic(err)
	}
	s.lastBlock = b

	return
}

func (s *BlockchainStorage) Close() (err error) {
	return s.storage.Close()
}

func (s *BlockchainStorage) Drop() (err error) {
	s.storage.Close()
	return s.storage.Drop()
}

// open db.transaction; verify block; save block and index-records
func (s *BlockchainStorage) PutBlock(block *blockchain.Block) (err error) {
	return s.storage.Exec(func(tr *goldb.Transaction) {

		// verify block headers
		if err := block.VerifyHeader(s.lastBlock); err != nil {
			tr.Fail(err) // dbTransaction fail
		}

		// put block
		tr.PutVar(goldb.Key(dbTabBlock, block.Num), block)

		// add index on transactions
		for _, it := range block.Items {

			tx := it.Tx
			txID := it.TxID()
			txUID := it.UID()

			// check transaction by txID
			if id, _ := tr.GetInt(goldb.Key(dbIdxTxID, txID)); id != 0 {
				tr.Fail(errTxHasBeenRegistered)
			}

			// handle user registration
			if user, ok := tx.(*object.User); ok {
				userID := user.ID()

				// get user by userID
				if usrTxUID, _ := tr.GetID(goldb.Key(dbTabUsers, userID)); usrTxUID != 0 {
					tr.Fail(errUserHasBeenRegistered)
				}
				tr.PutID(goldb.Key(dbTabUsers, userID), txUID) //
			}

			// put index transaction by txID
			tr.PutVar(goldb.Key(dbIdxTxID, txID), txUID)

			// verify state of transaction
			if config.VerifyTransactions {

				// make state by dbTransaction
				st := state.NewState(func(a assets.Asset, addr crypto.Address) state.Number { // get state from db
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
				if !it.State.Equal(newState) {
					tr.Fail(errIncorrectTxState)
				}
			}

			// save state to db-storage
			for seq, v := range it.State.Values() {

				tr.PutVar(goldb.Key(dbIdxAsset, v.Asset, txUID, seq, v.Address), v.Value)

				tr.PutVar(goldb.Key(dbIdxAssetAddr, v.Asset, v.Address, txUID, seq), v.Value)

				if v.Tag != 0 { // change state with tag
					tr.PutVar(goldb.Key(dbIdxAssetAddrTag, v.Asset, v.Address, v.Tag, txUID, seq), v.Value)
				}
			}
		}

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
	err = s.FetchBlocks(0, true, 1, func(b *blockchain.Block) error {
		block = b
		return nil
	})
	if block == nil {
		block = blockchain.GenesisBlock()
	}
	return
}

func (s *BlockchainStorage) GetBlock(num uint64) (*blockchain.Block, error) {
	if num == 0 {
		return blockchain.GenesisBlock(), nil
	}
	block := new(blockchain.Block)
	if ok, err := s.storage.GetVar(goldb.Key(dbTabBlock, num), block); err != nil {
		return nil, err
	} else if !ok {
		return nil, errBlockNotFound
	}
	return block, nil
}

func (s *BlockchainStorage) FetchBlocks(offset uint64, desc bool, limit int64, fn func(block *blockchain.Block) error) error {
	q := goldb.NewQuery(dbTabBlock)
	if offset > 0 {
		q.Offset(offset)
	}
	q.Order(desc)
	if limit > 0 {
		q.Limit(limit)
	}
	return s.storage.Fetch(q, func(rec goldb.Record) error {
		var block = new(blockchain.Block)
		rec.Decode(block)
		return fn(block)
	})
}

func (s *BlockchainStorage) TransactionByHash(txHash []byte) (*blockchain.BlockItem, error) {
	it, err := s.TransactionByID(blockchain.TxIDByHash(txHash))
	if err == nil && it != nil && !bytes.Equal(txHash, it.TxHash()) { // collision
		return nil, nil
	}
	return nil, err
}

func (s *BlockchainStorage) TransactionByID(txID uint64) (*blockchain.BlockItem, error) {
	if txUID, err := s.storage.GetID(goldb.Key(dbIdxTxID, txID)); err != nil {
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

func (s *BlockchainStorage) FetchTxUID(
	asset assets.Asset,
	addr crypto.Address,
	tag int64,
	offset uint64,
	desc bool,
	limit int64,
	fn func(txUID uint64, val state.Number) error,
) error {

	typ, q := dbIdxAsset, goldb.NewQuery(dbIdxAsset, asset)
	if tag != 0 { // fetch transactions by address tag
		typ, q = dbIdxAssetAddrTag, goldb.NewQuery(dbIdxAssetAddrTag, asset, addr, tag)
	} else if !addr.IsNil() { // get address history
		typ, q = dbIdxAssetAddr, goldb.NewQuery(dbIdxAssetAddr, asset, addr)
	}
	if offset > 0 {
		q.Offset(offset)
	}
	q.Order(desc)
	if limit > 0 {
		q.Limit(limit)
	}
	var txUID uint64
	return s.storage.Fetch(q, func(rec goldb.Record) error {
		switch typ { // get txUID from record-key
		case dbIdxAsset:
			rec.DecodeKey(&asset, &txUID)
		case dbIdxAssetAddr:
			rec.DecodeKey(&asset, &addr, &txUID)
		case dbIdxAssetAddrTag:
			rec.DecodeKey(&asset, &addr, &tag, &txUID)
		}
		return fn(txUID, rec.ValueBigInt())
	})
}

func (s *BlockchainStorage) FetchTransaction(
	asset assets.Asset,
	addr crypto.Address,
	tag int64,
	offset uint64,
	desc bool,
	limit int64,
	fn func(tx *blockchain.BlockItem, val state.Number) error,
) error {
	return s.FetchTxUID(asset, addr, tag, offset, desc, limit, func(txUID uint64, val state.Number) error {
		if tx, err := s.TransactionByUID(txUID); err != nil {
			return err
		} else {
			return fn(tx, val)
		}
	})
}
