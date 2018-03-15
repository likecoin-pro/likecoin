package merkle

import (
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoot(t *testing.T) {
	hashes := newHashes(15)

	root := Root(hashes)

	assert.Equal(t, "961d250d8cb847264302f32ce6afab3e6116e065f48074b9a9a05db2b9f026e5", str(root))
}

func TestProof(t *testing.T) {
	hashes := newHashes(15)

	proof := Proof(hashes, 4)

	assert.Equal(t, ""+
		"004f5aa6aec3fc78c6aae081ac8120c720efcd6cea84b6925e607be063716f96dd"+
		"00a2f8026f773c4717044f67fe0ff9554e8b472670d778462d447179ad52d8c50a"+
		"01502d4dcaced1fbefdfcea7725a9a71dd4957da154c01ccd5177709ef2aa67c52"+
		"00e63fe76c95bf01398570cf7781cc6dc254ffb5e740df2c5bf8ae7bdeb5de3eb8",
		str(proof),
	)
}

func TestVerify(t *testing.T) {

	hashes := newHashes(100)
	root := Root(hashes)

	for i, hash := range hashes {
		proof := Proof(hashes, i)

		ok := Verify(hash, root, proof)

		assert.True(t, ok)
	}
}

func TestVerify_fail(t *testing.T) {
	hashes := newHashes(100)
	root := Root(hashes)
	proof := Proof(hashes, 13)

	proof = proof[:len(proof)-1] // corrupt proof (cut last byte)
	ok := Verify(hashes[13], root, proof)

	assert.False(t, ok)
}

//-------------------------------------------------------------------
func newHashes(n int) (data [][]byte) {
	r := rand.New(rand.NewSource(0))
	for i := 0; i < n; i++ {
		buf := make([]byte, HashSize)
		r.Read(buf)
		data = append(data, buf)
	}
	return
}

func str(b []byte) string {
	return hex.EncodeToString(b)
}
