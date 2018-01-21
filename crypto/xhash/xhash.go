package xhash

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"sort"
)

func GenerateKeyByPassword(pass string, keyLen int) []byte {
	return XHash([]byte(pass))[:keyLen/8]
}

const countRounds = 84673

func XHash(data []byte) []byte {
	bb := make([][]byte, countRounds)
	for i := 0; i < countRounds; i++ {
		r := sha256.Sum256(data)
		data = r[:]
		bb[i] = data
	}
	sort.Slice(bb, func(i, j int) bool {
		return bytes.Compare(bb[i], bb[j]) == -1
	})
	h512 := sha512.New()
	for _, b := range bb {
		h512.Write(b)
	}
	return h512.Sum(nil)
}
