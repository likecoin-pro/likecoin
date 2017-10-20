package merkle

import "golang.org/x/crypto/sha3"

func hash256(data []byte) []byte {
	h := sha3.Sum256(data)
	return h[:]
}

func Root(hh [][]byte) []byte {
	if len(hh) == 0 {
		return nil
	}
	for len(hh) > 1 {
		n := len(hh)
		m := n / 2
		for i := 0; i < m; i++ {
			hh[i] = hash256(append(hh[i*2], hh[i*2+1]...))
		}
		if n%2 != 0 {
			hh[m] = hh[n-1]
			m++
		}
		hh = hh[:m]
	}
	return hh[0]
}

//func MerkleProf(leafs [][]byte, i int) (res [][]byte) {
//	return nil
//}

//func MerkleVerify(iPart uint64, hash []byte, lvHashes [][]byte) (root []byte) {
//	root = hash
//	for _, h := range lvHashes {
//		root = hash256()
//	}
//}
