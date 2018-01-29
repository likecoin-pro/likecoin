package crypto

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/likecoin-pro/likecoin/std/enc"
	"golang.org/x/crypto/sha3"
)

type PublicKey struct {
	x *big.Int
	y *big.Int
}

var (
	errPublicKeyDecode = errors.New("crypto.PublicKey.Decode error: incorrect length of key")
)

//func (pub *PublicKey) VerifyPoW(data, sign []byte, difficulty uint64) bool {
//	p := pow.NewPoW(nil, difficulty)
//	return p.CheckHashDifficulty(hash256(sign)) && pub.Verify(data, sign)
//}

// Verify verifies the signature in r, s of hash using the public key, pub. Its
// return value records whether the signature is valid.
func (pub *PublicKey) Verify(data []byte, sign []byte) bool {
	if pub.Empty() {
		return false
	}
	if len(sign) != PublicKeySize {
		return false
	}
	r := new(big.Int).SetBytes(sign[:KeySize])
	s := new(big.Int).SetBytes(sign[KeySize:])

	if r.Sign() == 0 || r.Cmp(curveParams.N) >= 0 {
		return false
	}
	if s.Sign() == 0 || s.Cmp(curveParams.N) >= 0 {
		return false
	}

	e := hashInt(data)
	w := new(big.Int).ModInverse(s, curveParams.N)

	u1 := e.Mul(e, w)
	u2 := w.Mul(r, w)

	u1.Mod(u1, curveParams.N)
	u2.Mod(u2, curveParams.N)

	x1, y1 := curve.ScalarBaseMult(u1.Bytes())
	x2, y2 := curve.ScalarMult(pub.x, pub.y, u2.Bytes())
	x, y := curve.Add(x1, y1, x2, y2)
	if x.Sign() == 0 && y.Sign() == 0 {
		return false
	}
	x.Mod(x, curveParams.N)
	return x.Cmp(r) == 0
}

func (pub *PublicKey) Empty() bool {
	return pub == nil || pub.x == nil && pub.y == nil
}

func (pub *PublicKey) String() string {
	return enc.Base64Encode(pub.Encode())
}

//func (pub *PublicKey) ID() uint64 {
//	return pub.Address().ID()
//}

func (pub *PublicKey) Is(p *PublicKey) bool {
	return pub != nil && p != nil && pub.x.Cmp(p.x) == 0 && pub.y.Cmp(p.y) == 0
}

//func (pub *PublicKey) StrAddress() string {
//	return EncodeAddress(pub.Address())
//}

func (pub *PublicKey) Address() Address {
	hash256 := newHash256()
	hash256.Write(intToBytes(pub.x))
	hash256.Write(intToBytes(pub.y))
	h := sha3.Sum512(hash256.Sum(nil))
	return newAddress(h[:AddressSize])
}

//func AddressToUserID(addr160 []byte) uint64 {
//	if IsValidAddress(addr160) {
//		return bin.BytesToUint64(addr160[:8])
//	}
//	return 0
//}

func (pub *PublicKey) Encode() []byte {
	return append(
		intToBytes(pub.x),    // 32 bytes X
		intToBytes(pub.y)..., // 32 bytes Y
	)
}

func (pub *PublicKey) Decode(data []byte) error {
	if len(data) != PublicKeySize {
		return errPublicKeyDecode
	}
	pub.x = new(big.Int).SetBytes(data[:KeySize])
	pub.y = new(big.Int).SetBytes(data[KeySize:])
	return nil
}

func (pub *PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(enc.Base64Encode(pub.Encode()))
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

func MustParsePublicKey(pubkey string) *PublicKey {
	if pub, err := ParsePublicKey(pubkey); err != nil {
		panic(err)
	} else {
		return pub
	}
}

func ParsePublicKey(str64 string) (pub *PublicKey, err error) {
	data, err := enc.Base64Decode(str64)
	if err != nil {
		return
	}
	if len(data) != PublicKeySize {
		err = errors.New("Invalid public key")
		return
	}
	pub, err = DecodePublicKey(data)
	if err != nil {
		return nil, err
	}
	return
}

func DecodePublicKey(data []byte) (pub *PublicKey, err error) {
	pub = &PublicKey{}
	err = pub.Decode(data)
	return
}
