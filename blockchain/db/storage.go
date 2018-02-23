package db

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/denisskin/goldb"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/crypto"
)

type BlockchainStorage struct {
	storage   *goldb.Storage
	lastBlock *blockchain.Block
}

const (
	dbTabBlock = 0x01

	dbIdxTxID     = 0x10
	dbIdxState    = 0x11
	dbIdxStateTag = 0x12
)

var (
	errBlockNotFound       = errors.New("block not found")
	errTxHasBeenRegistered = errors.New("tx has been registered")
	errTxNotFound          = errors.New("tx not found")
	errIncorrectTxState    = errors.New("incorrect tx state")
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
			panic(err) // dbTransaction fail
		}

		// put block
		tr.PutVar(goldb.Key(dbTabBlock, block.Num), block)

		// add index on transactions
		for txIdx, it := range block.Items {

			txID := blockchain.TxID(it.Tx)
			txUID := blockchain.EncodeTxUID(block.Num, txIdx)

			// check transaction by txID
			if id, err := tr.GetInt(goldb.Key(dbIdxTxID, txID)); err != nil {
				panic(err)
			} else if id != 0 {
				panic(errTxHasBeenRegistered)
			}
			// put index transaction by txID
			tr.PutVar(goldb.Key(dbIdxTxID, txID), txUID)

			// verify tx & tx-states
			trState := state.NewStateEx(
				// get state from db
				func(key state.Key) state.Number {
					var v *big.Int
					if err := tr.QueryValue(goldb.NewQuery(dbIdxState, key).Last(), &v); err != nil {
						panic(err) // dbTransaction fail
					}
					return v

				},
				// set state to db.Transaction
				func(key state.Key, val state.Number, tag int64) {
					if err := tr.PutVar(goldb.Key(dbIdxState, key, txUID), val); err != nil {
						panic(err) // dbTransaction fail
					}
					if tag != 0 { // change state with tag
						if err := tr.PutVar(goldb.Key(dbIdxStateTag, key, tag, txUID), val); err != nil {
							panic(err) // dbTransaction fail
						}
					}
				},
			)

			it.Tx.Execute(trState)

			if !it.State.Equal(trState) {
				panic(errIncorrectTxState) // dbTransaction fail
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
	var block = new(blockchain.Block)
	return s.storage.FetchObject(q, block, func() error {
		return fn(block)
	})
}

func (s *BlockchainStorage) TransactionByHash(txHash []byte) (*blockchain.BlockItem, error) {
	it, err := s.TransactionByID(blockchain.TxIDByHash(txHash))
	if err == nil && it != nil && !bytes.Equal(txHash, it.Tx.Hash()) { // collision
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
	asset crypto.Asset,
	addr crypto.Address,
	tag int64,
	offset uint64,
	desc bool,
	limit int64,
	fn func(txUID uint64, val state.Number) error,
) error {
	q := goldb.NewQuery(dbIdxState, state.NewKey(addr, asset))
	filterByTag := tag != 0
	if filterByTag {
		q = goldb.NewQuery(dbIdxStateTag, state.NewKey(addr, asset), tag)
	}
	if offset > 0 {
		q.Offset(offset)
	}
	q.Order(desc)
	if limit > 0 {
		q.Limit(limit)
	}
	return s.storage.Fetch(q, func(rec goldb.Record) (err error) {
		var key state.Key
		var t int64
		var txUID uint64
		var val state.Number
		if filterByTag {
			err = rec.DecodeKey(&key, &t, &txUID)
		} else {
			err = rec.DecodeKey(&key, &txUID)
		}
		if err != nil {
			return
		}
		if err = rec.Decode(&val); err != nil {
			return
		}
		return fn(txUID, val)
	})
}

func (s *BlockchainStorage) FetchTransaction(
	asset crypto.Asset,
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
