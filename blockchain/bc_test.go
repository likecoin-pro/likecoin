package blockchain

import (
	"errors"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/commons/bignum"
)

var TestCounter = assets.Asset{1}

type TestTxObject struct {
	Transaction
	Msg string
}

var _ = RegisterTxObject(127, &TestTxObject{})

func (obj *TestTxObject) Encode() []byte {
	return bin.Encode(
		obj.Msg,
	)
}

func (obj *TestTxObject) Decode(data []byte) error {
	return bin.Decode(data,
		&obj.Msg,
	)
}

func (obj *TestTxObject) Verify(tx *Transaction) error {
	if len(obj.Msg) > 100 {
		return errors.New("tx.msg is too long")
	}
	return nil
}

func (obj *TestTxObject) Execute(tx *Transaction, st *state.State) {
	st.Increment(TestCounter, obj.Sender.Address(), bignum.NewInt(int64(len(obj.Msg))), 0)
}
