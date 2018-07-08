package blockchain

import "github.com/likecoin-pro/likecoin/config"

func GenesisBlock() *Block {
	return &Block{
		Version:   0,
		Num:       0,
		ChainID:   config.ChainID,
		Network:   config.NetworkID,
		Timestamp: 0,
		PrevHash:  nil,
		TxRoot:    nil,
		Nonce:     0,
		Miner:     config.MasterPublicKey,
	}
}
