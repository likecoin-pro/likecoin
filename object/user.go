package object

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/commons/bignum"
	"github.com/likecoin-pro/likecoin/commons/hex"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
)

type User struct {
	Nick       string     `json:"nick"`
	ReferrerID hex.Uint64 `json:"referrer"`
	Data       []byte     `json:"data"`
}

var _ = blockchain.RegisterTxObject(TxTypeUser, &User{})

func NewUser(
	cfg *blockchain.Config,
	from *crypto.PrivateKey,
	nick string,
	referrerID uint64,
	data []byte,
) *blockchain.Transaction {
	return blockchain.NewTx(cfg, from, 0, &User{
		Nick:       nick,
		ReferrerID: hex.Uint64(referrerID),
		Data:       data,
	})
}

func (obj *User) String() string {
	return obj.Nick
}

func (obj *User) Encode() []byte {
	return bin.Encode(
		0, //ver
		obj.Nick,
		obj.ReferrerID,
		obj.Data,
	)
}

func (obj *User) Decode(data []byte) error {
	return bin.Decode(data,
		new(int),
		&obj.Nick,
		&obj.ReferrerID,
		&obj.Data,
	)
}

var (
	reNick = regexp.MustCompile(`^[a-z][a-z0-9\-]{2,20}$`)

	errInvalidNickname   = errors.New("tx-user-verify: incorrect nickname")
	errUserDataIsTooLong = errors.New("tx-user-verify: data is too long")
)

func (obj *User) Verify(tx *blockchain.Transaction) error {
	if !reNick.MatchString(obj.Nick) {
		return errInvalidNickname
	}
	if len(obj.Data) > config.TxUserDataSizeLimit {
		return errUserDataIsTooLong
	}
	return nil
}

func (obj *User) Execute(tx *blockchain.Transaction, st *state.State) {
	nameAsset := assets.NewName(obj.Nick)
	userAddr := tx.SenderAddress()

	// set username as asset to user-address
	st.Set(nameAsset, userAddr, bignum.NewInt(1), 0)
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
