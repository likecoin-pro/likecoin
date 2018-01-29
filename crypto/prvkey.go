package crypto

import (
	"encoding/json"
	"math/big"

	"github.com/likecoin-pro/likecoin/crypto/xhash"
	"github.com/likecoin-pro/likecoin/std/enc"
)

const PrivateKeyVersion = '\x01'

type PrivateKey struct {
	d   *big.Int
	pub *PublicKey
}

func NewPrivateKey() *PrivateKey {
	return generateKey(randInt())
}

func MustParsePrivateKey(prvKey64 string) (prv *PrivateKey) {
	prv, err := ParsePrivateKey(prvKey64)
	if err != nil {
		panic(err)
	}
	return
}

func ParsePrivateKey(prvKey64 string) (prv *PrivateKey, err error) {
	bb, err := enc.Base64Decode(prvKey64)
	if err == nil {
		prv = DecodePrivateKey(bb)
	}
	return
}

func NewPrivateKeyBySecret(secret string) *PrivateKey {
	key := xhash.GenerateKeyByPassword(secret, KeySize*8)
	return generateKey(normInt(key))
}

func DecodePrivateKey(b []byte) *PrivateKey {
	if len(b) < 1 || b[0] != PrivateKeyVersion {
		return nil
	}
	return generateKey(new(big.Int).SetBytes(b[1:]))
}

func (prv *PrivateKey) SubKey(subKeyName string) *PrivateKey {
	d := intToBytes(prv.d)
	secret := []byte{}
	secret = append(secret, d...)
	secret = append(secret, []byte(subKeyName)...)
	secret = append(secret, d...)
	return generateKey(hashInt(secret))
}

func (prv *PrivateKey) String() string {
	return enc.Base64Encode(prv.Encode())
}

func (prv *PrivateKey) Encode() []byte {
	buf := []byte{PrivateKeyVersion} // head
	return append(buf, intToBytes(prv.d)...)
}

func (prv *PrivateKey) PublicKey() *PublicKey {
	return prv.pub
}

// Sign signs a data using the private key, prv. It returns the signature as a
// pair of integers. The security of the private key depends on the entropy of
// rand.
func (prv *PrivateKey) Sign(data []byte) []byte {
	var k, s, r *big.Int
	e := hashInt(data)
	for {
		for {
			k = randInt()
			r, _ = curve.ScalarBaseMult(k.Bytes())
			r.Mod(r, curveParams.N)
			if r.Sign() != 0 {
				break
			}
		}
		s = new(big.Int).Mul(prv.d, r)
		s.Add(s, e)
		s.Mul(s, fermatInverse(k, curveParams.N))
		s.Mod(s, curveParams.N)
		if s.Sign() != 0 {
			break
		}
	}
	return append(intToBytes(r), intToBytes(s)...)
}

func (prv *PrivateKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(prv.d)
}

func (prv *PrivateKey) UnmarshalJSON(data []byte) (err error) {
	prv.d = new(big.Int)
	if err = json.Unmarshal(data, prv.d); err == nil {
		prv.generatePub()
	}
	return
}

func (prv *PrivateKey) generatePub() {
	prv.pub = new(PublicKey)
	prv.pub.x, prv.pub.y = curve.ScalarBaseMult(prv.d.Bytes())
}

// generateKey generates a public and private key pair.
func generateKey(k *big.Int) *PrivateKey {
	prv := new(PrivateKey)
	prv.d = k
	prv.generatePub()
	return prv
}
