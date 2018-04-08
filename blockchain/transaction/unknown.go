package transaction

import (
	"fmt"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/likecoin-pro/likecoin/crypto/merkle"
)

// Unknown transaction
type Unknown struct {
	Header
	RawData []byte `json:"raw_data"`
}

func (tx *Unknown) Encode() []byte {
	buf := bin.NewBuffer(nil)
	buf.WriteVar(tx.Header)
	buf.Write(tx.RawData)
	return buf.Bytes()
}

func (tx *Unknown) Decode(data []byte) (err error) {
	buf := bin.NewBuffer(data)
	if err = buf.ReadVar(&tx.Header); err != nil {
		return
	}
	tx.RawData = data[buf.CntRead:]
	err = tx.recoverSender(tx)
	return
}

func (tx *Unknown) Hash() []byte {
	return merkle.Root(
		tx.HeaderHash(),
		tx.dataHash(),
	)
}

func (tx *Unknown) dataHash() []byte {
	return crypto.Hash256Raw(tx.RawData)
}

func (tx *Unknown) Verify() error {
	return tx.Header.VerifySign(tx)
}

func (tx *Unknown) Execute(st *state.State) {
	st.Fail(fmt.Errorf("transaction> unknown tx-type: %d. Can`t be executed", tx.Type))
}
