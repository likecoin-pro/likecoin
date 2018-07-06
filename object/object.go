package object

import "errors"

const (
	TxTypeEmission = 0
	TxTypeTransfer = 1
	TxTypeUser     = 2
)

var (
	ErrTxIncorrectAmount    = errors.New("tx-Error: Incorrect amount")
	ErrTxIncorrectIssuer    = errors.New("tx-Error: Incorrect issuer")
	ErrTxIncorrectAssetType = errors.New("tx-Error: Incorrect asset type")

	ErrInvalidUserID = errors.New("invalid userID")
)
