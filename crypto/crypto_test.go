package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePrivateKey(t *testing.T) {

	prv1 := NewPrivateKey()
	prv2 := NewPrivateKey()
	pub1 := prv1.PublicKey

	assert.Equal(t, 33, len(prv1.Encode()))
	assert.NotEqual(t, prv1, prv2)
	assert.Equal(t, PublicKeySize, len(pub1.Encode()))
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
	assert.Equal(t, "3VBoiiUF6eipdLiSJGFrh37MiQsyAQuwiWootfFXBgXZaRxj6Grm8YezWAFKivi737b7kz95Kc3aXkPeTcQKpk7W", pub1.String())
	assert.Equal(t, 33, len(prv1.Encode()))
	assert.Equal(t, prv1.Encode(), prv2.Encode())
	assert.Equal(t, pub1.Encode(), pub2.Encode())
	assert.NotEqual(t, prv1.Encode(), prv3.Encode())
	assert.NotEqual(t, pub1.Encode(), pub3.Encode())
}

func TestSign(t *testing.T) {
	prv := NewPrivateKey()
	data := []byte("Лавров выразил надежду, что «ловцы покемонов» не разучатся говорить")

	sign1 := prv.Sign(data)
	sign2 := prv.Sign(data)

	assert.Equal(t, PublicKeySize, len(sign1))
	assert.Equal(t, PublicKeySize, len(sign2))
	assert.False(t, bytes.Equal(sign1, sign2))
}

func TestVerify(t *testing.T) {
	prv := NewPrivateKey()
	pub := prv.PublicKey
	data := []byte("Совет по туризму Норвегии определил места обитания редких покемонов")
	sign := prv.Sign(data)

	verify1 := pub.Verify(data, sign)
	verify2 := pub.Verify(data, sign)

	assert.True(t, verify1)
	assert.True(t, verify2)
}

func TestVerifyFail(t *testing.T) {
	prv := NewPrivateKey()
	pub := prv.PublicKey
	data := []byte("Москвичи смогут увидеть верхнее соединение Меркурия с Солнцем")

	sign := prv.Sign(data)
	sign[0]++
	verify := pub.Verify(data, sign)

	assert.False(t, verify)
}
