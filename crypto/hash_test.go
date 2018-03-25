package crypto

import (
	"testing"

	"github.com/denisskin/bin"
	"github.com/stretchr/testify/assert"
)

func TestHash256(t *testing.T) {
	a, b, c := 1, 2, 3

	h1 := Hash256(a, b, c)
	h2 := Hash256Raw(bin.Encode(a, b, c))

	assert.Equal(t, h1, h2)
}
