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

		assert.True(t, strings.HasPrefix(sAddr, "LikeC"))
		assert.Equal(t, 44, len(sAddr))
	}
}

func TestAddress_isValidBase58(t *testing.T) {
	strAddr := newAddress(randBytes(AddressSize)).String()

	_, err := base58.Decode(strAddr)

	assert.NoError(t, err)
}

func TestAddress_Encode(t *testing.T) {
	addr := newAddress(randBytes(AddressSize))

	sAddr := addr.Encode(666)

	assert.True(t, strings.HasPrefix(sAddr, "Like"))
	assert.True(t, len(sAddr) > 44)
}

func TestParseAddress(t *testing.T) {
	strAddr := "LikeC3ercQghTCsynJiNBQdMNhzxGYegR44Dw6mzQije"

	addr, magicNum, err := ParseAddress(strAddr)

	assert.NoError(t, err)
	assert.Equal(t, "Playing with a Full Deck", string(addr[:]))
	assert.Equal(t, int64(0), magicNum)
}

func TestParseAddress_withMagic(t *testing.T) {
	strAddr := newAddress(randBytes(AddressSize)).Encode(0x19720000abba0000)

	_, magicNum, err := ParseAddress(strAddr)

	assert.NoError(t, err)
	assert.Equal(t, int64(0x19720000abba0000), magicNum)
}

func TestParseAddress_Ex(t *testing.T) {
	for i := 1; i < 1e3; i++ {
		strAddr := newAddress(randBytes(AddressSize)).Encode(int64(i))

		_, mg, err := ParseAddress(strAddr)

		assert.NoError(t, err)
		assert.Equal(t, int64(i), mg)
	}
}

func TestParseAddress_Fail(t *testing.T) {
	strAddr := newAddress(randBytes(AddressSize)).String()
	strAddr += "a"

	_, _, err := ParseAddress(strAddr)

	assert.Error(t, err)
}
