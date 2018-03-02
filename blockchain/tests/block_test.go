package tests

import (
	"testing"

	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/stretchr/testify/assert"
)

func TestBlock_Encode(t *testing.T) {
	st := state.NewState()
	b := blockchain.GenesisBlock().NewBlock()
	b.AddTx(st, NewTestEmission(AliceAddr, 100, Coin)).UID()
	b.AddTx(st, NewTestTransaction(AliceKey, BobAddr, 5, Coin, "transfer-1 to Bob", 0))
	b.AddTx(st, NewTestTransaction(AliceKey, BobAddr, 5, Coin, "transfer-2 to Bob", 0))
	b.AddTx(st, NewTestTransaction(AliceKey, CatAddr, 5, Coin, "transfer-3 to Cat", 0))
	b.SetSign(MasterKey)
	data := b.Encode()

	var B = new(blockchain.Block)
	err := B.Decode(data)
	assert.NoError(t, err)
	assert.JSONEq(t, enc.JSON(b), enc.JSON(B))
}

func TestBlockItem_UID(t *testing.T) {
	st := state.NewState()
	b1 := blockchain.GenesisBlock().NewBlock()
	txUID1 := b1.AddTx(st, NewTestEmission(AliceAddr, 100, Coin)).UID()
	txUID2 := b1.AddTx(st, NewTestTransaction(AliceKey, BobAddr, 5, Coin, "transfer-1 to Bob", 0)).UID()
	txUID3 := b1.AddTx(st, NewTestTransaction(AliceKey, BobAddr, 5, Coin, "transfer-2 to Bob", 0)).UID()
	txUID4 := b1.AddTx(st, NewTestTransaction(AliceKey, CatAddr, 5, Coin, "transfer-3 to Cat", 0)).UID()

	assert.EqualValues(t, 0x100000000, txUID1)
	assert.EqualValues(t, 0x100000001, txUID2)
	assert.EqualValues(t, 0x100000002, txUID3)
	assert.EqualValues(t, 0x100000003, txUID4)
}

func TestBlock_AddTx(t *testing.T) {
	st := state.NewState()

	err := st.Execute(func() {
		b1 := blockchain.GenesisBlock().NewBlock()
		b1.AddTx(st, NewTestEmission(AliceAddr, 100, Coin))
		b1.AddTx(st, NewTestTransaction(AliceKey, BobAddr, 5, Coin, "transfer-1 to Bob", 0))
		b1.AddTx(st, NewTestTransaction(AliceKey, BobAddr, 5, Coin, "transfer-2 to Bob", 0))
		b1.AddTx(st, NewTestTransaction(AliceKey, CatAddr, 5, Coin, "transfer-3 to Cat", 0))
		b1.SetSign(MasterKey)
	})

	assert.NoError(t, err)
	assert.JSONEq(t, `[
		{"address":"Like5A2PEu6eCHQzy1tMsa6b3kc1xXS7ywj2NQZr8xL","asset":"0001","value":85},
		{"address":"Like4fGoCMKi9LNqBbAdG3ppFuWRmGDM5bqSsQq9b37","asset":"0001","value":10},
		{"address":"Like3ssGc6gvkhbpveJRSPLxK8ZnKP7HEbhDXfwJfvk","asset":"0001","value":5}
	]`, enc.JSON(st))
}

func TestBlock_AddTx_failNotEnough(t *testing.T) {
	st := state.NewState()

	err := st.Execute(func() {
		b1 := blockchain.GenesisBlock().NewBlock()
		b1.AddTx(st, NewTestEmission(AliceAddr, 100, Coin))
		b1.AddTx(st, NewTestTransaction(AliceKey, BobAddr, 55, Coin, "transfer-1 to Bob", 0))
		b1.AddTx(st, NewTestTransaction(AliceKey, BobAddr, 55, Coin, "transfer-2 to Bob", 0))
		b1.SetSign(MasterKey)
	})

	assert.Error(t, err)
	assert.Equal(t, state.ErrNegativeValue, err)
}

func TestBlock_AddTx_failVerify(t *testing.T) {
	st := state.NewState()

	err := st.Execute(func() {
		b1 := blockchain.GenesisBlock().NewBlock()
		tx := NewTestEmission(AliceAddr, 100, Coin)
		tx.Sign[0] = tx.Sign[0] - 1
		b1.AddTx(st, tx)
	})

	assert.Error(t, err)
	assert.Equal(t, blockchain.ErrInvalidSign, err)
}
