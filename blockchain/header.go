package blockchain

import (
	"bytes"
	"fmt"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/likecoin-pro/likecoin/crypto/merkle"
)

type BlockHeader struct {
	Version   int       `json:"version"`       // version
	Network   int       `json:"network"`       // networkID
	ChainID   uint64    `json:"chain"`         //
	Num       uint64    `json:"height"`        // number of block in blockchain
	Timestamp int64     `json:"timestamp"`     // timestamp of block in µsec
	PrevHash  bin.Bytes `json:"previous_hash"` // hash of previous block
	TxRoot    bin.Bytes `json:"tx_root"`       // merkle root of block-transactions
	StateRoot bin.Bytes `json:"state_root"`    // patricia root of global state
	ChainRoot bin.Bytes `json:"chain_root"`    // patricia root of chain

	// miner params
	Nonce uint64            `json:"nonce"` //
	Miner *crypto.PublicKey `json:"miner"` // miner public-key
	Sig   bin.Bytes         `json:"sig"`   // miner signature  := minerKey.Sign( blockHash + chainRoot )

	// reserved
	Reserved1 []byte `json:"-"`
	Reserved2 []byte `json:"-"`
	Reserved3 []byte `json:"-"`
}

func (b *BlockHeader) String() string {
	h := b.Hash()
	return fmt.Sprintf("[BLOCK-%d 0x%x size:%d]", b.Num, h[:8], b.Size())
}

// block.Hash | chainRoot
func (b *BlockHeader) sigHash() []byte {
	return merkle.Root(b.Hash(), b.ChainRoot)
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

// Size returns block-header size
func (b *BlockHeader) Size() int {
	return len(b.Encode())
}

func (b *BlockHeader) Encode() []byte {
	return bin.Encode(
		b.Version,
		b.Network,
		b.ChainID,
		b.Num,
		b.Timestamp,
		b.PrevHash,
		b.TxRoot,
		b.StateRoot,
		b.ChainRoot,
		b.Nonce,
		b.Miner,
		b.Reserved1,
		b.Reserved2,
		b.Reserved3,
		b.Sig,
	)
}

func (b *BlockHeader) Decode(data []byte) (err error) {
	return bin.Decode(data,
		&b.Version,
		&b.Network,
		&b.ChainID,
		&b.Num,
		&b.Timestamp,
		&b.PrevHash,
		&b.TxRoot,
		&b.StateRoot,
		&b.ChainRoot,
		&b.Nonce,
		&b.Miner,
		&b.Reserved1,
		&b.Reserved2,
		&b.Reserved3,
		&b.Sig,
	)
}

func (b *BlockHeader) VerifyHeader(pre *BlockHeader, cfg *Config) error {
	if b.Network != cfg.NetworkID {
		return ErrInvalidNetwork
	}
	if b.ChainID != cfg.ChainID {
		return ErrInvalidChainID
	}
	blockHash := b.Hash()
	if b.Num == 0 && bytes.Equal(blockHash, GenesisBlockHeader(cfg).Hash()) { // is genesis
		return ErrInvalidGenesisBlock
	}
	if pre != nil {
		if b.Network != pre.Network {
			return ErrInvalidNetwork
		}
		if b.ChainID != pre.ChainID {
			return ErrInvalidChainID
		}
		if b.Num != pre.Num+1 {
			return ErrInvalidBlockNum
		}
		if b.Timestamp < pre.Timestamp {
			return ErrInvalidBlockTs
		}
		if !bytes.Equal(b.PrevHash, pre.Hash()) {
			return ErrInvalidPrevHash
		}
	}
	if b.Miner.Empty() {
		return ErrEmptyMinerKey
	}
	if !b.Miner.Equal(config.MasterPublicKey) {
		return ErrInvalidMinerKey
	}
	if !b.Miner.Verify(b.sigHash(), b.Sig) {
		return ErrInvalidBlockSig
	}
	return nil
}
