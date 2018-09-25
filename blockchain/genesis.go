package blockchain

import "github.com/likecoin-pro/likecoin/config"

func GenesisBlockHeader(cfg *Config) *BlockHeader {
	return &BlockHeader{
		Version:   0,
		Num:       0,
		ChainID:   cfg.ChainID,
		Network:   cfg.NetworkID,
		Timestamp: 0,
		PrevHash:  nil,
		TxRoot:    nil,
		Nonce:     0,
		Miner:     config.MasterPublicKey,
	}
}
