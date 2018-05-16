package xhash

import (
	"bytes"
	"sort"

	"golang.org/x/crypto/sha3"
)

func GenerateKeyByPassword(pass string, keyLen int) []byte {
	return XHash([]byte(pass))[:keyLen/8]
}

func XHash(data []byte) []byte {
	const rounds = 200003
	bb := make([][]byte, rounds)
	for i := 0; i < rounds; i++ {
		data = shake256(data)
		bb[i] = data[16:]
	}
	sort.Slice(bb, func(i, j int) bool {
		return bytes.Compare(bb[i], bb[j]) == -1
	})
	h := sha3.NewShake256()
	for _, b := range bb {
		h.Write(b)
	}
	buf := make([]byte, 64)
	h.Read(buf)
	return buf
}

func shake256(data []byte) []byte {
	h := sha3.NewShake256()
	h.Write(data)
	buf := make([]byte, 32)
	h.Read(buf)
	return buf
}
