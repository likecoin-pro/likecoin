package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrivateKey_Hex(t *testing.T) {
	prv := NewPrivateKeyBySecret("abc")
	pub := prv.PublicKey

	assert.Equal(t, "6c23f9a5fc3609c74a37e2fb7982653c97e39f00a8f700f99e3b770bb872bd8c", prv.Hex())
	assert.Equal(t, "cca27aa571d2838209895c8151cff2ade07e56c0c62a64cd3c84ad73a4287141b207d20a99a5c3169f5f086c3bd05480cad1ad1359a7f01151ed2fd9b2c67601", pub.Hex())
}

func TestParsePrivateKeyHex(t *testing.T) {
	prv, err := ParsePrivateKeyHex("6c23f9a5fc3609c74a37e2fb7982653c97e39f00a8f700f99e3b770bb872bd8c")
	pub := prv.PublicKey

	assert.NoError(t, err)
	assert.Equal(t, "6c23f9a5fc3609c74a37e2fb7982653c97e39f00a8f700f99e3b770bb872bd8c", prv.Hex())
	assert.Equal(t, "cca27aa571d2838209895c8151cff2ade07e56c0c62a64cd3c84ad73a4287141b207d20a99a5c3169f5f086c3bd05480cad1ad1359a7f01151ed2fd9b2c67601", pub.Hex())
}
