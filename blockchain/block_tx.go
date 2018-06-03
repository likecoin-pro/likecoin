package blockchain

import (
	"bytes"
	"encoding/json"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/likecoin-pro/likecoin/commons/hex"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/likecoin-pro/likecoin/crypto/merkle"
)

type BlockTx struct {
	Tx           *Transaction
	StateUpdates state.Values
	Sig          []byte

	// not imported
	TxSeq   uint64 // <blockNum:8byte><blockTxIdx:8byte>
	BlockTs int64  //
}

func NewBlockTx(sender *crypto.PrivateKey, obj TxObject) *BlockTx {
	tx := &Transaction{
		Type:    typeByObject(obj),   //
		Version: 0,                   //
		Network: config.NetworkID,    //
		ChainID: config.ChainID,      //
		Sender:  sender.PublicKey,    //
		Nonce:   uint64(timestamp()), //
		Data:    obj.Encode(),        // encoded tx-object
	}
	return &BlockTx{
		Tx:  tx,
		Sig: sender.Sign(tx.Hash()),
	}
}

func (btx *BlockTx) String() string {
	return enc.JSON(btx)
}

func (btx *BlockTx) BlockNum() uint64 {
	return uint64(btx.TxSeq) >> 32
}

func (btx *BlockTx) Hash() []byte {
	return merkle.Root(btx.Tx.Hash(), btx.StateUpdates.Hash())
}

func (btx *BlockTx) TxObject() TxObject {
	obj, _ := btx.Tx.Object()
	return obj
}

func (btx *BlockTx) TxType() TxType {
	return btx.Tx.Type
}

func (btx *BlockTx) TxAddress() crypto.Address {
	return btx.Tx.Sender.Address()
}

func (btx *BlockTx) TxHash() []byte {
	return btx.Tx.Hash()
}

func (btx *BlockTx) Equal(btx1 *BlockTx) bool {
	return bytes.Equal(btx.Encode(), btx1.Encode())
}

func (btx *BlockTx) Encode() []byte {
	return bin.Encode(
		btx.Tx,
		btx.StateUpdates,
		btx.Sig,
	)
}

func (btx *BlockTx) Decode(data []byte) error {
	return bin.Decode(data,
		&btx.Tx,
		&btx.StateUpdates,
		&btx.Sig,
	)
}

func (btx *BlockTx) Verify() error {

	//-- verify transaction data
	txObj, err := btx.Tx.Object()
	if err != nil {
		return err
	}
	if err := txObj.Verify(btx.Tx); err != nil {
		return err
	}

	//-- verify sender signature
	if !btx.Tx.Sender.Verify(btx.Tx.Hash(), btx.Sig) {
		return ErrInvalidSign
	}
	return nil
}

type blockTxJSON struct {
	Tx interface{} `json:"tx"` //
	//TxObj        TxObject     `json:"tx_obj"`    //
	TxID         hex.Uint64   `json:"tx_id"`     //
	TxHash       hex.Bytes    `json:"tx_hash"`   //
	Sig          hex.Bytes    `json:"tx_sig"`    //
	BlockNum     int          `json:"block_num"` //
	BlockIdx     int          `json:"block_idx"` //
	BlockTs      int64        `json:"block_ts"`  //
	StateUpdates state.Values `json:"state"`     //
}

func (btx *BlockTx) MarshalJSON() ([]byte, error) {
	if btx == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(&blockTxJSON{
		Tx: btx.Tx,
		//TxObj:        btx.TxObject(),
		TxID:         hex.Uint64(btx.Tx.ID()),
		TxHash:       hex.Bytes(btx.Tx.Hash()),
		Sig:          hex.Bytes(btx.Sig),
		BlockNum:     int(btx.TxSeq >> 32),
		BlockIdx:     int(btx.TxSeq & 0xffffffff),
		BlockTs:      btx.BlockTs,
		StateUpdates: btx.StateUpdates,
	})
}
