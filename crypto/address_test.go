package crypto

import (
	"strings"
	"testing"

	"github.com/likecoin-pro/likecoin/crypto/base58"
	"github.com/stretchr/testify/assert"
)

func TestAddress_String(t *testing.T) {
	for i := 1; i < 1e3; i++ {
		addr := newAddress(randBytes(AddressSize))

		sAddr := addr.String()

		assert.True(t, strings.HasPrefix(sAddr, "Like"))
		assert.Equal(t, 43, len(sAddr))
	}
}

func TestAddress_isValidBase58(t *testing.T) {
	strAddr := newAddress(randBytes(AddressSize)).String()

	_, err := base58.Decode(strAddr)

	assert.NoError(t, err)
}

func TestAddress_IsNil(t *testing.T) {
	var addr Address

	isNil := addr.IsNil()

	assert.True(t, isNil)
}

func TestAddress_TaggedString(t *testing.T) {
	addr := newAddress(randBytes(AddressSize))

	sAddr := addr.TaggedString(666)

	assert.True(t, strings.HasPrefix(sAddr, "Like"))
	assert.True(t, len(sAddr) > 43)
}

func TestParseAddress(t *testing.T) {
	randData := randBytes(AddressSize)
	strAddr := newAddress(randData).String()

	addr, tag, err := ParseAddress(strAddr)

	assert.NoError(t, err)
	assert.Equal(t, randData, addr[:])
	assert.EqualValues(t, 0, tag)
}

func TestParseAddress_withTag(t *testing.T) {
	strAddr := newAddress(randBytes(AddressSize)).TaggedString(0x19720000abba0000)

	_, tag, err := ParseAddress(strAddr)

	assert.NoError(t, err)
	assert.EqualValues(t, 0x19720000abba0000, tag)
}

func TestParseAddress_withTag2(t *testing.T) {
	for i := uint64(1); i < 1e3; i++ {
		strAddr := newAddress(randBytes(AddressSize)).TaggedString(i)

		_, tag, err := ParseAddress(strAddr)

		assert.NoError(t, err)
		assert.Equal(t, i, tag)
	}
}

func TestParseAddress_withTagSuffix(t *testing.T) {
	strAddr := "Like5A2PEu6eCHQzy1tMsa6b3kc1xXS7ywj2NQaSsgE0x666"

	addr, tag, err := ParseAddress(strAddr)

	assert.NoError(t, err)
	assert.Equal(t, "Like5A2PEu6eCHQzy1tMsa6b3kc1xXS7ywj2NQaSsgE", addr.String())
	assert.EqualValues(t, 0x666, tag)
}

func TestParseAddress_fail(t *testing.T) {
	strAddr := newAddress(randBytes(AddressSize)).String()
	strAddr += "a"

	_, _, err := ParseAddress(strAddr)

	assert.Error(t, err)
}

func TestAddress_Encode_nilAddress(t *testing.T) {
	var addr Address

	data := addr.Encode()

	assert.Equal(t, 0, len(data))
}

func TestAddress_Decode(t *testing.T) {
	a := newAddress(randBytes(AddressSize))
	data := a.Encode()

	var b Address
	err := b.Decode(data)

	assert.NoError(t, err)
	assert.Equal(t, a, b)
}

func TestAddress_Decode_nilAddress(t *testing.T) {
	var addr Address
	err := addr.Decode(nil)

	assert.NoError(t, err)
	assert.True(t, addr.IsNil())
}
