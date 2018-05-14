package blockchain

import "github.com/likecoin-pro/likecoin/config"

var genesisBlock = &Block{
	Version:   0,
	Num:       0,
	ChainID:   1,
	Timestamp: 0,
	PrevHash:  nil,
	TxRoot:    nil,
	Nonce:     0,
	Miner:     config.MasterPublicKey,
}

var genesisBlockHash = genesisBlock.Hash()

func GenesisBlock() *Block {
	return genesisBlock
}
