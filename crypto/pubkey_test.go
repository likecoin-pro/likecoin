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

func TestPublicKey_Hex(t *testing.T) {
	pub := NewPrivateKeyBySecret("abc").PublicKey

	hex := pub.Hex()

	assert.Equal(t, "0x022f86f8c408c20e8bdcef6471676a2157624915355fe662b568ac5e2a2a76fed5d34d4a184176a3e4a28bac7203a860510e363601f7c8f8657067173ed83f6e", hex)
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

func TestRecoverPublicKey(t *testing.T) {
	prv := NewPrivateKey()
	pub1 := prv.PublicKey
	hash := HashSum256([]byte("Test text"))
	sign := prv.Sign(hash)

	pub2, err := RecoverPublicKey(hash, sign)
	verify := pub2.Verify(hash, sign)

	assert.NoError(t, err)
	assert.True(t, pub1.Equal(pub2))
	assert.True(t, verify)
}

func TestRecoverPublicKey_fail(t *testing.T) {
	prv := NewPrivateKey()
	pub1 := prv.PublicKey
	hash := HashSum256([]byte("Test text"))
	sign := prv.Sign(hash)

	sign[13]++ // corrupt signature

	pub2, err := RecoverPublicKey(hash, sign)

	assert.True(t, err != nil || !pub1.Equal(pub2))
}
