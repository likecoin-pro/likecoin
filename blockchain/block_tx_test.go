package blockchain

import (
	"testing"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/stretchr/testify/assert"
)

func TestBlockTx_Decode(t *testing.T) {

	prv := crypto.NewPrivateKey()
	tx1 := NewBlockTx(prv, &TestTxObject{Msg: "test-msg"})
	data := bin.Encode(tx1)

	var tx2 *BlockTx
	err := bin.Decode(data, &tx2)

	assert.NoError(t, err)
	assert.EqualValues(t, tx1, tx2)
}
