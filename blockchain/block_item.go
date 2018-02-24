package blockchain

import (
	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/blockchain/state"
)

type BlockItem struct {
	Ts    int64        `json:"timestamp"` // timestamp in microsec
	Tx    Transaction  `json:"tx"`        //
	State *state.State `json:"state"`     // new state values

	block    *Block `json:"-"` // link on parent-block
	blockIdx int    `json:"-"` //
}

func (it *BlockItem) UID() uint64 {
	return EncodeTxUID(it.block.Num, it.blockIdx)
}

func (it *BlockItem) Hash() []byte {
	return bin.Hash256(
		it.Ts,
		it.Tx.Hash(),
		it.State,
	)
}

func (it *BlockItem) Encode() []byte {
	return bin.Encode(
		it.Ts,
		int(it.Tx.Type()),
		it.Tx,
		it.State,
	)
}

func (it *BlockItem) Decode(data []byte) (err error) {
	var txType int

	r := bin.NewBuffer(data)
	r.ReadVar(&it.Ts)
	r.ReadVar(&txType)
	if it.Tx, err = newTransaction(TxType(txType)); err != nil {
		return
	}
	r.ReadVar(it.Tx)
	r.ReadVar(&it.State)
	return r.Error()
}

func EncodeTxUID(blockNum uint64, txIdx int) uint64 {
	return (blockNum << 32) | uint64(txIdx)
}

func DecodeTxUID(txUID uint64) (blockNum uint64, txIdx int) {
	return txUID >> 32, int(txUID & 0xffffffff)
}
