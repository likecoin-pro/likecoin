package transaction

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/commons/hex"
)

type Transaction interface {
	GetHeader() *Header
	Hash() []byte
	Encode() []byte
	Decode([]byte) error
	Verify() error
	Execute(*state.State)
}

type Type = uint8

var (
	ErrUnsupportedTxType   = errors.New("unsupported transaction-type")
	ErrInvalidTxData       = errors.New("invalid tx data")
	errTxTypeHasRegistered = errors.New("transaction type has been registered")
	errUnknownTxIDFormat   = errors.New("unknown txID format")

	reTxID = regexp.MustCompile(`^(?:0x)?([0-9a-f]{16})(?:[0-9a-f]{48})?$`)
)

var (
	txTypes     = map[Type]reflect.Type{}
	txTypeNames = map[Type]string{}
)

func Register(txType Type, tx Transaction) error {
	if _, ok := txTypes[txType]; ok {
		panic(errTxTypeHasRegistered)
	}
	typ := reflect.TypeOf(tx)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	txTypes[txType] = typ
	txTypeNames[txType] = typ.Name()
	return nil
}

func newTxByType(typ Type) (Transaction, error) {
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
	return hex.EncodeUint(TxID(tx))
}

func ParseTxID(s string) (uint64, error) {
	if ss := reTxID.FindStringSubmatch(s); len(ss) > 0 {
		return strconv.ParseUint(ss[1], 16, 64)
	}
	return 0, errUnknownTxIDFormat
}

func TxIDByHash(txHash []byte) uint64 {
	return bin.BytesToUint64(txHash[:8])
}

func Decode(data []byte) (tx Transaction, err error) {
	var tx0 = new(Unknown)
	if err = tx0.Decode(data); err != nil {
		return
	}
	if tx, _ = newTxByType(tx0.Header.Type); tx != nil {
		err = tx.Decode(data)
		tx.GetHeader().Sender = tx0.Sender
	} else {
		tx = tx0
	}
	return
}
