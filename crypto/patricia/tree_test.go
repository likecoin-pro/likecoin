package patricia

import (
	"encoding/hex"
	"testing"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/crypto/merkle"
	"github.com/stretchr/testify/assert"
)

func TestTree_Root(t *testing.T) {
	tree := NewTree(nil)
	for k, v := range testValues(5000) {
		tree.PutVar(k, v)
	}

	root, _ := tree.Root()

	assert.Equal(t, "8d3846d06998ea8ba2cb81e79bebd05b98d2e04067671b0e69a0eee8c79567ba", hex.EncodeToString(root))
}

func TestTree_Root_empty(t *testing.T) {
	tree := NewTree(nil)

	root0, _ := tree.Root()

	assert.Equal(t, "", hex.EncodeToString(root0))
}

func TestTree_Root_one(t *testing.T) {
	tree := NewTree(nil)
	tree.PutVar("123", "abc")

	root, _ := tree.Root()

	assert.Equal(t, merkle.Root([]byte("123"), []byte("abc")), root)
}

func TestTree_Put(t *testing.T) {
	a := NewTree(nil)
	b := NewTree(nil)

	for k, v := range testValues(5000) {
		a.PutVar(k, v)
	}
	for k, v := range testValues(5000) {
		a.PutVar(k, v)
		b.PutVar(k, v)
	}

	aRoot, _ := a.Root()
	bRoot, _ := b.Root()

	assert.Equal(t, aRoot, bRoot)
}

func TestTree_Get(t *testing.T) {
	a := NewTree(nil)
	for k, v := range testValues(5000) {
		a.Put(encode(k), v)
	}

	for k, v := range testValues(5000) {
		val, err := a.Get(encode(k))

		assert.NoError(t, err)
		assert.Equal(t, v, val)
	}
}

func TestTree_GetProof(t *testing.T) {
	a := NewTree(nil)
	keys := testValues(5000)

	for k, v := range keys {
		key := encode(k)
		err := a.Put(key, v)
		assert.NoError(t, err)

		val, proof, root, err := a.GetProof(key)
		assert.NoError(t, err)
		assert.Equal(t, v, val)

		ok := merkle.Verify(merkle.Root(key, val), proof, root)
		assert.True(t, ok)
	}

	for k, v := range keys {
		key := encode(k)

		err := a.Put(key, v)
		assert.NoError(t, err)

		val, proof, root, err := a.GetProof(key)
		assert.NoError(t, err)
		assert.Equal(t, v, val)

		ok := merkle.Verify(merkle.Root(key, val), proof, root)
		assert.True(t, ok)
	}
}

func TestTree_GetProof_len(t *testing.T) {
	tree := NewTree(nil)

	key := []byte("001")
	tree.Put(key, []byte("val"))
	val, proof, root, err := tree.GetProof(key)
	verify := merkle.Verify(merkle.Root(key, val), proof, root)
	assert.NoError(t, err)
	assert.True(t, verify)
	assert.Equal(t, 0, len(proof))

	key = []byte("002")
	tree.Put(key, []byte("val"))
	val, proof, root, err = tree.GetProof(key)
	verify = merkle.Verify(merkle.Root(key, val), proof, root)
	assert.NoError(t, err)
	assert.True(t, verify)
	assert.Equal(t, 33, len(proof))

	key = []byte("003")
	tree.Put(key, []byte("val"))
	val, proof, root, err = tree.GetProof(key)
	verify = merkle.Verify(merkle.Root(key, val), proof, root)
	assert.NoError(t, err)
	assert.True(t, verify)
	assert.Equal(t, 33, len(proof))

	key = []byte("004")
	tree.Put(key, []byte("val"))
	val, proof, root, err = tree.GetProof(key)
	verify = merkle.Verify(merkle.Root(key, val), proof, root)
	assert.NoError(t, err)
	assert.True(t, verify)
	assert.Equal(t, 66, len(proof))

	key = []byte("010")
	tree.Put(key, []byte("val"))
	val, proof, root, err = tree.GetProof(key)
	verify = merkle.Verify(merkle.Root(key, val), proof, root)
	assert.NoError(t, err)
	assert.True(t, verify)
	assert.Equal(t, 33, len(proof))
}

func TestTree_AppendingProof(t *testing.T) {
	a := NewTree(nil)

	for k, val := range testValues(5000) {
		key := encode(k)

		root1, err := a.Root()
		assert.NoError(t, err)

		// get appending key proof and new root
		proof2A, root2A, err := a.AppendingProof(key, val)
		assert.NoError(t, err)
		assert.NotEqual(t, root1, root2A) // new root

		// verify proof
		ok := merkle.Verify(merkle.Root(key, val), proof2A, root2A)
		assert.True(t, ok)

		// check root; root is not changed
		root1A, err := a.Root()
		assert.NoError(t, err)
		assert.Equal(t, root1, root1A)

		// insert new key
		err = a.Put(key, val)
		assert.NoError(t, err)

		// get new root
		root2, err := a.Root()
		assert.NoError(t, err)
		assert.Equal(t, root2A, root2)

		// get new proof
		val2B, proof2B, root2B, err := a.GetProof(key)
		assert.NoError(t, err)
		assert.Equal(t, val, val2B)
		assert.Equal(t, proof2A, proof2B)
		assert.Equal(t, root2A, root2B)
		assert.Equal(t, root2, root2B)
	}
}

func testValues(n int) map[int][]byte {
	v := make(map[int][]byte, n)
	for i := 0; i < n; i++ {
		v[i] = bin.Hash128(i)
	}
	return v
}
