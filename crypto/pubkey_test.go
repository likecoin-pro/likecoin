package crypto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testPub = MustParsePublicKey("4pv2QxPs618pCCokGdD2U71A2ANfNc59i4xQavpKM9L3QjDw7mGLSzbrBkGmcjEzGJWTD6AadgmM9kZp8nssUyfa")

func TestPublicKey_Encode(t *testing.T) {

	buf := testPub.Encode()

	assert.Equal(t, PublicKeySize, len(buf))
}

func TestPublicKey_Decode(t *testing.T) {
	buf := testPub.Encode()

	var pub = new(PublicKey)
	err := pub.Decode(buf)

	assert.NoError(t, err)
	assert.True(t, testPub.Equal(pub))
	assert.Equal(t, testPub, pub)
}

func TestDecodePublicKey(t *testing.T) {
	buf := testPub.Encode()

	pub, err := DecodePublicKey(buf)

	assert.Equal(t, testPub, pub)
	assert.NoError(t, err)
}

func TestPublicKey_String(t *testing.T) {

	str := testPub.String()

	assert.Equal(t, "4pv2QxPs618pCCokGdD2U71A2ANfNc59i4xQavpKM9L3QjDw7mGLSzbrBkGmcjEzGJWTD6AadgmM9kZp8nssUyfa", str)
}

func TestPublicKey_MarshalJSON(t *testing.T) {

	data, err := json.Marshal(testPub)

	assert.NoError(t, err)
	assert.Equal(t, `"4pv2QxPs618pCCokGdD2U71A2ANfNc59i4xQavpKM9L3QjDw7mGLSzbrBkGmcjEzGJWTD6AadgmM9kZp8nssUyfa"`, string(data))
}

func TestPublicKey_UnmarshalJSON(t *testing.T) {
	buf, _ := json.Marshal(testPub)

	var pub *PublicKey
	err := json.Unmarshal(buf, &pub)

	assert.NoError(t, err)
	assert.Equal(t, testPub, pub)
	assert.True(t, testPub.Equal(pub))
}

func TestPublicKey_Is(t *testing.T) {
	var pub0 *PublicKey
	pub1 := MustParsePublicKey("4pv2QxPs618pCCokGdD2U71A2ANfNc59i4xQavpKM9L3QjDw7mGLSzbrBkGmcjEzGJWTD6AadgmM9kZp8nssUyfa")
	pub2 := MustParsePublicKey("4pv2QxPs618pCCokGdD2U71A2ANfNc59i4xQavpKM9L3QjDw7mGLSzbrBkGmcjEzGJWTD6AadgmM9kZp8nssUyfa")
	pub3 := MustParsePublicKey("5pv2QxPs618pCCokGdD2U71A2ANfNc59i4xQavpKM9L3QjDw7mGLSzbrBkGmcjEzGJWTD6AadgmM9kZp8nssUyfa")

	assert.True(t, pub1.Equal(pub2))
	assert.True(t, pub1.Equal(pub2))
	assert.True(t, !pub0.Equal(pub1))
	assert.True(t, !pub1.Equal(pub0))
	assert.True(t, !pub1.Equal(pub3))
}

func TestPublicKey_Empty(t *testing.T) {

	var pub1 *PublicKey
	var pub2 = new(PublicKey)

	assert.True(t, pub1.Empty())
	assert.True(t, pub2.Empty())
	assert.True(t, !testPub.Empty())
}
