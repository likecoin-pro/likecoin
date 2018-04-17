package patricia

import (
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/likecoin-pro/likecoin/crypto/merkle"
	"github.com/stretchr/testify/assert"
)

func TestTree_Root(t *testing.T) {
	tree := NewTree(nil, nil)
	for _, v := range randSlice(10000) {
		tree.Put(v)
	}

	aRoot, _ := tree.Root()

	assert.Equal(t, "ee50ff497016e338ad8b41f320245cd6bc9d7efcccdc32143d49f47cd7408114", hex.EncodeToString(aRoot))
}

func TestTree_Put(t *testing.T) {
	a := NewTree(nil, nil)
	b := NewTree(nil, nil)

	for _, v := range randSlice(10000) {
		a.Put(v)
	}
	for _, v := range randSlice(10000) {
		a.Put(v)
		b.Put(v)
	}

	aRoot, _ := a.Root()
	bRoot, _ := b.Root()

	assert.Equal(t, aRoot, bRoot)
}

func TestTree_GetProof(t *testing.T) {
	a := NewTree(nil, nil)
	keys := randSlice(5000)

	for _, key := range keys {
		err := a.Put(key)
		assert.NoError(t, err)

		proof, root, err := a.GetProof(key)
		assert.NoError(t, err)

		ok := merkle.Verify(key, proof, root)
		assert.True(t, ok)
	}

	for _, key := range keys {
		err := a.Put(key)
		assert.Error(t, err)
		assert.Equal(t, errKeyHasExists, err)

		proof, root, err := a.GetProof(key)
		assert.NoError(t, err)

		ok := merkle.Verify(key, proof, root)
		assert.True(t, ok)
	}
}

func TestTree_AppendingProof(t *testing.T) {
	a := NewTree(nil, nil)
	keys := randSlice(5000)

	for i, key := range keys {

		root1, err := a.Root()
		assert.NoError(t, err)

		// get appending key proof and new root
		proof2A, root2A, err := a.AppendingProof(key)
		assert.NoError(t, err)
		assert.NotEqual(t, root1, root2A) // new root
		if i > 0 {
			assert.True(t, len(proof2A) > 0)
		}

		// verify proof
		ok := merkle.Verify(key, proof2A, root2A)
		assert.True(t, ok)

		// check root; root is not changed
		root1A, err := a.Root()
		assert.NoError(t, err)
		assert.Equal(t, root1, root1A)

		// insert new key
		err = a.Put(key)
		assert.NoError(t, err)

		// get new root
		root2, err := a.Root()
		assert.NoError(t, err)
		assert.Equal(t, root2A, root2)

		// get new proof
		proof2B, root2B, err := a.GetProof(key)
		assert.NoError(t, err)
		assert.Equal(t, proof2A, proof2B)
		assert.Equal(t, root2A, root2B)
		assert.Equal(t, root2, root2B)
	}
}

func randSlice(n int) [][]byte {
	r := rand.New(rand.NewSource(0))
	a := make([][]byte, n)
	for i := range a {
		a[i] = make([]byte, merkle.HashSize)
		r.Read(a[i])
	}
	rand.Shuffle(len(a), func(i, j int) {
		a[i], a[j] = a[j], a[i]
	})
	return a
}
