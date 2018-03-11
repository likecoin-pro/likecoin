package object

import (
	"strconv"
	"strings"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/blockchain/transaction"
	"github.com/likecoin-pro/likecoin/commons/hex"
	"github.com/likecoin-pro/likecoin/crypto"
)

type User struct {
	transaction.Header
	Nick       string            `json:"nick"`
	PubKey     *crypto.PublicKey `json:"pubkey"`
	ReferrerID hex.Uint64        `json:"referrer"`
	Data       []byte            `json:"data"`
	Sign       bin.Bytes         `json:"signature"`
}

var _ = transaction.Register(TxTypeUser, &User{})

func NewUser(
	prv *crypto.PrivateKey,
	nick string,
	referrerID uint64,
	data []byte,
) (t *User) {
	t = &User{
		Header:     transaction.NewHeader(TxTypeUser, 0),
		Nick:       nick,
		PubKey:     prv.PublicKey,
		ReferrerID: hex.Uint64(referrerID),
		Data:       data,
	}
	t.SetSign(prv)
	return
}

func (tx *User) ID() uint64 {
	return tx.PubKey.ID()
}

func (tx *User) Address() crypto.Address {
	return tx.PubKey.Address()
}

func (tx *User) hash() []byte {
	return crypto.Hash256(
		tx.Header,
		tx.PubKey,
		tx.Nick,
		tx.ReferrerID,
		tx.Data,
	)
}

func (tx *User) SetSign(prv *crypto.PrivateKey) {
	tx.PubKey = prv.PublicKey
	tx.Sign = prv.Sign(tx.hash())
}

func (tx *User) Encode() []byte {
	return bin.Encode(
		tx.Header,
		tx.PubKey,
		tx.Nick,
		tx.ReferrerID,
		tx.Data,
		tx.Sign,
	)
}

func (tx *User) Decode(data []byte) error {
	return bin.Decode(data,
		&tx.Header,
		&tx.PubKey,
		&tx.Nick,
		&tx.ReferrerID,
		&tx.Data,
		&tx.Sign,
	)
}

func (tx *User) Verify() error {
	if !tx.PubKey.Verify(tx.hash(), tx.Sign) {
		return ErrTxIncorrectSign
	}
	return nil
}

func (tx *User) Execute(st *state.State) {
	nameAsset := assets.NewName(tx.Nick)
	userAddr := tx.Address()

	// set username as asset to user-address
	st.Set(nameAsset, userAddr, state.Int(1), 0)
}

func ParseUserID(s string) (userID uint64, err error) {
	if len(s) <= 18 { // "0xFFFFFFFFFFFFFFFF" | "FFFFFFFFFFFFFFFF"
		s = strings.TrimPrefix(s, "0x")
		return strconv.ParseUint(s, 16, 64)
	}
	// todo: parse address
	// todo: parse public key
	return 0, ErrInvalidUserID
}
