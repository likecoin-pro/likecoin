package crypto

import (
	"math/big"

	"github.com/likecoin-pro/likecoin/crypto/base58"
	"github.com/likecoin-pro/likecoin/crypto/xhash"
)

const PrivateKeyVersion = '\x01'

type PrivateKey struct {
	d         *big.Int
	PublicKey *PublicKey
}

// newPrvKey generates a public and private key pair.
func newPrvKey(d *big.Int) *PrivateKey {
	x, y := curve.ScalarBaseMult(d.Bytes())
	return &PrivateKey{
		d:         d,
		PublicKey: &PublicKey{x, y},
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
	return base58.Encode(prv.Encode())
}

func (prv *PrivateKey) Encode() []byte {
	buf := []byte{PrivateKeyVersion} // head
	return append(buf, intToBytes(prv.d)...)
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
