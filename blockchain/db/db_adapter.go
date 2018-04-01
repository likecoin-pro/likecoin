package db

import (
	"github.com/denisskin/goldb"
	"github.com/likecoin-pro/likecoin/crypto/patricia"
)

// implement patricia.Storage interface for db transaction or storage
type dbAdapter struct {
	tab goldb.Entity
	db  patricia.Storage // db.transaction or db.storage
}

func newPatriciaTree(db patricia.Storage, tab goldb.Entity) *patricia.Tree {
	return patricia.NewTree(&dbAdapter{tab, db})
}

func (s *dbAdapter) Get(key []byte) ([]byte, error) {
	return s.db.Get(goldb.Key(s.tab, key))
}

func (s *dbAdapter) Put(key, data []byte) error {
	return s.db.Put(goldb.Key(s.tab, key), data)
}

//func (s *dbAdapter) Delete(key []byte) error {
//	return s.db.Delete(goldb.Key(s.tab, key))
//}

func (s *BlockchainStorage) ChainTree() *patricia.Tree {
	return patricia.NewTree(&dbAdapter{dbTabChainTree, s.db})
}

func (s *BlockchainStorage) StateTree() *patricia.Tree {
	return patricia.NewTree(&dbAdapter{dbTabStateTree, s.db})
}

//func (s *BlockchainStorage) AddressBalanceProof(addr crypto.Address, a assets.Asset) (proof []byte, err error) {
//	bal, lastTx, err := s.GetBalance(addr, a)
//	if err != nil {
//		return
//	}
//	lastTx.State.Values()
//	return patricia.NewTree(&dbAdapter{dbTabStateTree, s.db})
//}
