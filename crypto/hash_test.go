package crypto

import (
	"testing"

	"encoding/hex"

	"github.com/denisskin/bin"
	"github.com/stretchr/testify/assert"
)

func TestHash256(t *testing.T) {
	a, b, c := 1, 2, 3

	h1 := Hash256(a, b, c)
	h2 := HashSum256(bin.Encode(a, b, c))

	assert.Equal(t, h1, h2)
}

func TestHashSum256(t *testing.T) {

	hash0 := HashSum256([]byte(""))
	hash1 := HashSum256([]byte("abc"))

	assert.Equal(t, "46b9dd2b0ba88d13233b3feb743eeb243fcd52ea62b81b82b50c27646ed5762f", hex.EncodeToString(hash0))
	assert.Equal(t, "483366601360a8771c6863080cc4114d8db44530f8f1e1ee4f94ea37e78b5739", hex.EncodeToString(hash1))
}
