package object

import "errors"

const (
	TxTypeEmission = 0
	TxTypeTransfer = 1
	TxTypeUser     = 2
)

var (
	ErrTxIncorrectAmount    = errors.New("txExecError: Incorrect amount")
	ErrTxIncorrectIssuer    = errors.New("txExecError: Incorrect issuer")
	ErrTxIncorrectSign      = errors.New("txExecError: Incorrect signature")
	ErrTxIncorrectAssetType = errors.New("txExecError: Incorrect asset type")

	ErrInvalidUserID = errors.New("invalid userID")
)
