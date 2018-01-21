package xhash

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateKey(t *testing.T) {
	key := GenerateKeyByPassword("abc", 512)

	assert.Equal(t,
		"6c23f9a5fc3609c74a37e2fb7982653c97e39f00a8f700f99e3b770bb872bd8b9b819d546a2cf5a2aebc10a28f75886a76ccc8c4f1ec8999652c9bb31ec8c8a7",
		hex.EncodeToString(key),
	)
}
