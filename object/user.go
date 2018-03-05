package object

import (
	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/commons/hex"
	"github.com/likecoin-pro/likecoin/crypto"
)

type User struct {
	Version   int                    `json:"version"`
	Nick      string                 `json:"nick"`
	PubKey    *crypto.PublicKey      `json:"pubkey"`
	RefererID hex.Uint64             `json:"referer"`
	Data      map[string]interface{} `json:"data"`
	Sign      hex.Bytes              `json:"signature"`
}

var _ = blockchain.RegisterTransactionType(&User{})

func (t *User) Type() blockchain.TxType {
	return TxTypeUser
}

func NewUser(
	prv *crypto.PrivateKey,
	nick string,
	refererID uint64,
	data map[string]interface{},
) (t *User) {
	t = &User{
		Version:   0,
		Nick:      nick,
		PubKey:    prv.PublicKey,
		RefererID: hex.Uint64(refererID),
		Data:      data,
	}
	t.SetSign(prv)
	return
}

func (t *User) ID() uint64 {
	return t.PubKey.ID()
}

func (t *User) Address() crypto.Address {
	return t.PubKey.Address()
}

func (t *User) hash() []byte {
	return bin.Hash256(
		t.Version,
		t.PubKey,
		t.Nick,
		t.PubKey,
		t.RefererID,
		t.Data,
	)
}

func (t *User) SetSign(prv *crypto.PrivateKey) {
	t.PubKey = prv.PublicKey
	t.Sign = prv.Sign(t.hash())
}

func (t *User) Encode() []byte {
	return bin.Encode(
		t.Version,
		t.PubKey,
		t.Nick,
		t.PubKey,
		t.RefererID,
		t.Data,
		t.Sign,
	)
}

func (t *User) Decode(data []byte) error {
	return bin.Decode(data,
		&t.Version,
		&t.PubKey,
		&t.Nick,
		&t.PubKey,
		&t.RefererID,
		&t.Data,
		&t.Sign,
	)
}

func (t *User) Verify() bool {
	return t.PubKey.Verify(t.hash(), t.Sign)
}

func (t *User) Execute(st *state.State) {
	if !t.Verify() {
		st.Fail(ErrTxIncorrectSign)
	}

	nameAsset := assets.NewName(t.Nick)
	userAddr := t.Address()

	// increment amount to new address
	st.Increment(nameAsset, userAddr, state.Int(1), 0)
}
