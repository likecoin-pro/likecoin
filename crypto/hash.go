package crypto

import (
	"hash"

	"github.com/denisskin/bin"
	"github.com/ethereum/go-ethereum/crypto/sha3"
)

func newHash256() hash.Hash {
	return sha3.NewKeccak256()
}

func HashSum256(data []byte) []byte {
	sha := newHash256()
	sha.Write(data)
	return sha.Sum(nil)
}

func Hash256(values ...interface{}) []byte {
	h := newHash256()
	w := bin.NewWriter(h)
	for _, val := range values {
		w.WriteVar(val)
	}
	return h.Sum(nil)
}
