package blockchain

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"time"

	"bytes"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/likecoin-pro/likecoin/commons/hex"
	"github.com/likecoin-pro/likecoin/crypto"
)

type TxType = uint8

type Transaction struct {
	Type    TxType            // tx type
	Version int               // tx version
	Network int               //
	ChainID uint64            //
	Nonce   uint64            // sender nonce
	Sender  *crypto.PublicKey // tx sender
	Data    []byte            // encoded tx-object

	// not imported fields
	_obj TxObject //
}

type transactionJSON struct {
	Type    TxType            `json:"type"`    // tx type
	Version int               `json:"version"` // tx version
	Network int               `json:"network"` //
	ChainID uint64            `json:"chain"`   //
	Nonce   uint64            `json:"nonce"`   //
	Sender  *crypto.PublicKey `json:"sender"`  // tx sender
	ObjRaw  hex.Bytes         `json:"data"`    // encoded tx-data
	Obj     TxObject          `json:"obj"`     // unserialized data
}

var (
	ErrInvalidTxData     = errors.New("invalid tx data")
	errUnknownTxIDFormat = errors.New("unknown txID format")
	errEmptyTxSender     = errors.New("tx-error: empty tx-sender")
	errEmptyTxData       = errors.New("tx-error: empty tx-data")

	reTxID = regexp.MustCompile(`^(?:0x)?([0-9a-f]{16})(?:[0-9a-f]{48})?$`)
)

func (tx *Transaction) String() string {
	if obj, err := tx.Object(); err == nil {
		return enc.IndentJSON(obj)
	}
	return enc.IndentJSON(tx)
}

func (tx *Transaction) ID() uint64 {
	return TxIDByHash(tx.Hash())
}

func (tx *Transaction) StrID() string {
	return hex.EncodeUint(tx.ID())
}

func (tx *Transaction) Address() crypto.Address {
	return tx.Sender.Address()
}

func (tx *Transaction) Hash() []byte {
	return crypto.Hash256(
		tx.Type,
		tx.Version,
		tx.Network,
		tx.ChainID,
		tx.Nonce,
		tx.Sender,
		tx.Data,
	)
}

func (tx *Transaction) Size() int {
	return len(tx.Encode())
}

func (tx *Transaction) StrType() string {
	return TxTypeStr(tx.Type)
}

func (tx *Transaction) Equal(tx1 *Transaction) bool {
	return bytes.Equal(tx.Encode(), tx1.Encode())
}

func (tx *Transaction) Encode() []byte {
	if len(tx.Data) == 0 {
		panic(errEmptyTxData)
	}
	return bin.Encode(
		tx.Type,
		tx.Version,
		tx.Network,
		tx.ChainID,
		tx.Nonce,
		tx.Sender,
		tx.Data,
	)
}

func (tx *Transaction) Decode(data []byte) (err error) {
	return bin.Decode(data,
		&tx.Type,
		&tx.Version,
		&tx.Network,
		&tx.ChainID,
		&tx.Nonce,
		&tx.Sender,
		&tx.Data,
	)
}

func (tx *Transaction) Object() (obj TxObject, err error) {
	if tx._obj != nil {
		return tx._obj, nil
	}
	obj, err = newObjectByType(tx.Type)
	if err != nil {
		return
	}
	if err = obj.Decode(tx.Data); err != nil {
		return
	}
	tx._obj = obj
	return
}

// Timestamp returns user timestamp from nonce
func (tx *Transaction) Timestamp() time.Time {
	return time.Unix(0, int64(tx.Nonce)*1e3)
}

// Execute executes tx, changes state, returns state-updates
func (tx *Transaction) Execute(s *state.State) (updates state.Values, err error) {
	defer func() {
		if r, _ := recover().(error); r != nil {
			err = r
		}
	}()

	obj, err := tx.Object()
	if err != nil {
		return
	}

	newState := s.NewSubState()

	obj.Execute(tx, newState)

	updates = newState.Values()

	return
}

func (tx *Transaction) MarshalJSON() ([]byte, error) {
	if tx == nil {
		return json.Marshal(nil)
	}
	obj, _ := tx.Object()
	return json.Marshal(&transactionJSON{
		Type:    tx.Type,
		Version: tx.Version,
		Network: tx.Network,
		ChainID: tx.ChainID,
		Nonce:   tx.Nonce,
		Sender:  tx.Sender,
		ObjRaw:  tx.Data,
		Obj:     obj,
	})
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
