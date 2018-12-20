package blockchain

import (
	"encoding/hex"
	"testing"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/stretchr/testify/assert"
)

var (
	testPrv = crypto.NewPrivateKeyBySecret("abc")
	testPub = testPrv.PublicKey
)

func TestTransaction_Encode(t *testing.T) {
	tx := &Transaction{
		Type:    1,
		Version: 0,
		Network: NetworkWorking,
		ChainID: 1,
		Nonce:   12345,
		Sender:  testPub,
		Data:    []byte("abcdefg"),
	}

	data := bin.Encode(tx)

	assert.Equal(t,
		"3501000001823039076162636465666700002102022f86f8c408c20e8bdcef6471676a2157624915355fe662b568ac5e2a2a76fe0000",
		hex.EncodeToString(data),
	)
}

func TestTransaction_Decode(t *testing.T) {
	data, _ := hex.DecodeString(
		"3501000001823039076162636465666700002102022f86f8c408c20e8bdcef6471676a2157624915355fe662b568ac5e2a2a76fe0000",
	)

	var tx *Transaction
	err := bin.Decode(data, &tx)

	assert.NoError(t, err)
	assert.EqualValues(t, &Transaction{
		Type:    1,
		Version: 0,
		Network: NetworkWorking,
		ChainID: 1,
		Nonce:   12345,
		Sender:  testPub,
		Data:    []byte("abcdefg"),
	}, tx)
}
