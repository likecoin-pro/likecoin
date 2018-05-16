package xhash

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXHash(t *testing.T) {
	key := XHash([]byte("abc"))

	assert.Equal(t, "d85f68a0dd6b2ebeb5a60b47b70d2e4b63a842ac0510116e2d52f153e535cd60635f76fd52b34cceb671e0ed11093e923c39ee1a5ff32088ebf5f2415a285eef", hex.EncodeToString(key))
}

func TestGenerateKey(t *testing.T) {
	key := GenerateKeyByPassword("abc", 512)

	assert.Equal(t, "d85f68a0dd6b2ebeb5a60b47b70d2e4b63a842ac0510116e2d52f153e535cd60635f76fd52b34cceb671e0ed11093e923c39ee1a5ff32088ebf5f2415a285eef", hex.EncodeToString(key))
}
