package crypto

import (
	"crypto/rand"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

const (
	// KeySize is Size of Private Key in bytes
	KeySize       = 256 / 8       // 32 bytes
	publicKeySize = KeySize + 1   // 33 bytes
	signatureSize = KeySize*2 + 1 // 65 bytes
)

var (
	one = big.NewInt(1)

	curve    = secp256k1.S256()
	curveP   = curve.Params().P
	curveN   = curve.Params().N
	curveN_1 = new(big.Int).Sub(curveN, one)
)

// ------------------------------------
func intToBytes(i *big.Int) []byte {
	buf := make([]byte, KeySize)
	b := i.Bytes()
	copy(buf[KeySize-len(b):], b)
	return buf
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
	k.Mod(k, curveN_1)
	k.Add(k, one)
	return k
}
