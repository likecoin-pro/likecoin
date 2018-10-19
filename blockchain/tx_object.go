package blockchain

import (
	"errors"
	"reflect"

	"github.com/likecoin-pro/likecoin/blockchain/state"
)

type TxObject interface {
	Encode() []byte
	Decode([]byte) error
	SetContext(*Transaction)
	Verify() error
	Execute(*state.State)
}

var (
	txTypes    = map[TxType]reflect.Type{}
	txTypeStr  = map[TxType]string{}
	txObjTypes = map[reflect.Type]TxType{}
)

var (
	ErrUnsupportedTxType   = errors.New("unsupported transaction-type")
	errTxTypeHasRegistered = errors.New("transaction type has been registered")
)

func RegisterTxObject(txType TxType, txObj TxObject) error {
	if _, ok := txTypes[txType]; ok {
		panic(errTxTypeHasRegistered)
	}
	typ := reflect.TypeOf(txObj)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	txTypes[txType] = typ
	txTypeStr[txType] = typ.Name()
	txObjTypes[typ] = txType
	return nil
}

func TxTypeStr(typ TxType) string {
	return txTypeStr[typ]
}

func typeByObject(obj TxObject) TxType {
	typ := reflect.TypeOf(obj)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return txObjTypes[typ]
}

func newObjectByType(typ TxType) (TxObject, error) {
	rt, ok := txTypes[typ]
	if !ok {
		return nil, ErrUnsupportedTxType
	}
	ptr := reflect.New(rt)
	if obj, ok := ptr.Interface().(TxObject); ok {
		return obj, nil
	} else {
		return nil, ErrUnsupportedTxType
	}
}
