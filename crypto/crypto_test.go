package crypto

import (
	"bytes"
	"testing"

	"github.com/likecoin-pro/likecoin/crypto/merkle"
	"github.com/stretchr/testify/assert"
)

func Test_data(t *testing.T) {
	prv := NewPrivateKeyBySecret("alice::Alice secret")
	pub := prv.PublicKey
	addr := pub.Address()
	addrHex := addr.Hex()

	assert.Equal(t, "0x01fd5a70cc093eca84e98cd8b0ba9c4670aecc695b20b09ad0918ab4922a3fd7ad", prv.String())
	assert.Equal(t, "y32va5HWnYeZpqg4TJ2ydFiauspUUwZ4bswwz6NMDTNq", pub.String())
	assert.Equal(t, "0x4093cdf68e4fbeea9307530b20138fd56675f386a4eb0daa1f8067435e4eef9ac29042e58725dbb9247699460ede385c80ceac2409eeadd88ab329496875ae0d", pub.Hex())
	assert.Equal(t, "Like62D4Rq3s8D4Y5Q92YBoRiVpcMFEcXTyGLbqrtAv", addr.String())
	assert.Equal(t, "0xe8284b25224d6eac303d80f493774cb8369e9eff304ef786", addrHex)
}

func TestSign_Deterministic(t *testing.T) {
	prv := NewPrivateKey()
	hash := HashSum256([]byte("Лавров выразил надежду, что «ловцы покемонов» не разучатся говорить"))

	sig1 := prv.Sign(hash)
	sig2 := prv.Sign(hash)

	assert.Equal(t, signatureSize, len(sig1))
	assert.Equal(t, signatureSize, len(sig2))
	assert.True(t, bytes.Equal(sig1, sig2))
}

func TestVerify(t *testing.T) {
	prv := NewPrivateKey()
	pub := prv.PublicKey
	hash := HashSum256([]byte("Совет по туризму Норвегии определил места обитания редких покемонов"))
	sig := prv.Sign(hash)

	verify1 := pub.Verify(hash, sig)
	verify2 := pub.Verify(hash, sig)

	assert.True(t, verify1)
	assert.True(t, verify2)
}

func TestVerifyFail(t *testing.T) {
	prv := NewPrivateKey()
	pub := prv.PublicKey
	hash := HashSum256([]byte("Москвичи смогут увидеть верхнее соединение Меркурия с Солнцем"))

	sig := prv.Sign(hash)
	sig[13]++ // corrupt signature
	verify := pub.Verify(hash, sig)

	assert.False(t, verify)
}

func TestVerify_withMerkleProof(t *testing.T) {
	prv := NewPrivateKey()
	pub := prv.PublicKey
	hashes := make([][]byte, 66)
	for i := range hashes {
		hashes[i] = Hash256(i)
	}
	hash := hashes[55]
	proof, root := merkle.Proof(hashes, 55)
	sig := prv.Sign(root)
	sig = append(proof, sig...) // append merkle-proof to signature

	verify := pub.Verify(hash, sig)

	assert.True(t, verify)
}
