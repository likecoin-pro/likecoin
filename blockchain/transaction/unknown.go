package transaction

import (
	"fmt"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/blockchain/state"
)

type UnknownTransaction struct {
	Header
	RawData []byte `json:"raw_data"`
}

func (tx *UnknownTransaction) DecodeTransaction() (t Transaction, err error) {
	t, err = newTxByType(tx.Type)
	if err != nil {
		return tx, nil
	}
	err = t.Decode(tx.RawData)
	return
}

func (tx *UnknownTransaction) Encode() []byte {
	return tx.RawData
}

func (tx *UnknownTransaction) Decode(data []byte) error {
	if err := bin.Decode(data, &tx.Header); err != nil {
		return err
	}
	tx.RawData = data
	return nil
}

func (tx *UnknownTransaction) Verify() error {
	return nil
}

func (tx *UnknownTransaction) Execute(st *state.State) {
	st.Fail(fmt.Errorf("unknown transaction type %d. Can`t be executed", tx.Type))
}
