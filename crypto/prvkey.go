package crypto

import (
	"encoding/hex"
	"errors"
	"math/big"
	"strings"

	"github.com/likecoin-pro/likecoin/crypto/secp256k1"
	"github.com/likecoin-pro/likecoin/crypto/xhash"
)

const PrivateKeyVersion = '\x01'

var errPrvUnknownFormat = errors.New("crypto> unknown private key format")

type PrivateKey struct {
	d *big.Int

	PublicKey *PublicKey
}

// newPrvKey generates a public and private key pair.
func newPrvKey(d *big.Int) *PrivateKey {
	x, y := curve.ScalarBaseMult(d.Bytes())
	return &PrivateKey{
		d:         d,
		PublicKey: &PublicKey{x: x, y: y},
	}
}

func NewPrivateKey() *PrivateKey {
	return newPrvKey(randInt())
}

func NewPrivateKeyBySecret(seed string) *PrivateKey {
	key := xhash.GenerateKeyByPassword(seed, KeySize*8)
	return newPrvKey(normInt(key))
}

func (prv *PrivateKey) String() string {
	return prv.Hex()
}

func (prv *PrivateKey) Hex() string {
	return "0x" + hex.EncodeToString(prv.Encode())
}

func (prv *PrivateKey) Encode() []byte {
	buf := make([]byte, 1+KeySize)
	buf[0] = PrivateKeyVersion // head
	copy(buf[1:], intToBytes(prv.d))
	return buf
}

func (prv *PrivateKey) Sign(hash []byte) []byte {
	if len(hash) != KeySize {
		panic(errInvalidHashSize)
	}
	sig, err := secp256k1.Sign(hash, intToBytes(prv.d))
	if err != nil {
		panic(err)
	}
	return sig
}

func ParsePrivateKey(hexKey string) (prv *PrivateKey, err error) {
	hexKey = strings.TrimPrefix(hexKey, "0x")
	data, err := hex.DecodeString(hexKey)
	if err != nil {
		return
	}
	if len(data) == 33 {
		if data[0] != PrivateKeyVersion {
			return nil, errPrvUnknownFormat
		}
		data = data[1:]
	}
	if len(data) != 32 {
		return nil, errPrvUnknownFormat
	}
	return newPrvKey(new(big.Int).SetBytes(data)), nil
}

func MustParsePrivateKey(hexKey string) *PrivateKey {
	prv, err := ParsePrivateKey(hexKey)
	if err != nil {
		panic(err)
	}
	return prv
}
