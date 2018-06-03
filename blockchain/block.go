package blockchain

import (
	"bytes"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/likecoin-pro/likecoin/crypto/merkle"
	"github.com/likecoin-pro/likecoin/crypto/patricia"
)

type Block struct {
	Version   int       `json:"version"`       // version
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

func NewBlock(
	pre *Block,
	txs []*BlockTx,
	prv *crypto.PrivateKey,
	state *state.State,
	chainTree *patricia.Tree,
) (block *Block, err error) {

	for _, bTx := range txs {
		bTx.StateUpdates, err = bTx.Tx.Execute(state)
		if err != nil {
			return
		}
		state.Apply(bTx.StateUpdates)
	}

	block = &Block{
		Version:   0,
		ChainID:   pre.ChainID,
		Num:       pre.Num + 1,
		PrevHash:  pre.Hash(),
		TxRoot:    txRoot(txs),
		Timestamp: timestamp(),
		Nonce:     0,
		Miner:     prv.PublicKey,
	}

	err = chainTree.PutVar(block.Num, block.Hash())
	if err != nil {
		return nil, err
	}
	block.ChainRoot, err = chainTree.Root()
	if err != nil {
		return nil, err
	}

	// set signature( b.Hash + chainRoot )
	block.Sig = prv.Sign(block.sigHash())

	return
}

// block.Hash + chainRoot
func (b *Block) sigHash() []byte {
	return merkle.Root(b.Hash(), b.ChainRoot)
}

func (b *Block) Hash() []byte {
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
func (b *Block) Size() int64 {
	return int64(len(b.Encode()))
}

func (b *Block) Encode() []byte {
	return bin.Encode(
		b.Version,
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

func (b *Block) Decode(data []byte) (err error) {
	return bin.Decode(data,
		&b.Version,
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

func (b *Block) Verify(pre *Block) error {
	blockHash := b.Hash()
	if b.Num == 0 && bytes.Equal(blockHash, genesisBlockHash) { // is genesis
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
	if !b.Miner.Verify(b.sigHash(), b.Sig) {
		return ErrInvalidSign
	}
	return nil
}

func (b *Block) VerifyTxs(txs []*BlockTx) error {

	if len(txs) == 0 {
		return ErrEmptyBlock
	}

	for _, tx := range txs {
		// check tx-chain info
		if tx.Tx.ChainID != b.ChainID {
			return ErrInvalidChainID
		}
	}

	if txRoot := txRoot(txs); !bytes.Equal(b.TxRoot, txRoot) {
		return ErrInvalidMerkleRoot
	}

	return nil
}

func txRoot(txs []*BlockTx) []byte {
	var hh [][]byte
	for _, it := range txs {
		hh = append(hh, it.Hash())
	}
	return merkle.Root(hh...)
}
