package tests

import (
	"testing"

	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/stretchr/testify/assert"
)

func TestTxEncode(t *testing.T) {
	tx := NewTestTransaction(AliceKey, BobAddr, +10, Coin, "Test tx#1", 0)

	enc := blockchain.TxEncode(tx)

	assert.True(t, len(enc) > 0)
}

func TestTxDecode(t *testing.T) {
	data := blockchain.TxEncode(NewTestTransaction(AliceKey, BobAddr, +10, Coin, "Test tx#1", 0))

	obj, err := blockchain.TxDecode(data)
	dec, ok := obj.(*TestTransaction)

	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, TestTransactionType, dec.Type())
	assert.True(t, dec.From.Equal(AliceKey.PublicKey))
	assert.Equal(t, AliceKey.PublicKey.String(), dec.From.String())
	assert.Equal(t, BobAddr.String(), dec.To.String())
	assert.Equal(t, "0001", dec.Asset.String())
	assert.Equal(t, int64(10), dec.Value)
	assert.Equal(t, "Test tx#1", dec.Comment)
	assert.True(t, dec.From.Verify(dec.Hash(), dec.Sign))
}
