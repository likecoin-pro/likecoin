package db

import (
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/services/mempool"
)

type Info struct {
	Version   string                  `json:"ver"`        //
	ChainID   uint64                  `json:"chain"`      //
	Stat      *Statistic              `json:"stat"`       //
	LastBlock *blockchain.BlockHeader `json:"last_block"` //
	Mempool   mempool.Info            `json:"mempool"`    //
}

func (s *BlockchainStorage) Info() (inf Info, err error) {
	inf.Version = config.Version
	inf.ChainID = s.Cfg.ChainID
	inf.Stat = s.Totals()
	inf.LastBlock = s.LastBlock().BlockHeader
	inf.Mempool = s.Mempool.Info()
	return
}
