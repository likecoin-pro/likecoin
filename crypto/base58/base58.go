package base58

import (
	"bytes"
	"errors"
	"math/big"
	"strings"

	"github.com/denisskin/bin"
)

var (
	// alphabet is the modified base58 alphabet used by Bitcoin.
	btcAlphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

	zero  = big.NewInt(0)
	radix = big.NewInt(58)

	errInvalidNumber = errors.New("invalid base58 number")
)

// Decode decodes a modified base58 string to a byte slice, using BTC-Alphabet
func Decode(b string) ([]byte, error) {
	return decodeAlphabet([]byte(strings.TrimSpace(b)), btcAlphabet)
}

func DecodeFixed(b string, n int) (res []byte, err error) {
	if res, err = Decode(b); err == nil && len(res) < n {
		res = append(make([]byte, n-len(res)), res...)
	}
	return
}

// Encode encodes a byte slice to a modified base58 string, using BTC-Alphabet
func Encode(b []byte) string {
	return encodeAlphabet(b, btcAlphabet)
}

func Itoa(i uint64) string {
	return Encode(bin.Uint64ToBytes(i))
}

func Atoi(s string) (i uint64, err error) {
	b, err := Decode(s)
	if err == nil {
		i = bin.BytesToUint64(b)
	}
	return
}

// DecodeAlphabet decodes a modified base58 string to a byte slice, using alphabet.
func decodeAlphabet(b, alphabet []byte) ([]byte, error) {
	res := big.NewInt(0)
	j := big.NewInt(1)
	for i := len(b) - 1; i >= 0; i-- {
		if pos := bytes.IndexByte(alphabet, b[i]); pos == -1 {
			return nil, errInvalidNumber
		} else {
			idx := big.NewInt(int64(pos))
			t := big.NewInt(0)
			t.Mul(j, idx)

			res.Add(res, t)
			j.Mul(j, radix)
		}
	}
	return res.Bytes(), nil
}

// Encode encodes a byte slice to a modified base58 string, using alphabet
func encodeAlphabet(b, alphabet []byte) string {
	x := new(big.Int).SetBytes(b)
	res := make([]byte, 0, len(b)*136/100)
	for x.Cmp(zero) > 0 {
		mod := new(big.Int)
		x.DivMod(x, radix, mod)
		res = append(res, alphabet[mod.Int64()])
	}
	// reverse
	for i, n := 0, len(res); i < n/2; i++ {
		res[i], res[n-1-i] = res[n-1-i], res[i]
	}
	return string(res)
}
