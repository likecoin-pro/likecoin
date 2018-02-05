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

func TestAddress_ExtendedString(t *testing.T) {
	addr := newAddress(randBytes(AddressSize))

	sAddr := addr.ExtendedString(666)

	assert.True(t, strings.HasPrefix(sAddr, "Like"))
	assert.True(t, len(sAddr) > 43)
}

func TestParseAddress(t *testing.T) {
	randData := randBytes(AddressSize)
	strAddr := newAddress(randData).String()

	addr, magicNum, err := ParseAddress(strAddr)

	assert.NoError(t, err)
	assert.Equal(t, randData, addr[:])
	assert.Equal(t, int64(0), magicNum)
}

func TestParseAddress_withMagicNum(t *testing.T) {
	strAddr := newAddress(randBytes(AddressSize)).ExtendedString(0x19720000abba0000)

	_, magicNum, err := ParseAddress(strAddr)

	assert.NoError(t, err)
	assert.Equal(t, int64(0x19720000abba0000), magicNum)
}

func TestParseAddress_withMagicNum2(t *testing.T) {
	for i := 1; i < 1e3; i++ {
		strAddr := newAddress(randBytes(AddressSize)).ExtendedString(int64(i))

		_, mg, err := ParseAddress(strAddr)

		assert.NoError(t, err)
		assert.Equal(t, int64(i), mg)
	}
}

func TestParseAddress_fail(t *testing.T) {
	strAddr := newAddress(randBytes(AddressSize)).String()
	strAddr += "a"

	_, _, err := ParseAddress(strAddr)

	assert.Error(t, err)
}

func TestAddress_Decode(t *testing.T) {
	a := newAddress(randBytes(AddressSize))
	data := a.Encode()

	var b Address
	err := b.Decode(data)

	assert.NoError(t, err)
	assert.Equal(t, a, b)
}
