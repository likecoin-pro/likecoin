package blockchain

import (
	"bytes"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
)

type BlockHeader struct {
	Version   int       `json:"version"`       // version
	ChainID   uint64    `json:"chain"`         //
	Num       uint64    `json:"height"`        // number of block in blockchain
	Timestamp int64     `json:"timestamp"`     // timestamp of block in µsec
	PrevHash  bin.Bytes `json:"previous_hash"` // hash of previous block
	TxRoot    bin.Bytes `json:"tx_root"`       // merkle root of transactions in block
	StateRoot bin.Bytes `json:"state_root"`    // patricia root of global state

	// not imported
	ChainRoot bin.Bytes `json:"chain_root"` // patricia root of chain (just info; not encoding)

	// miner params
	Nonce uint64            `json:"nonce"`     //
	Miner *crypto.PublicKey `json:"miner"`     // pub-key of miner
	Sign  bin.Bytes         `json:"signature"` // miner-node sign

	// reserved
	Reserved1 []byte `json:"-"`
	Reserved2 []byte `json:"-"`
	Reserved3 []byte `json:"-"`
}

func (b *BlockHeader) Hash() []byte {
	return crypto.Hash256(
		b.Version,
		b.ChainID,
		b.Num,
		b.Timestamp,
		b.PrevHash,
		b.TxRoot,
		b.StateRoot,
		b.Nonce,
		b.Miner,
		b.Reserved1,
		b.Reserved2,
		b.Reserved3,
	)
}

func (b *BlockHeader) Encode() []byte {
	return bin.Encode(
		b.Version,
		b.ChainID,
		b.Num,
		b.Timestamp,
		b.PrevHash,
		b.TxRoot,
		b.StateRoot,
		b.Nonce,
		b.Miner,
		b.Reserved1,
		b.Reserved2,
		b.Reserved3,
		b.Sign,
	)
}

func (b *BlockHeader) Decode(data []byte) error {
	return bin.Decode(data,
		&b.Version,
		&b.ChainID,
		&b.Num,
		&b.Timestamp,
		&b.PrevHash,
		&b.TxRoot,
		&b.StateRoot,
		&b.Nonce,
		&b.Miner,
		&b.Reserved1,
		&b.Reserved2,
		&b.Reserved3,
		&b.Sign,
	)
}

func (b *BlockHeader) Verify(pre *BlockHeader) error {
	hash := b.Hash()
	if b.Num == 0 && bytes.Equal(hash, genesisBlockHeaderHash) { // is genesis
		return ErrInvalidGenesisBlock
	}
	if pre != nil {
		if b.ChainID != pre.ChainID {
			return ErrInvalidChainID
		}
		if b.Num != pre.Num+1 {
			return ErrInvalidNum
		}
		if !bytes.Equal(b.PrevHash, pre.Hash()) {
			return ErrInvalidPrevHash
		}
	}
	if b.Miner.Empty() {
		return ErrEmptyNodeKey
	}
	if !b.Miner.Equal(config.MasterPublicKey) {
		return ErrInvalidNodeKey
	}
	if !b.Miner.Verify(hash, b.Sign) {
		return ErrInvalidSign
	}
	return nil
}
