package merkle

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testData = [][]byte{
	[]byte("Abc"),
	[]byte("def"),
	[]byte("1234"),
	[]byte("56789"),
	[]byte("ёпрст"),
}

func TestRoot(t *testing.T) {

	root := Root(testData)

	assert.Equal(t, "9a6d2ec4cdd602faf3330086260dec73cd83306dcf8ae1ed7bf6fecd8e11112e", hex.EncodeToString(root))
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
