package db

import (
	"testing"
	"time"

	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/blockchain/transaction"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/object"
	"github.com/likecoin-pro/likecoin/tests"
	"github.com/stretchr/testify/assert"
)

func TestBlockchainStorage_PutBlock(t *testing.T) {
	c := newContext()
	defer c.Drop()

	// make block-1
	b1, _ := c.NewBlock(
		newTestEmission(tests.AliceAddr, +100, coin),
		newTestTransfer(tests.AliceKey, tests.BobAddr, 10, coin, "", 0),
		newTestTransfer(tests.AliceKey, tests.CatAddr, 20, coin, "", 0),
	)
	err1 := c.PutBlock(b1, true) // ok
	assert.NoError(t, err1)

	// make block-2
	b2, _ := c.NewBlock(
		newTestTransfer(tests.AliceKey, tests.CatAddr, 5, coin, "", 0),
	)
	err2 := c.PutBlock(b2, true) // ok
	assert.NoError(t, err2)

	// set incorrect state
	c.state.Increment(tests.Coin, tests.AliceAddr, state.Int(1), 0)

	// make invalid block-3
	b3, _ := c.NewBlock(
		newTestTransfer(tests.AliceKey, tests.CatAddr, 10, coin, "", 0),
	)
	err3 := c.PutBlock(b3, true) // fail
	assert.Error(t, err3)
}

func TestBlockchainStorage_PutBlock_fail(t *testing.T) {
	c := newContext()
	defer c.Drop()

	tx1 := newTestEmission(tests.AliceAddr, +100, coin)
	tx2 := newTestEmission(tests.AliceAddr, +99, coin)

	b0 := blockchain.GenesisBlock()

	// put block#1
	b1 := b0.NewBlock()
	b1.AddTx(tx1, c.state, c.stateTree)
	b1.SetSign(tests.MasterKey, c.chainTree)
	err1 := c.PutBlock(b1, true)
	assert.NoError(t, err1) // ok

	// make second block#1
	b2 := b0.NewBlock()
	b2.AddTx(tx2, c.state, c.stateTree)
	b2.SetSign(tests.MasterKey, c.chainTree)
	err2 := c.PutBlock(b2, true)
	assert.Error(t, err2) // fail

	// make block#2 with duplicate tx1
	b2 = b1.NewBlock()
	b2.AddTx(tx1, c.state, c.stateTree) // duplicate tx
	b2.SetSign(tests.MasterKey, c.chainTree)
	err3 := c.PutBlock(b2, true) // fail
	assert.Error(t, err3)

	// make block#2 with invalid sign
	b2 = b1.NewBlock()
	b2.AddTx(tx2, c.state, c.stateTree)     // ok
	b2.SetSign(tests.AliceKey, c.chainTree) // but bad sign
	err4 := c.PutBlock(b2, true)            // fail
	assert.Error(t, err4)
}

func TestBlockchainStorage_GetBlock(t *testing.T) {
	c := newContext()
	defer c.Drop()

	// make block-1
	b1, _ := c.AddBlock(
		newTestEmission(tests.AliceAddr, +100, coin),
		newTestTransfer(tests.AliceKey, tests.BobAddr, 10, coin, "", 0),
		newTestTransfer(tests.AliceKey, tests.CatAddr, 20, coin, "", 0),
	)

	// make block-2
	b2, _ := c.AddBlock(
		newTestTransfer(tests.AliceKey, tests.CatAddr, 5, coin, "", 0),
	)

	B0, err0 := c.GetBlock(0) // ok - genesis block
	B1, err1 := c.GetBlock(1) // ok
	B2, err2 := c.GetBlock(2) // ok
	B3, err3 := c.GetBlock(3) // empty

	assert.NoError(t, err0)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Error(t, err3)

	assert.Equal(t, blockchain.GenesisBlock(), B0)
	assert.Nil(t, B3)
	assert.JSONEq(t, enc.JSON(b1), enc.JSON(B1))
	assert.JSONEq(t, enc.JSON(b2), enc.JSON(B2))
	assert.Equal(t, b1.Encode(), B1.Encode())
	assert.Equal(t, b2.Encode(), B2.Encode())
}

func TestBlockchainStorage_TxByID(t *testing.T) {
	c := newContext()
	defer c.Drop()

	tx1 := newTestEmission(tests.AliceAddr, +100, coin)
	tx2 := newTestTransfer(tests.AliceKey, tests.BobAddr, 10, coin, "", 0)

	txID1 := transaction.TxID(tx1)
	txID2 := transaction.TxID(tx2)

	// make block#1
	c.AddBlock(tx1)

	// make block#2
	c.AddBlock(tx2)

	// get transactions by txID
	it0, err0 := c.TransactionByID(uint64(time.Now().UnixNano())) // fail
	it1, err1 := c.TransactionByID(txID1)                         // ok
	it2, err2 := c.TransactionByID(txID2)                         // ok

	assert.Error(t, err0)
	assert.Nil(t, it0)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.JSONEq(t, enc.JSON(tx1), enc.JSON(it1.Tx))
	assert.JSONEq(t, enc.JSON(tx2), enc.JSON(it2.Tx))
	assert.Equal(t, tx1.Encode(), it1.Tx.Encode())
	assert.Equal(t, tx2.Encode(), it2.Tx.Encode())
}

