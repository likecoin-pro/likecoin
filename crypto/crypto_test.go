package crypto

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePrivateKey(t *testing.T) {

	prv1 := NewPrivateKey()
	prv2 := NewPrivateKey()
	pub1 := prv1.PublicKey

	assert.Equal(t, KeySize+1, len(prv1.Encode()))
	assert.NotEqual(t, prv1, prv2)
	assert.Equal(t, publicKeySize, len(pub1.Encode()))
}

func TestGeneratePrivateKeyByPassword(t *testing.T) {
	password := "SuperPuperSecret"

	prv1 := NewPrivateKeyBySecret(password)
	pub1 := prv1.PublicKey

	prv2 := NewPrivateKeyBySecret(password)
	pub2 := prv2.PublicKey

	prv3 := NewPrivateKeyBySecret(password + ".")
	pub3 := prv3.PublicKey

	assert.Equal(t, "X9K2MudKn3qwitjGvQ2mp3TwwqTjRcvzLWkQ8aMzhCva", prv1.String())
	assert.Equal(t, "gwfhXouT2r7XjZTge8ZEPmmLqD3fdS9ZSztLutrELsQj", pub1.String())
	assert.Equal(t, 33, len(prv1.Encode()))
	assert.Equal(t, prv1.Encode(), prv2.Encode())
	assert.Equal(t, pub1.Encode(), pub2.Encode())
	assert.NotEqual(t, prv1.Encode(), prv3.Encode())
	assert.NotEqual(t, pub1.Encode(), pub3.Encode())
}

func TestSign_Deterministic(t *testing.T) {
	prv := NewPrivateKey()
	hash := hash256([]byte("Лавров выразил надежду, что «ловцы покемонов» не разучатся говорить"))

	sign1 := prv.Sign(hash)
	sign2 := prv.Sign(hash)

	assert.Equal(t, signatureSize, len(sign1))
	assert.Equal(t, signatureSize, len(sign2))
	assert.True(t, bytes.Equal(sign1, sign2))
}

func TestVerify(t *testing.T) {
	prv := NewPrivateKey()
	pub := prv.PublicKey
	hash := hash256([]byte("Совет по туризму Норвегии определил места обитания редких покемонов"))
	sign := prv.Sign(hash)

	verify1 := pub.Verify(hash, sign)
	verify2 := pub.Verify(hash, sign)

	assert.True(t, verify1)
	assert.True(t, verify2)
}

func TestVerifyFail(t *testing.T) {
	prv := NewPrivateKey()
	pub := prv.PublicKey
	hash := hash256([]byte("Москвичи смогут увидеть верхнее соединение Меркурия с Солнцем"))

	sign := prv.Sign(hash)
	sign[0]++
	verify := pub.Verify(hash, sign)

	assert.False(t, verify)
}

func TestRecoverPublicKey(t *testing.T) {
	prv := NewPrivateKey()
	pub1 := prv.PublicKey
	hash := hash256([]byte("Test text"))
	sign := prv.Sign(hash)

	fmt.Printf("PUB: %x\n", pub1.bytes())
	pub2, err := RecoverPublicKey(hash, sign)
	verify := pub2.Verify(hash, sign)

	assert.NoError(t, err)
	assert.True(t, pub1.Equal(pub2))
	assert.True(t, verify)
}
