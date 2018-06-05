package blockchain

import (
	"testing"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_Decode(t *testing.T) {
	tx1 := &Transaction{
		Type:    1,
		Version: 0,
		Network: config.NetworkID,
		ChainID: config.ChainID,
		Nonce:   123456789,
		Sender:  crypto.NewPrivateKey().PublicKey,
		Data:    []byte("0123456789abcdefg"),
	}
	data := bin.Encode(tx1)

	var tx2 *Transaction
	err := bin.Decode(data, &tx2)

	assert.NoError(t, err)
	assert.EqualValues(t, tx1, tx2)
}
