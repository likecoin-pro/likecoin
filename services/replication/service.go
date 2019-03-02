package replication

import (
	"fmt"
	"time"

	"github.com/likecoin-pro/likecoin/blockchain/db"
	"github.com/likecoin-pro/likecoin/commons/log"
	"github.com/likecoin-pro/likecoin/services/client"
)

type Service struct {
	client *client.Client
	bc     *db.BlockchainStorage
}

func NewService(cl *client.Client, bc *db.BlockchainStorage) *Service {
	if cl == nil {
		cl = client.NewClient("")
	}
	return &Service{
		client: cl,
		bc:     bc,
	}
}

func (s *Service) StartReplication() {
	go s.startBlockchainReplication()
	go s.startMempoolReplication()
}

func (s *Service) startBlockchainReplication() {
	for {
		nBlocks, err := s.loadBlocksBatch(s.bc.LastBlock().Num, 100)
		if err != nil {
			log.Error.Printf("replication> loadBlocksBatch Error: %v", err)
		}
		if nBlocks == 0 || err != nil {
			time.Sleep(5 * time.Second)
		}
	}
}

func (s *Service) loadBlocksBatch(blockOffset uint64, batchSize int) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("replication> loadBlocksBatch-Panic: %v", r)
		}
	}()

	blocks, err := s.client.GetBlocks(blockOffset, batchSize)
	if err != nil {
		log.Error.Printf("replication> client.GetBlocks-Error: %v", err)
		return
	}
	if len(blocks) == 0 {
		return
	}
	if err = s.bc.PutBlock(blocks...); err != nil {
		log.Error.Printf("replication> bc.PutBlock-Error: %v", err)
		return
	}
	log.Printf("replication> ✅ replicated block#%d ", blocks[len(blocks)-1].Num)
	return len(blocks), nil
}

func (s *Service) startMempoolReplication() {

	// todo: (it`s temporary scheme) refactor me! use decentralize replication;

	for ; ; time.Sleep(time.Second) {

		ok, err := s.putMempoolTxs()
		if err != nil {
			log.Error.Printf("replication> loadBlocksBatch-Error: %v", err)
		}
		if !ok || err != nil {
			time.Sleep(time.Second)
		}
	}
}

func (s *Service) putMempoolTxs() (ok bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("replication> loadBlocksBatch-Panic: %v", r)
		}
	}()

	//-- get from local mempool
	txs, _ := s.bc.Mempool.AllTxs()
	if len(txs) == 0 {
		return
	}
	//-- put to remote node
	err = s.client.PutTxs(txs)
	if err != nil {
		log.Error.Printf("replication> client.PutTxs-Error: %v", err)
		return
	}
	//-- remove from mempool
	txIDs := make([]uint64, len(txs))
	for _, tx := range txs {
		txIDs = append(txIDs, tx.ID())
	}
	s.bc.Mempool.RemoveTxs(txIDs)

	log.Printf("replication> putTxs(%d). OK", len(txs))
	return true, nil
}
