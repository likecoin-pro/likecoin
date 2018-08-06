package object

import "errors"

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
