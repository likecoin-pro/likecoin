package blockchain

import (
	"bytes"
	"encoding/hex"
	"encoding/json"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/commons/merkle"
	"github.com/likecoin-pro/likecoin/crypto"
)

type Block struct {
	BlockHeader
	Items []*BlockItem
}

func (b *Block) NewBlock() *Block {
	return &Block{
		BlockHeader: BlockHeader{
			Version:  0,
			Num:      b.Num + 1,
			PrevHash: b.Hash(),
		},
	}
}

func (b *Block) VerifyHeader(pre *Block) error {

	if len(b.Items) == 0 {
		return ErrEmptyBlock
	}
	if !bytes.Equal(b.MerkleRoot, b.merkleRoot()) {
		return ErrInvalidMerkleRoot
	}

	// verify block header
	if err := b.BlockHeader.Verify(&pre.BlockHeader); err != nil {
		return err
	}

	return nil
}

func (b *Block) merkleRoot() []byte {
	var hh [][]byte
	for _, it := range b.Items {
		hh = append(hh, it.Hash())
	}
	return merkle.Root(hh)
}

func (b *Block) AddTx(st *state.State, tx Transaction) *BlockItem {
	txState := st.NewSubState()
	tx.Execute(txState)

	it := &BlockItem{
		Tx:    tx,
		Ts:    timestamp(),
		State: txState,

		block:    b,
		blockIdx: len(b.Items),
	}
	b.Items = append(b.Items, it)
	return it
}

func (b *Block) SetSign(prv *crypto.PrivateKey) {
	b.MerkleRoot = b.merkleRoot()
	b.Timestamp = timestamp()
	b.Nonce = 0
	b.Node = prv.PublicKey
	b.Sign = prv.Sign(b.Hash())
}

func (b *Block) Encode() []byte {
	return bin.Encode(

		// header
		b.Version,
		b.Num,
		b.Timestamp,
		b.PrevHash,
		b.MerkleRoot,
		b.Nonce,
		b.Node,
		b.Sign,

		// items
		b.Items,
	)
}

func (b *Block) Decode(data []byte) error {
	err := bin.Decode(data,

		// header
		&b.Version,
		&b.Num,
		&b.Timestamp,
		&b.PrevHash,
		&b.MerkleRoot,
		&b.Nonce,
		&b.Node,
		&b.Sign,

		// items
		&b.Items,
	)
	for i, it := range b.Items {
		it.block = b
		it.blockIdx = i
	}
	return err
}

func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Version    int          `json:"version"`
		Num        uint64       `json:"num"`
		Timestamp  int64        `json:"timestamp"`
		PrevHash   string       `json:"prev_hash"`
		MerkleRoot string       `json:"merkle_root"`
		Nonce      uint64       `json:"nonce"`
		Node       string       `json:"node"`
		Sign       string       `json:"sign"`
		Items      []*BlockItem `json:"items"`
	}{
		Version:    b.Version,
		Num:        b.Num,
		Timestamp:  b.Timestamp,
		PrevHash:   hex.EncodeToString(b.PrevHash),
		MerkleRoot: hex.EncodeToString(b.MerkleRoot),
		Nonce:      b.Nonce,
		Node:       b.Node.String(),
		Sign:       hex.EncodeToString(b.Sign),
		Items:      b.Items,
	})
}
