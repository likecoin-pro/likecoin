package blockchain

import (
	"bytes"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/blockchain/transaction"
	"github.com/likecoin-pro/likecoin/commons/merkle"
	"github.com/likecoin-pro/likecoin/crypto"
)

type Block struct {
	*BlockHeader
	Items []*BlockItem `json:"txs"`
}

func (b *Block) NewBlock() *Block {
	return &Block{
		BlockHeader: &BlockHeader{
			Version:  0,
			ChainID:  b.ChainID,
			Num:      b.Num + 1,
			PrevHash: b.Hash(),
		},
	}
}

func (b *Block) VerifyHeader(pre *Block) error {

	if len(b.Items) == 0 {
		return ErrEmptyBlock
	}
	if !bytes.Equal(b.TxRoot, b.txRoot()) {
		return ErrInvalidMerkleRoot
	}

	// verify block header
	if err := b.BlockHeader.Verify(pre.BlockHeader); err != nil {
		return err
	}

	return nil
}

func (b *Block) txRoot() []byte {
	var hh [][]byte
	for _, it := range b.Items {
		hh = append(hh, it.Hash())
	}
	return merkle.Root(hh)
}

func (b *Block) AddTx(st *state.State, tx transaction.Transaction) (it *BlockItem, err error) {
	txState, err := st.Execute(tx)
	if err != nil {
		return
	}
	it = &BlockItem{
		Tx:    tx,
		State: txState,

		block:    b,
		blockIdx: len(b.Items),
	}
	b.Items = append(b.Items, it)
	return
}

func (b *Block) SetSign(prv *crypto.PrivateKey) {
	b.TxRoot = b.txRoot()
	b.Timestamp = timestamp()
	b.Nonce = 0
	b.Miner = prv.PublicKey
	b.Sign = prv.Sign(b.Hash())
}

func (b *Block) Size() int {
	return len(b.Encode())
}

func (b *Block) Encode() []byte {
	return bin.Encode(
		b.BlockHeader,
		b.Items,
	)
}

func (b *Block) Decode(data []byte) error {
	err := bin.Decode(data,
		&b.BlockHeader,
		&b.Items,
	)
	for i, it := range b.Items {
		it.block = b
		it.blockIdx = i
	}
	return err
}
