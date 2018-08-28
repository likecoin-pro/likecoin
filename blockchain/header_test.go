package blockchain

import (
	"testing"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/stretchr/testify/assert"
)

func TestBlockHeader_Decode(t *testing.T) {
	block1 := &BlockHeader{
		Version:   0,
		Network:   1,
		ChainID:   1,
		Num:       123,
		Timestamp: 1500000000e6,
		PrevHash:  []byte("0123456789abcdef0123456789abcdef"),
		TxRoot:    []byte("0123456789abcdef0123456789abcdef"),
		StateRoot: []byte("0123456789abcdef0123456789abcdef"),
		ChainRoot: []byte("0123456789abcdef0123456789abcdef"),
		Nonce:     0x123456789,
		Miner:     crypto.MustParsePublicKey("rggH2X4N7JsBtekY1isiutwJZpRhQQoeYaVqYdRSH4mR"),
		Sig:       []byte("0123456789abcdef0123456789abcdef"),
	}
	data := bin.Encode(block1)

	var block2 *BlockHeader
	err := bin.Decode(data, &block2)

	assert.NoError(t, err)
	assert.EqualValues(t, block1, block2)
	assert.Equal(t, block1.Hash(), block2.Hash())
	assert.JSONEq(t, enc.JSON(block1), enc.JSON(block2))
}
