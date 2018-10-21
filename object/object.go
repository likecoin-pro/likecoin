package object

import (
	"errors"

	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/crypto"
)

const (
	TxTypeEmission = 0
	TxTypeTransfer = 1
	TxTypeUser     = 2
)

var (
	ErrTxIncorrectAmount     = errors.New("tx-Error: Incorrect amount")
	ErrTxIncorrectSender     = errors.New("tx-Error: Incorrect sender")
	ErrTxIncorrectAssetType  = errors.New("tx-Error: Incorrect asset type")
	ErrTxIncorrectOutAddress = errors.New("tx-Error: Incorrect out address")

	ErrInvalidUserID = errors.New("invalid userID")
)

type Object struct {
	tx *blockchain.Transaction `json:"-"`
}

func (obj *Object) Sender() *crypto.PublicKey {
	if obj.tx != nil {
		return obj.tx.Sender
	}
	return nil
}
func (obj *Object) SenderAddress() crypto.Address {
	if obj.tx != nil {
		return obj.tx.SenderAddress()
	}
	return crypto.NilAddress
}

func (obj *Object) Tx() *blockchain.Transaction {
	return obj.tx
}

func (obj *Object) SetContext(tx *blockchain.Transaction) {
	obj.tx = tx
}
