package db

import (
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/services/mempool"
)

type Info struct {
	ChainID   uint64            `json:"chain"`      //
	Stat      *Statistic        `json:"stat"`       //
	LastBlock *blockchain.Block `json:"last_block"` //
	Mempool   mempool.Info      `json:"mempool"`    //
}

func (s *BlockchainStorage) Info() (inf Info, err error) {
	inf.ChainID = s.chainID
	inf.Stat = s.Totals()
	inf.LastBlock = s.LastBlock()
	inf.Mempool = s.Mempool.Info()
	return
}
