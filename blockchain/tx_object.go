package blockchain

import (
	"errors"
	"reflect"

	"github.com/likecoin-pro/likecoin/blockchain/state"
)

type TxObject interface {
	Encode() []byte
	Decode([]byte) error
	Verify(*Transaction) error
	Execute(*Transaction, *state.State)
}

var (
	txTypes     = map[Type]reflect.Type{}
	txTypeNames = map[Type]string{}
	txObjTypes  = map[reflect.Type]Type{}
)

var (
	ErrUnsupportedTxType   = errors.New("unsupported transaction-type")
	errTxTypeHasRegistered = errors.New("transaction type has been registered")
)

func RegisterTxObject(txType Type, txObj TxObject) error {
	if _, ok := txTypes[txType]; ok {
		panic(errTxTypeHasRegistered)
	}
	typ := reflect.TypeOf(txObj)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	txTypes[txType] = typ
	txTypeNames[txType] = typ.Name()
	txObjTypes[typ] = txType
	return nil
}

func typeByObject(obj TxObject) Type {
	typ := reflect.TypeOf(obj)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return txObjTypes[typ]
}

func newObjectByType(typ Type) (TxObject, error) {
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
