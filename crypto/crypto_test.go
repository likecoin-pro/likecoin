package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSign_Deterministic(t *testing.T) {
	prv := NewPrivateKey()
	hash := HashSum256([]byte("Лавров выразил надежду, что «ловцы покемонов» не разучатся говорить"))

	sign1 := prv.Sign(hash)
	sign2 := prv.Sign(hash)

	assert.Equal(t, signatureSize, len(sign1))
	assert.Equal(t, signatureSize, len(sign2))
	assert.True(t, bytes.Equal(sign1, sign2))
}

func TestVerify(t *testing.T) {
	prv := NewPrivateKey()
	pub := prv.PublicKey
	hash := HashSum256([]byte("Совет по туризму Норвегии определил места обитания редких покемонов"))
	sign := prv.Sign(hash)

	verify1 := pub.Verify(hash, sign)
	verify2 := pub.Verify(hash, sign)

	assert.True(t, verify1)
	assert.True(t, verify2)
}

func TestVerifyFail(t *testing.T) {
	prv := NewPrivateKey()
	pub := prv.PublicKey
	hash := HashSum256([]byte("Москвичи смогут увидеть верхнее соединение Меркурия с Солнцем"))

	sign := prv.Sign(hash)
	sign[13]++ // corrupt signature
	verify := pub.Verify(hash, sign)

	assert.False(t, verify)
}
