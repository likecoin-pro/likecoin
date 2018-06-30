package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrivateKey_String(t *testing.T) {
	prv := NewPrivateKeyBySecret("abc")
	pub := prv.PublicKey
	addr := pub.Address()

	assert.Equal(t, "0x01d85f68a0dd6b2ebeb5a60b47b70d2e4b63a842ac0510116e2d52f153e535cd61", prv.String())
	assert.Equal(t, "0x022f86f8c408c20e8bdcef6471676a2157624915355fe662b568ac5e2a2a76fed5d34d4a184176a3e4a28bac7203a860510e363601f7c8f8657067173ed83f6e", pub.Hex())
	assert.Equal(t, "0x9a8a9d2b5766b5c3962f4dd301c01765bdc37a6387f24250", addr.Hex())
	assert.Equal(t, "Like5DuaVTk8KgpRh98xDvHvnpaAWxSoYh6uLRvyar5", addr.String())
}

func TestParsePrivateKey(t *testing.T) {
	prv, err := ParsePrivateKey("0x016c23f9a5fc3609c74a37e2fb7982653c97e39f00a8f700f99e3b770bb872bd8c")

	assert.NoError(t, err)
	assert.Equal(t, "0x016c23f9a5fc3609c74a37e2fb7982653c97e39f00a8f700f99e3b770bb872bd8c", prv.Hex())
	assert.Equal(t, "0xcca27aa571d2838209895c8151cff2ade07e56c0c62a64cd3c84ad73a4287141b207d20a99a5c3169f5f086c3bd05480cad1ad1359a7f01151ed2fd9b2c67601", prv.PublicKey.Hex())
}

func TestParsePrivateKey_oldFormat(t *testing.T) {
	prv, err := ParsePrivateKey("0x6c23f9a5fc3609c74a37e2fb7982653c97e39f00a8f700f99e3b770bb872bd8c")

	assert.NoError(t, err)
	assert.Equal(t, "0x016c23f9a5fc3609c74a37e2fb7982653c97e39f00a8f700f99e3b770bb872bd8c", prv.Hex())
	assert.Equal(t, "0xcca27aa571d2838209895c8151cff2ade07e56c0c62a64cd3c84ad73a4287141b207d20a99a5c3169f5f086c3bd05480cad1ad1359a7f01151ed2fd9b2c67601", prv.PublicKey.Hex())
}

func TestParsePrivateKey_fail(t *testing.T) {
	_, err := ParsePrivateKey("0x026c23f9a5fc3609c74a37e2fb7982653c97e39f00a8f700f99e3b770bb872bd8c")

	assert.Error(t, err)
}

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

	assert.Equal(t, "0x015491dbe82c98002178a5bd37f211c09bea225a9c2b7ccf0f52c4d4ee2db4e123", prv1.String())
	assert.Equal(t, "mcPnH3mnGzQmAmcWRJRLaQqYNVSKFzvaBUaC1BYKMEXS", pub1.String())
	assert.Equal(t, 33, len(prv1.Encode()))
	assert.Equal(t, prv1.Encode(), prv2.Encode())
	assert.Equal(t, pub1.Encode(), pub2.Encode())
	assert.NotEqual(t, prv1.Encode(), prv3.Encode())
	assert.NotEqual(t, pub1.Encode(), pub3.Encode())
}
