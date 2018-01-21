package crypto

import (
	"crypto/sha256"
	"hash"
)

func hash256(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

func newHash256() hash.Hash {
	return sha256.New()
}
