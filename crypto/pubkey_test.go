package crypto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublicKey_Encode(t *testing.T) {
	pub := NewPrivateKey().PublicKey

	buf := pub.Encode()

	assert.Equal(t, publicKeySize, len(buf))
}

func TestPublicKey_Decode(t *testing.T) {
	pub1 := NewPrivateKey().PublicKey
	buf := pub1.Encode()

	var pub2 = new(PublicKey)
	err := pub2.Decode(buf)

	assert.NoError(t, err)
	assert.True(t, pub1.Equal(pub2))
	assert.Equal(t, pub1, pub2)
}

func TestDecodePublicKey(t *testing.T) {
	pub1 := NewPrivateKey().PublicKey
	buf := pub1.Encode()

	pub2, err := DecodePublicKey(buf)

	assert.NoError(t, err)
	assert.Equal(t, pub1, pub2)
}

func TestPublicKey_String(t *testing.T) {
	pub := MustParsePublicKey("rggH2X4N7JsBtekY1isiutwJZpRhQQoeYaVqYdRSH4mR")

	str := pub.String()

	assert.Equal(t, "rggH2X4N7JsBtekY1isiutwJZpRhQQoeYaVqYdRSH4mR", str)
}

func TestPublicKey_MarshalJSON(t *testing.T) {
	pub := MustParsePublicKey("rggH2X4N7JsBtekY1isiutwJZpRhQQoeYaVqYdRSH4mR")

	data, err := json.Marshal(pub)

	assert.NoError(t, err)
	assert.Equal(t, `"rggH2X4N7JsBtekY1isiutwJZpRhQQoeYaVqYdRSH4mR"`, string(data))
}

func TestPublicKey_UnmarshalJSON(t *testing.T) {
	pub1 := MustParsePublicKey("rggH2X4N7JsBtekY1isiutwJZpRhQQoeYaVqYdRSH4mR")
	buf, _ := json.Marshal(pub1)

	var pub2 *PublicKey
	err := json.Unmarshal(buf, &pub2)

	assert.NoError(t, err)
	assert.Equal(t, pub1.x, pub2.x)
	assert.Equal(t, pub1.y, pub2.y)
	assert.Equal(t, pub1.Encode(), pub2.Encode())
	assert.True(t, pub1.Equal(pub2))
}

func TestPublicKey_Is(t *testing.T) {
	var pub0 *PublicKey
	pub1 := MustParsePublicKey("rggH2X4N7JsBtekY1isiutwJZpRhQQoeYaVqYdRSH4mR")
	pub2 := MustParsePublicKey("rggH2X4N7JsBtekY1isiutwJZpRhQQoeYaVqYdRSH4mR")
	pub3 := MustParsePublicKey("rggH2X4N7JsBtekY1isiutwJZpRhQQoeYaVqYdRSH4mS")

	assert.True(t, pub1.Equal(pub2))
	assert.True(t, pub1.Equal(pub2))
	assert.True(t, !pub0.Equal(pub1))
	assert.True(t, !pub1.Equal(pub0))
	assert.True(t, !pub1.Equal(pub3))
}

func TestPublicKey_Empty(t *testing.T) {

	var pub1 *PublicKey
	var pub2 = new(PublicKey)
	var pub3 = MustParsePublicKey("rggH2X4N7JsBtekY1isiutwJZpRhQQoeYaVqYdRSH4mR")

	assert.True(t, pub1.Empty())
	assert.True(t, pub2.Empty())
	assert.True(t, !pub3.Empty())
}
