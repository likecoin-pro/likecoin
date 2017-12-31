package base58

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {

	data := []byte("abcdefg012345")

	str := Encode(data)

	assert.Equal(t, "97TZNJviu2mG6A4VFr", str)
}

func TestEncode_WithNullPrefix(t *testing.T) {

	data := []byte("\x00\x00\x00\x00abcdefg012345")

	str := Encode(data)

	assert.Equal(t, "97TZNJviu2mG6A4VFr", str)
}

func TestDecode(t *testing.T) {

	str := "97TZNJviu2mG6A4VFr"

	data, err := Decode(str)

	assert.NoError(t, err)
	assert.Equal(t, "abcdefg012345", string(data))
}
