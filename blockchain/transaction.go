package blockchain

import (
	"errors"
	"reflect"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/commons/enc"
)

type Transaction interface {
	Type() TxType
	Hash() []byte
	bin.Encoder
	bin.Decoder
	Execute(*state.State)
}

type TxType uint8

var (
	ErrUnsupportedTxType   = errors.New("unsupported transaction-type")
	ErrInvalidTxData       = errors.New("invalid tx data")
	errTxTypeHasRegistered = errors.New("transaction type has been registered")
)

var txTypes = map[TxType]reflect.Type{}

func RegisterTransactionType(tx Transaction) error {
	if _, ok := txTypes[tx.Type()]; ok {
		panic(errTxTypeHasRegistered)
	}
	typ := reflect.TypeOf(tx)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	txTypes[tx.Type()] = typ
	return nil
}

func newTransaction(typ TxType) (Transaction, error) {
	rt, ok := txTypes[typ]
	if !ok {
		return nil, ErrUnsupportedTxType
	}
	ptr := reflect.New(rt)
	if obj, ok := ptr.Interface().(Transaction); ok {
		return obj, nil
	} else {
		return nil, ErrUnsupportedTxType
	}
}

func TxID(tx Transaction) uint64 {
	return TxIDByHash(tx.Hash())
}

func StrTxID(tx Transaction) string {
	return enc.UintToHex(TxID(tx))
}

func TxIDByHash(txHash []byte) uint64 {
	return bin.BytesToUint64(txHash[:8])
}

func TxEncode(tx Transaction) []byte {
	return append([]byte{byte(tx.Type())}, tx.Encode()...)
}

func TxDecode(data []byte) (tx Transaction, err error) {
	if len(data) < 2 {
		return nil, ErrInvalidTxData
	}
	txType := TxType(data[0])
	if tx, err = newTransaction(txType); err == nil {
		err = tx.Decode(data[1:])
	}
	return
}
