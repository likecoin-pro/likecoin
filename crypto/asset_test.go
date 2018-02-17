package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsset_Decode(t *testing.T) {
	a := Asset([]byte{1, 2, 3})
	data := a.Encode()

	var b Asset
	err := b.Decode(data)

	assert.NoError(t, err)
	assert.Equal(t, a, b)
}
