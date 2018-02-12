package blockchain

import (
	"bytes"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
)

type BlockHeader struct {
	Version    int    // version
	Num        uint64 // number of block in chain
	Timestamp  int64  // timestamp of block in microSec
	PrevHash   []byte // hash of previous block
	MerkleRoot []byte // merkle hash of transactions

	// node sign
	Nonce uint64            //
	Node  *crypto.PublicKey // pub-key of master node
	Sign  []byte            // master-node sign
}

func (b *BlockHeader) Hash() []byte {
	return bin.Hash256(
		b.Version,
		b.Num,
		b.Timestamp,
		b.PrevHash,
		b.MerkleRoot,
		b.Nonce,
		b.Node,
	)
}

func (b *BlockHeader) Verify(pre *BlockHeader) error {
	hash := b.Hash()
	if b.Num == 0 && bytes.Equal(hash, genesisBlockHeaderHash) { // is genesis
		return ErrInvalidGenesisBlock
	}
	if pre != nil {
		if b.Num != pre.Num+1 {
			return ErrInvalidNum
		}
		if !bytes.Equal(b.PrevHash, pre.Hash()) {
			return ErrInvalidPrevHash
		}
	}
	if b.Node.Empty() {
		return ErrEmptyNodeKey
	}
	if !b.Node.Equal(config.MasterPublicKey) {
		return ErrInvalidNodeKey
	}
	if !b.Node.Verify(hash, b.Sign) {
		return ErrInvalidSign
	}
	return nil
}
