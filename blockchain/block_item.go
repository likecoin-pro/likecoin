package blockchain

import (
	"encoding/json"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/blockchain/transaction"
	"github.com/likecoin-pro/likecoin/crypto"
)

type BlockItem struct {
	Tx    transaction.Transaction //
	State *state.State            // state changes

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

func (it *BlockItem) BlockNum() uint64 {
	return it.block.Num
}

func (it *BlockItem) Timestamp() int64 {
	return it.block.Timestamp
}

func (it *BlockItem) Hash() []byte {
	return crypto.Hash256(
		it.TxHash(),
		it.State.Hash(),
	)
}

func (it *BlockItem) Encode() []byte {
	return bin.Encode(
		it.Tx,
		it.State,
	)
}

func (it *BlockItem) Decode(data []byte) (err error) {
	tx0 := new(transaction.UnknownTransaction)
	err = bin.Decode(data,
		&tx0,
		&it.State,
	)
	if err != nil {
		return
	}
	it.Tx, err = tx0.DecodeTransaction()
	return
}

func (it *BlockItem) TxHeader() transaction.Header {
	return it.Tx.GetHeader()
}

func (it *BlockItem) TxType() transaction.Type {
	return it.Tx.GetHeader().Type
}

func (it *BlockItem) StrTxType() string {
	return it.Tx.GetHeader().StrType()
}

func (it *BlockItem) StrTxID() string {
	return transaction.StrTxID(it.Tx)
}

func (it *BlockItem) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		BlockNum  uint64                  `json:"block"`    //
		BlockTs   int64                   `json:"block_ts"` //
		TxStrType string                  `json:"tx_type"`  //
		TxID      string                  `json:"tx_id"`    //
		TxHash    bin.Bytes               `json:"tx_hash"`  //
		Tx        transaction.Transaction `json:"tx"`       // transaction object
		State     *state.State            `json:"state"`    // new state values
	}{
		BlockNum:  it.block.Num,
		BlockTs:   it.block.Timestamp,
		TxStrType: it.TxHeader().StrType(),
		TxID:      it.StrTxID(),
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
