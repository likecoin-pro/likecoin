package blockchain

import (
	"encoding/json"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/blockchain/transaction"
)

type BlockItem struct {
	Ts    int64                   // timestamp in µsec
	Tx    transaction.Transaction //
	State *state.State            // new state values

	// not imported fields
	block    *Block // link on parent-block
	blockIdx int    //
}

func (it *BlockItem) TxID() uint64 {
	return transaction.TxID(it.Tx)
}

func (it *BlockItem) TxHash() []byte {
	return transaction.Hash(it.Tx)
}

func (it *BlockItem) UID() uint64 {
	return EncodeTxUID(it.block.Num, it.blockIdx)
}

func (it *BlockItem) Hash() []byte {
	return bin.Hash256(
		it.Ts,
		it.TxHash(),
		it.State.Hash(),
	)
}

func (it *BlockItem) Encode() []byte {
	return bin.Encode(
		it.Ts,
		it.Tx,
		it.State,
	)
}

func (it *BlockItem) Decode(data []byte) (err error) {
	tx0 := new(transaction.UnknownTransaction)
	err = bin.Decode(data,
		&it.Ts,
		&tx0,
		&it.State,
	)
	if err != nil {
		return
	}
	it.Tx, err = tx0.DecodeTransaction()
	return
}

func (it *BlockItem) MarshalJSON() ([]byte, error) {
	h := it.Tx.GetHeader()
	return json.Marshal(struct {
		Ts        int64                   `json:"timestamp"` // timestamp in µsec
		TxType    transaction.Type        `json:"tx_type"`   //
		TxStrType string                  `json:"tx_stype"`  //
		TxID      string                  `json:"tx_id"`     //
		TxHash    bin.Bytes               `json:"tx_hash"`   //
		Tx        transaction.Transaction `json:"tx"`        // transaction object
		State     *state.State            `json:"state"`     // new state values
	}{
		Ts:        it.Ts,
		TxType:    h.Type,
		TxStrType: transaction.TypeName(h.Type),
		TxID:      transaction.StrTxID(it.Tx),
		TxHash:    it.TxHash(),
		Tx:        it.Tx,
		State:     it.State,
	})
}

func EncodeTxUID(blockNum uint64, txIdx int) uint64 {
	return (blockNum << 32) | uint64(txIdx)
}

func DecodeTxUID(txUID uint64) (blockNum uint64, txIdx int) {
	return txUID >> 32, int(txUID & 0xffffffff)
}
