package crypto

import (
	"encoding/hex"
	"math/big"

	"strings"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/likecoin-pro/likecoin/crypto/xhash"
)

const PrivateKeyVersion = '\x01'

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

func NewPrivateKeyBySecret(secret string) *PrivateKey {
	key := xhash.GenerateKeyByPassword(secret, KeySize*8)
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
	return newPrvKey(new(big.Int).SetBytes(data[1:])), nil
}

func MustParsePrivateKey(hexKey string) *PrivateKey {
	prv, err := ParsePrivateKey(hexKey)
	if err != nil {
		panic(err)
	}
	return prv
}
