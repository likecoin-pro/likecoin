package merkle

import (
	"bytes"

	"golang.org/x/crypto/sha3"
)

const HashSize = 256 / 8 // hash size

func hash(a, b []byte) []byte {
	h := sha3.New256()
	h.Write(a)
	h.Write(b)
	return h.Sum(nil)
}

func copySlice(src [][]byte) [][]byte {
	bb := make([][]byte, len(src))
	copy(bb, src)
	return bb
}

func Root(hashes [][]byte) []byte {
	if len(hashes) == 0 {
		return nil
	}
	hashes = copySlice(hashes)
	for len(hashes) > 1 {
		n := len(hashes)
		m := n / 2
		for j := 0; j < m; j++ {
			hashes[j] = hash(hashes[j*2], hashes[j*2+1])
		}
		if n%2 != 0 {
			hashes[m] = hashes[n-1]
			m++
		}
		hashes = hashes[:m]
	}
	return hashes[0]
}

func Proof(hashes [][]byte, i int) (proof []byte) {
	hashes = copySlice(hashes)
	for len(hashes) > 1 {
		n := len(hashes)
		m := n / 2
		if i < m*2 {
			if i%2 == 0 {
				proof = append(proof, 0)
				proof = append(proof, hashes[i+1]...)
			} else {
				proof = append(proof, 1)
				proof = append(proof, hashes[i-1]...)
			}
		}
		i /= 2
		for j := 0; j < m; j++ {
			hashes[j] = hash(hashes[j*2], hashes[j*2+1])
		}
		if n%2 != 0 {
			hashes[m] = hashes[n-1]
			m++
		}
		hashes = hashes[:m]
	}
	return
}

func Verify(key, root, proof []byte) bool {
	for len(proof) > 0 {
		if len(proof) < HashSize+1 {
			return false
		}
		if proof[0] == 0 {
			key = hash(key, proof[1:HashSize+1])
		} else {
			key = hash(proof[1:HashSize+1], key)
		}
		proof = proof[HashSize+1:]
	}
	return bytes.Equal(key, root)
}
