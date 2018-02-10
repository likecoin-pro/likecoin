package db

import (
	"bytes"
	"errors"

	"math/big"

	"github.com/denisskin/goldb"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
)

type BlockchainStorage struct {
	storage   *goldb.Storage
	lastBlock *blockchain.Block
}

const (
	dbTabBlock = 0x01
	//dbTabState = 0x02

	dbIdxTxID  = 0x10
	dbIdxState = 0x11
	//dbIdxAddressNumTs = 0x12
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
			trState := state.NewStateEx(func(key state.Key) state.Number { // get state from db
				//val, err := tr.GetInt(goldb.Key(dbTabDBState, key))
				var v *big.Int
				if err := tr.QueryValue(goldb.NewQuery(dbIdxState, key).Last(), &v); err != nil {
					panic(err) // dbTransaction fail
				}
				return v

			}, func(key state.Key, val state.Number) { // set state to db.tx
				if err := tr.PutBigInt(goldb.Key(dbIdxState, key, txUID), val); err != nil {
					panic(err) // dbTransaction fail
				}
			})

			it.Tx.Execute(trState)

			if !it.State.Equal(trState) {
				panic(errIncorrectTxState) // dbTransaction fail
			}

			// todo: index transactions
			// todo: index state
		}

		s.lastBlock = block
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

func (s *BlockchainStorage) TxByHash(txHash []byte) (*blockchain.BlockItem, error) {
	it, err := s.TxByID(blockchain.TxIDByHash(txHash))
	if err == nil && it != nil && !bytes.Equal(txHash, it.Tx.Hash()) { // collision
		return nil, nil
	}
	return nil, err
}

func (s *BlockchainStorage) TxByID(txID uint64) (*blockchain.BlockItem, error) {
	if txUID, err := s.storage.GetID(goldb.Key(dbIdxTxID, txID)); err != nil {
		return nil, err
	} else {
		return s.txByUID(txUID)
	}
}

func (s *BlockchainStorage) txByUID(txUID uint64) (it *blockchain.BlockItem, err error) {
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
