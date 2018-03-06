package blockchain

import "github.com/likecoin-pro/likecoin/config"

var genesisBlockHeader = BlockHeader{
	Version:    0,
	Num:        0,
	Timestamp:  0,
	PrevHash:   nil,
	MerkleRoot: nil,
	Nonce:      0,
	Miner:      config.MasterPublicKey,
}

var genesisBlockHeaderHash = genesisBlockHeader.Hash()

func GenesisBlock() *Block {
	return &Block{BlockHeader: genesisBlockHeader}
}
