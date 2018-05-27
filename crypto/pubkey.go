package crypto

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/likecoin-pro/likecoin/crypto/base58"
)

type PublicKey struct {
	x *big.Int
	y *big.Int
}

var (
	errInvalidHashSize = errors.New("crypto: invalid hash length")
	errPublicKeyDecode = errors.New("crypto: incorrect public key")
)

func (pub *PublicKey) Empty() bool {
	return pub == nil || pub.x == nil && pub.y == nil
}

func (pub *PublicKey) String() string {
	return base58.Encode(pub.Encode())
}

func (pub *PublicKey) Hex() string {
	return "0x" + hex.EncodeToString(pub.bytes()[1:])
}

func (pub *PublicKey) Equal(p *PublicKey) bool {
	return pub != nil && p != nil && pub.x.Cmp(p.x) == 0 && pub.y.Cmp(p.y) == 0
}

func (pub *PublicKey) Address() Address {

	// address := last 24 bytes of SHAKE512(SHAKE512(x||y))
	buf := make([]byte, 64)
	h := sha3.NewShake256()
	h.Write(intToBytes(pub.x))
	h.Write(intToBytes(pub.y))
	h.Read(buf)

	h = sha3.NewShake256()
	h.Write(buf)
	h.Read(buf)

	return newAddress(buf[64-AddressLength:])
}

func (pub *PublicKey) ID() uint64 {
	return pub.Address().ID()
}

func (pub *PublicKey) bytes() []byte {
	x, y := pub.x.Bytes(), pub.y.Bytes()

	ret := make([]byte, 1+2*KeySize)
	ret[0] = 4 // uncompressed point
	copy(ret[1+KeySize-len(x):], x)
	copy(ret[1+2*KeySize-len(y):], y)
	return ret
}

func (pub *PublicKey) Encode() []byte {
	// compressed key
	return secp256k1.CompressPubkey(pub.x, pub.y)
}

func (pub *PublicKey) Decode(data []byte) error {
	pub.x, pub.y = secp256k1.DecompressPubkey(data)
	if pub.x == nil || pub.y == nil {
		return errPublicKeyDecode
	}
	return nil
}

func (pub *PublicKey) MarshalJSON() ([]byte, error) {
	return []byte(`"` + pub.String() + `"`), nil
}

func (pub *PublicKey) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	if p, err := ParsePublicKey(str); err != nil {
		return err
	} else {
		pub.x = p.x
		pub.y = p.y
		return nil
	}
}

func (pub *PublicKey) Verify(hash []byte, sig []byte) bool {
	if pub.Empty() || len(sig) != signatureSize || len(hash) != KeySize {
		return false
	}
	return secp256k1.VerifySignature(pub.bytes(), hash, sig[:64])
}

func RecoverPublicKey(hash, sig []byte) (*PublicKey, error) {
	if len(hash) != KeySize {
		panic(errInvalidHashSize)
	}
	pub, err := secp256k1.RecoverPubkey(hash, sig)
	if err != nil {
		return nil, err
	}
	if len(pub) != 1+2*KeySize || pub[0] != 4 {
		return nil, errPublicKeyDecode
	}
	x := new(big.Int).SetBytes(pub[1 : 1+KeySize])
	y := new(big.Int).SetBytes(pub[1+KeySize:])
	if x.Cmp(curveP) >= 0 || y.Cmp(curveP) >= 0 || !curve.IsOnCurve(x, y) {
		return nil, errPublicKeyDecode
	}
	return &PublicKey{x, y}, nil
}

func MustParsePublicKey(pubkey string) *PublicKey {
	if pub, err := ParsePublicKey(pubkey); err != nil {
		panic(err)
	} else {
		return pub
	}
}

func ParsePublicKey(str64 string) (pub *PublicKey, err error) {
	data, err := base58.DecodeFixed(str64, publicKeySize)
	if err != nil {
		return
	}
	return DecodePublicKey(data)
}

func DecodePublicKey(data []byte) (pub *PublicKey, err error) {
	pub = &PublicKey{}
	err = pub.Decode(data)
	return
}
