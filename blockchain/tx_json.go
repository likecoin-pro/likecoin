package blockchain

import (
	"encoding/json"

	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/commons/hex"
	"github.com/likecoin-pro/likecoin/crypto"
)

type transactionJSON struct {
	TxID         hex.Uint64        `json:"id"`        //
	TxHash       hex.Bytes         `json:"hash"`      //
	BlockNum     uint64            `json:"block_num"` //
	BlockIdx     int               `json:"block_idx"` //
	BlockTs      int64             `json:"block_ts"`  //
	TxSeq        string            `json:"seq"`       //
	Type         TxType            `json:"type"`      // tx type
	Version      int               `json:"version"`   // tx version
	Network      int               `json:"network"`   //
	ChainID      uint64            `json:"chain"`     //
	Nonce        uint64            `json:"nonce"`     //
	Sender       *crypto.PublicKey `json:"sender"`    // tx sender
	ObjRaw       hex.Bytes         `json:"data"`      // encoded tx-data
	Obj          TxObject          `json:"obj"`       // unserialized data
	Sig          hex.Bytes         `json:"sig"`       //
	StateUpdates state.Values      `json:"state"`     //
}

func (tx *Transaction) MarshalJSON() ([]byte, error) {
	if tx == nil {
		return json.Marshal(nil)
	}
	obj, _ := tx.Object()
	return json.Marshal(&transactionJSON{
		Type:         tx.Type,
		Version:      tx.Version,
		Network:      tx.Network,
		ChainID:      tx.ChainID,
		Nonce:        tx.Nonce,
		Sender:       tx.Sender,
		ObjRaw:       tx.Data,
		Obj:          obj,
		TxID:         hex.Uint64(tx.ID()),
		TxHash:       hex.Bytes(tx.Hash()),
		Sig:          hex.Bytes(tx.Sig),
		BlockNum:     tx.BlockNum(),
		BlockIdx:     tx.BlockIdx(),
		BlockTs:      tx.BlockTs(),
		TxSeq:        "0x" + hex.EncodeUint(tx.Seq()),
		StateUpdates: tx.StateUpdates,
	})
}
