package replication

import (
	"fmt"
	"time"

	"github.com/likecoin-pro/likecoin/services/client"

	"github.com/likecoin-pro/likecoin/blockchain/db"
	"github.com/likecoin-pro/likecoin/commons/log"
)

type Service struct {
	client *client.Client
	bc     *db.BlockchainStorage
}

func NewService(client *client.Client, bc *db.BlockchainStorage) *Service {
	return &Service{
		client: client,
		bc:     bc,
	}
}

func (r *Service) StartReplication() {
	for {
		ok, err := r.loadBatch(r.bc.LastBlock().Num, 100)
		if err != nil {
			log.Error.Printf("replication> loadBatch Error: %v", err)
		}
		if !ok || err != nil {
			time.Sleep(time.Second)
		}
	}
}

func (r *Service) loadBatch(blockOffset uint64, batchSize int) (ok bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("replication.loadBatch-Panic: %v", r)
		}
	}()

	blocks, err := r.client.GetBlocks(blockOffset, batchSize)
	if err != nil {
		log.Error.Printf("replication> client.GetBlocks. Error: %v", err)
		return
	}
	if len(blocks) == 0 {
		return
	}
	if err = r.bc.PutBlocks(blocks); err != nil {
		log.Error.Printf("replication> bc.PutBlocks. Error: %v", err)
		return
	}
	log.Printf("replication> ✅ replicated block#%d ", blocks[len(blocks)-1].Num)
	return true, nil
}
