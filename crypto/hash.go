package crypto

import (
	"hash"

	"github.com/denisskin/bin"
	"golang.org/x/crypto/sha3"
)

func hash256(data []byte) []byte {
	h := sha3.Sum256(data)
	return h[:]
}

func newHash256() hash.Hash {
	return sha3.New256()
}

func Hash256Raw(data []byte) []byte {
	return hash256(data)
}

func Hash256(values ...interface{}) []byte {
	h := newHash256()
	w := bin.NewWriter(h)
	for _, val := range values {
		w.WriteVar(val)
	}
	return h.Sum(nil)
}