func TestBlockchainStorage_FetchTxUID(t *testing.T) {
	c := newContext()
	defer c.Drop()

	tx1 := newTestEmission(tests.AliceAddr, +100, coin)
	tx2 := newTestTransfer(tests.AliceKey, tests.BobAddr, 10, coin, "", 0)
	tx3 := newTestTransfer(tests.AliceKey, tests.CatAddr, 20, coin, "", 0)

	// make block#1
	b1, _ := c.AddBlock(tx1)
	it1 := b1.Items[0]

	// make block#2
	b2, _ := c.AddBlock(tx2, tx3)
	it2 := b2.Items[0]
	it3 := b2.Items[1]

	// get transaction-UIDs by address
	var txUIDs []uint64
	err := c.fetchTxUID(tests.Coin, tests.AliceAddr, 0, 0, 0, false,
		func(txUID uint64, val state.Number) error {
			txUIDs = append(txUIDs, txUID)
			return nil
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, []uint64{it1.UID(), it2.UID(), it3.UID()}, txUIDs)
}

func TestBlockchainStorage_FetchTransaction(t *testing.T) {
	c := newContext()
	defer c.Drop()

	tx1 := newTestEmission(tests.AliceAddr, +100, coin)
	tx2 := newTestTransfer(tests.AliceKey, tests.BobAddr, 10, coin, "", 0)
	tx3 := newTestTransfer(tests.AliceKey, tests.CatAddr, 20, coin, "", 0)

	c.AddBlock(tx1)      // make block#1
	c.AddBlock(tx2, tx3) // make block#2

	// get transactions by address
	var txs []transaction.Transaction
	err := c.FetchTransactions(tests.Coin, tests.AliceAddr, 0, 0, 0, false,
		func(tx *blockchain.BlockItem, val state.Number) error {
			txs = append(txs, tx.Tx)
			return nil
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, []transaction.Transaction{tx1, tx2, tx3}, txs)
}

func TestBlockchainStorage_FetchTransactionByTag(t *testing.T) {
	c := newContext()
	defer c.Drop()

	tx1 := newTestEmission(tests.AliceAddr, +100, coin)
	tx2 := newTestTransfer(tests.AliceKey, tests.BobAddr, 10, coin, "", 222)
	tx3 := newTestTransfer(tests.AliceKey, tests.CatAddr, 20, coin, "", 333)
	tx4 := newTestTransfer(tests.AliceKey, tests.BobAddr, 30, coin, "", 222)

	c.AddBlock(tx1, tx2, tx3, tx4)

	// get transactions by address and by tag <222>
	var txs []transaction.Transaction
	err := c.FetchTransactions(tests.Coin, tests.AliceAddr, 222, 0, 0, false, func(tx *blockchain.BlockItem, _ state.Number) error {
		txs = append(txs, tx.Tx)
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, []transaction.Transaction{tx2, tx4}, txs)
}

func TestBlockchainStorage_NameOwner(t *testing.T) {
	c := newContext()
	defer c.Drop()

	aliceName := assets.NewName("@alice")

	// emission name
	c.AddBlock(
		newTestEmission(tests.AliceAddr, +1, aliceName),
	)

	// get owner of name
	addr1, txUID1, err1 := c.NameAddress("@alice")

	// transfer name
	c.AddBlock(
		newTestTransfer(tests.AliceKey, tests.BobAddr, 1, aliceName, "transfer @alice-name to Bob", 0),
	)

	// get new owner of name
	addr2, txUID2, err2 := c.NameAddress("@alice")

	// second transfer name
	c.AddBlock(
		newTestTransfer(tests.BobKey, tests.CatAddr, 1, aliceName, "transfer @alice-name to Cat", 0),
	)

	// get new owner of name
	addr3, txUID3, err3 := c.NameAddress("@alice")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.Equal(t, addr1, tests.AliceAddr)
	assert.Equal(t, addr2, tests.BobAddr)
	assert.Equal(t, addr3, tests.CatAddr)
	assert.Equal(t, uint64(0x100000000), txUID1)
	assert.Equal(t, uint64(0x200000000), txUID2)
	assert.Equal(t, uint64(0x300000000), txUID3)
}

func TestBlockchainStorage_UserByID(t *testing.T) {
	c := newContext()
	defer c.Drop()

	// register users
	_, err := c.AddBlock(
		object.NewUser(tests.AliceKey, "alice", config.MasterPublicKey.ID(), nil),
		object.NewUser(tests.BobKey, "bob", tests.AliceID, nil),
		object.NewUser(tests.CatKey, "cat", tests.AliceID, nil),
	)
	assert.NoError(t, err)

	// get user by ID
	tx, user, err := c.UserByID(tests.AliceID)

	assert.NoError(t, err)
	assert.Equal(t, "alice", user.Nick)
	assert.Equal(t, tests.AliceID, user.UserID())
	assert.Equal(t, tests.AliceAddr, user.Address())
	assert.Equal(t, uint64(0x100000000), tx.UID())
}

func TestBlockchainStorage_UserByNick(t *testing.T) {
	c := newContext()
	defer c.Drop()

	// register users
	c.AddBlock(
		object.NewUser(tests.AliceKey, "alice", config.MasterPublicKey.ID(), nil),
		object.NewUser(tests.BobKey, "bob", tests.AliceID, nil),
		object.NewUser(tests.CatKey, "cat", tests.AliceID, nil),
	)

	// get user by name
	tx, user, err := c.UserByNick("bob")

	assert.NoError(t, err)
	assert.Equal(t, "bob", user.Nick)
	assert.Equal(t, tests.BobID, user.UserID())
	assert.Equal(t, tests.BobAddr, user.Address())
	assert.Equal(t, uint64(0x100000001), tx.UID())
}
