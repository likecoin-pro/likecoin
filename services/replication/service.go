package replication

import (
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
		ok, err := r.loadBlock(r.bc.LastBlock().Num + 1)
		if err != nil {
			log.Error.Printf("replication> loadBlock Error: %v", err)
		}
		if !ok || err != nil {
			//log.Error.Printf("replication> sleep...")
			print(".")
			time.Sleep(1e9)
		}
	}
}

func (r *Service) loadBlock(num uint64) (ok bool, err error) {
	block, err := r.client.GetBlock(num)
	if err != nil {
		log.Error.Printf("replication> client.GetBlock. Error: %v", err)
		return
	}
	if block == nil {
		return
	}
	if err = r.bc.PutBlock(block); err != nil {
		log.Error.Printf("replication> bc.PutBlock. Error: %v", err)
		return
	}
	log.Printf("replication> ✅ replicated block#%d ", num)
	return true, nil
}
