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
	keys := randSlice(10000)
	for _, v := range keys {
		a.Put(v)
	}
	root, _ := a.Root()

	for _, key := range keys {

		proof, err := a.GetProof(key)

		ok := merkle.Verify(key, root, proof)

		assert.NoError(t, err)
		assert.True(t, len(proof) > 0)
		assert.True(t, ok)
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
