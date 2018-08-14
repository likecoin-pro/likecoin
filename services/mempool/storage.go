package mempool

import (
	"sync"

	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/crypto"
)

type Storage struct {
	mx   sync.RWMutex
	vals []*blockchain.Transaction
}

type Info struct {
	Size int `json:"size"`
}

func NewStorage() *Storage {
	return &Storage{}
}

func (s *Storage) Info() (i Info) {
	i.Size = s.Size()
	return
}

func (s *Storage) Size() int {
	s.mx.RLock()
	defer s.mx.RUnlock()
	return len(s.vals)
}

// todo: add counters by txType
func (s *Storage) SizeOf(txType blockchain.TxType) (count int) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	for _, tx := range s.vals {
		if tx.Type == txType {
			count++
		}
	}
	return
}

func (s *Storage) Put(tx *blockchain.Transaction) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.vals = append(s.vals, tx)
	return nil
}

func (s *Storage) Pop() (tx *blockchain.Transaction) {
	s.mx.Lock()
	defer s.mx.Unlock()
	if len(s.vals) > 0 {
		tx = s.vals[0]
		s.vals = s.vals[1:]
	}
	return
}

func (s *Storage) PopAll() (txs []*blockchain.Transaction) {
	s.mx.Lock()
	defer s.mx.Unlock()
	txs = s.vals
	s.vals = nil
	return
}

func (s *Storage) TxsByAddress(addr crypto.Address) (txs []*blockchain.Transaction, err error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	for _, tx := range s.vals {
		if tx.SenderAddress() == addr {
			txs = append(txs, tx)
		}
	}
	return
}

func (s *Storage) TxsAll() (txs []*blockchain.Transaction, err error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	txs = make([]*blockchain.Transaction, len(s.vals))
	copy(txs, s.vals)
	return
}
