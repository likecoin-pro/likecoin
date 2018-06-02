package mempool

import (
	"sync"

	"github.com/likecoin-pro/likecoin/blockchain"
)

type Storage struct {
	mx   sync.RWMutex
	vals []*blockchain.BlockTx
}

func NewStorage() *Storage {
	return &Storage{}
}

func (s *Storage) Put(tx *blockchain.BlockTx) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.vals = append(s.vals, tx)
	return nil
}

func (s *Storage) Pop() (tx *blockchain.BlockTx) {
	s.mx.Lock()
	defer s.mx.Unlock()
	if len(s.vals) > 0 {
		tx = s.vals[0]
		s.vals = s.vals[1:]
	}
	return
}

func (s *Storage) PopAll() (txs []*blockchain.BlockTx) {
	s.mx.Lock()
	defer s.mx.Unlock()
	txs = s.vals
	s.vals = nil
	return
}
