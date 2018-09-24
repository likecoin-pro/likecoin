package crypto

import (
	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/crypto/sha3"
)

func HashSum256(data []byte) []byte {
	var hash [32]byte
	sha := sha3.NewShake256()
	sha.Write(data)
	sha.Read(hash[:])
	return hash[:]
}

func Hash256(values ...interface{}) []byte {
	var hash [32]byte
	sha := sha3.NewShake256()
	w := bin.NewWriter(sha)
	for _, val := range values {
		w.WriteVar(val)
	}
	sha.Read(hash[:])
	return hash[:]
}
