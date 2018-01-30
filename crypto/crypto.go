package crypto

import (
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
)

const (
	// KeySize is Size of Private Key in bytes
	KeySize       = 256 / 8     // 32 bytes
	PublicKeySize = KeySize * 2 // 64 bytes
)

var (
	curve       = elliptic.P256()
	curveParams = curve.Params()
)

var (
	one = big.NewInt(1)
	two = big.NewInt(2)
)

// ------------------------------------
func intToBytes(i *big.Int) []byte {
	bb := i.Bytes()
	if n := len(bb); n < KeySize {
		return append(make([]byte, KeySize-n), bb...)
	}
	return bb
}

// fermatInverse calculates the inverse of k in GF(P) using Fermat's method.
// This has better constant-time properties than Euclid's method (implemented
// in math/big.Int.ModInverse) although math/big itself isn't strictly
// constant-time so it's not perfect.
func fermatInverse(k, N *big.Int) *big.Int {
	nMinus2 := new(big.Int).Sub(N, two)
	return new(big.Int).Exp(k, nMinus2, N)
}

func hashInt(data []byte) *big.Int {
	return normInt(hash256(data)[:KeySize])
}

func randBytes(n int) []byte {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return buf
}

func randInt() *big.Int {
	return normInt(randBytes(KeySize))
}

func normInt(b []byte) *big.Int {
	k := new(big.Int).SetBytes(b)
	n := new(big.Int).Sub(curveParams.N, one)
	k.Mod(k, n)
	k.Add(k, one)
	return k
}
