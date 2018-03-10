package db

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/blockchain/tests"
	"github.com/likecoin-pro/likecoin/blockchain/transaction"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/object"
	"github.com/stretchr/testify/assert"
)

func newTestBC() (*state.State, *BlockchainStorage) {
	config.VerifyTransactions = true
	config.NetworkID = 1 // test
	config.ChainID = 1   //
	bc := NewBlockchainStorage(config.NetworkID, config.ChainID, fmt.Sprintf("%s/test-db-%x", os.TempDir(), rand.Uint64()))
	return state.NewState(config.ChainID, nil), bc
}

func TestBlockchainStorage_PutBlock(t *testing.T) {
	st, bc := newTestBC()
	defer bc.Drop()

	// make block-1
	b1 := blockchain.GenesisBlock().NewBlock()
	b1.AddTx(st, tests.NewTestEmission(tests.AliceAddr, +100, tests.Coin))
	b1.AddTx(st, tests.NewTestTransaction(tests.AliceKey, tests.BobAddr, 10, tests.Coin, "", 0))
	b1.AddTx(st, tests.NewTestTransaction(tests.AliceKey, tests.CatAddr, 20, tests.Coin, "", 0))
	b1.SetSign(tests.MasterKey)
	err1 := bc.PutBlock(b1) // ok
	assert.NoError(t, err1)

	// make block-2
	b2 := b1.NewBlock()
	b2.AddTx(st, tests.NewTestTransaction(tests.AliceKey, tests.CatAddr, 5, tests.Coin, "", 0))
	b2.SetSign(tests.MasterKey)
	err2 := bc.PutBlock(b2) // ok
	assert.NoError(t, err2)

	// set incorrect state
	st.Increment(tests.Coin, tests.AliceAddr, state.Int(1), 0)

	// make invalid block-3
	b3 := b2.NewBlock()
	b3.AddTx(st, tests.NewTestTransaction(tests.AliceKey, tests.CatAddr, 10, tests.Coin, "", 0))
	b3.SetSign(tests.MasterKey)
	err3 := bc.PutBlock(b3) // fail
	assert.Error(t, err3)
}

func TestBlockchainStorage_PutBlock_fail(t *testing.T) {
	st, bc := newTestBC()
	defer bc.Drop()

	tx1 := tests.NewTestEmission(tests.AliceAddr, +100, tests.Coin)
	tx2 := tests.NewTestEmission(tests.AliceAddr, +99, tests.Coin)

	// put block#1
	b1 := blockchain.GenesisBlock().NewBlock()
	b1.AddTx(st, tx1)
	b1.SetSign(tests.MasterKey)
	err1 := bc.PutBlock(b1)
	assert.NoError(t, err1) // ok

	// make second block#1
	b2 := blockchain.GenesisBlock().NewBlock()
	b2.AddTx(st, tx2)
	b2.SetSign(tests.MasterKey)
	err2 := bc.PutBlock(b2)
	assert.Error(t, err2) // fail

	// make block#2 with duplicate tx1
	b2 = b1.NewBlock()
	b2.AddTx(st, tx1) // duplicate tx
	b2.SetSign(tests.MasterKey)
	err3 := bc.PutBlock(b2) // fail
	assert.Error(t, err3)

	// make block#2 with invalid sign
	b2 = b1.NewBlock()
	b2.AddTx(st, tx2)          // ok
	b2.SetSign(tests.AliceKey) // but bad sign
	err4 := bc.PutBlock(b2)    // fail
	assert.Error(t, err4)
}

func TestBlockchainStorage_GetBlock(t *testing.T) {
	st, bc := newTestBC()
	defer bc.Drop()

	// make block-1
	b1 := blockchain.GenesisBlock().NewBlock()
	b1.AddTx(st, tests.NewTestEmission(tests.AliceAddr, +100, tests.Coin))
	b1.AddTx(st, tests.NewTestTransaction(tests.AliceKey, tests.BobAddr, 10, tests.Coin, "", 0))
	b1.AddTx(st, tests.NewTestTransaction(tests.AliceKey, tests.CatAddr, 20, tests.Coin, "", 0))
	b1.SetSign(tests.MasterKey)
	bc.PutBlock(b1)

	// make block-2
	b2 := b1.NewBlock()
	b2.AddTx(st, tests.NewTestTransaction(tests.AliceKey, tests.CatAddr, 5, tests.Coin, "", 0))
	b2.SetSign(tests.MasterKey)
	bc.PutBlock(b2)

	B0, err0 := bc.GetBlock(0) // ok - genesis block
	B1, err1 := bc.GetBlock(1) // ok
	B2, err2 := bc.GetBlock(2) // ok
	B3, err3 := bc.GetBlock(3) // empty

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
	st, bc := newTestBC()
	defer bc.Drop()

	tx1 := tests.NewTestEmission(tests.AliceAddr, +100, tests.Coin)
	tx2 := tests.NewTestTransaction(tests.AliceKey, tests.BobAddr, 10, tests.Coin, "", 0)

	txID1 := transaction.TxID(tx1)
	txID2 := transaction.TxID(tx2)

	// make block#1
	b1 := blockchain.GenesisBlock().NewBlock()
	b1.AddTx(st, tx1)
	b1.SetSign(tests.MasterKey)
	bc.PutBlock(b1)

	// make block#2
	b2 := b1.NewBlock()
	b2.AddTx(st, tx2)
	b2.SetSign(tests.MasterKey)
	bc.PutBlock(b2)

	// get transactions by txID
	it0, err0 := bc.TransactionByID(uint64(time.Now().UnixNano())) // fail
	it1, err1 := bc.TransactionByID(txID1)                         // ok
	it2, err2 := bc.TransactionByID(txID2)                         // ok

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
	st, bc := newTestBC()
	defer bc.Drop()

	tx1 := tests.NewTestEmission(tests.AliceAddr, +100, tests.Coin)
	tx2 := tests.NewTestTransaction(tests.AliceKey, tests.BobAddr, 10, tests.Coin, "", 0)
	tx3 := tests.NewTestTransaction(tests.AliceKey, tests.CatAddr, 20, tests.Coin, "", 0)

	// make block#1
	b1 := blockchain.GenesisBlock().NewBlock()
	it1, _ := b1.AddTx(st, tx1)
	b1.SetSign(tests.MasterKey)
	bc.PutBlock(b1)

	// make block#2
	b2 := b1.NewBlock()
	it2, _ := b2.AddTx(st, tx2)
	it3, _ := b2.AddTx(st, tx3)
	b2.SetSign(tests.MasterKey)
	bc.PutBlock(b2)

	// get transaction-UIDs by address
	var txUIDs []uint64
	err := bc.fetchTxUID(tests.Coin, tests.AliceAddr, 0, 0, 0, false,
		func(txUID uint64, val state.Number) error {
			txUIDs = append(txUIDs, txUID)
			return nil
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, []uint64{it1.UID(), it2.UID(), it3.UID()}, txUIDs)
}

func TestBlockchainStorage_FetchTransaction(t *testing.T) {
	st, bc := newTestBC()
	defer bc.Drop()

	tx1 := tests.NewTestEmission(tests.AliceAddr, +100, tests.Coin)
	tx2 := tests.NewTestTransaction(tests.AliceKey, tests.BobAddr, 10, tests.Coin, "", 0)
	tx3 := tests.NewTestTransaction(tests.AliceKey, tests.CatAddr, 20, tests.Coin, "", 0)

	// make block#1
	b1 := blockchain.GenesisBlock().NewBlock()
	b1.AddTx(st, tx1)
	b1.SetSign(tests.MasterKey)
	bc.PutBlock(b1)

	// make block#2
	b2 := b1.NewBlock()
	b2.AddTx(st, tx2)
	b2.AddTx(st, tx3)
	b2.SetSign(tests.MasterKey)
	bc.PutBlock(b2)

	// get transactions by address
	var txs []transaction.Transaction
	err := bc.FetchTransactions(tests.Coin, tests.AliceAddr, 0, 0, 0, false,
		func(tx *blockchain.BlockItem, val state.Number) error {
			txs = append(txs, tx.Tx)
			return nil
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, []transaction.Transaction{tx1, tx2, tx3}, txs)
}

func TestBlockchainStorage_FetchTransactionByTag(t *testing.T) {
	st, bc := newTestBC()
	defer bc.Drop()

	tx1 := tests.NewTestEmission(tests.AliceAddr, +100, tests.Coin)
	tx2 := tests.NewTestTransaction(tests.AliceKey, tests.BobAddr, 10, tests.Coin, "", 222)
	tx3 := tests.NewTestTransaction(tests.AliceKey, tests.CatAddr, 20, tests.Coin, "", 333)
	tx4 := tests.NewTestTransaction(tests.AliceKey, tests.BobAddr, 30, tests.Coin, "", 222)

	b1 := blockchain.GenesisBlock().NewBlock()
	b1.AddTx(st, tx1)
	b1.AddTx(st, tx2)
	b1.AddTx(st, tx3)
	b1.AddTx(st, tx4)
	b1.SetSign(tests.MasterKey)
	bc.PutBlock(b1)

	// get transactions by address and by tag <222>
	var txs []transaction.Transaction
	err := bc.FetchTransactions(tests.Coin, tests.AliceAddr, 222, 0, 0, false, func(tx *blockchain.BlockItem, _ state.Number) error {
		txs = append(txs, tx.Tx)
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, []transaction.Transaction{tx2, tx4}, txs)
}

func TestBlockchainStorage_NameOwner(t *testing.T) {
	st, bc := newTestBC()
	defer bc.Drop()

	aliceName := assets.NewName("@alice")

	// emission name
	b := blockchain.GenesisBlock().NewBlock()
	b.AddTx(st, tests.NewTestEmission(tests.AliceAddr, +1, aliceName))
	b.SetSign(tests.MasterKey)
	bc.PutBlock(b)

	// get owner of name
	addr1, txUID1, err1 := bc.NameAddress("@alice")

	// transfer name
	b = b.NewBlock()
	b.AddTx(st, tests.NewTestTransaction(tests.AliceKey, tests.BobAddr, 1, aliceName, "transfer @alice-name to Bob", 0))
	b.SetSign(tests.MasterKey)
	bc.PutBlock(b)

	// get new owner of name
	addr2, txUID2, err2 := bc.NameAddress("@alice")

	// second transfer name
	b = b.NewBlock()
	b.AddTx(st, tests.NewTestTransaction(tests.BobKey, tests.CatAddr, 1, aliceName, "transfer @alice-name to Cat", 0))
	b.SetSign(tests.MasterKey)
	bc.PutBlock(b)

	// get new owner of name
	addr3, txUID3, err3 := bc.NameAddress("@alice")

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
	st, bc := newTestBC()
	defer bc.Drop()

	// register users
	b := blockchain.GenesisBlock().NewBlock()
	b.AddTx(st, object.NewUser(tests.AliceKey, "alice", config.MasterPublicKey.ID(), nil))
	b.AddTx(st, object.NewUser(tests.BobKey, "bob", tests.AliceID, nil))
	b.AddTx(st, object.NewUser(tests.CatKey, "cat", tests.AliceID, nil))
	b.SetSign(tests.MasterKey)
	err := bc.PutBlock(b)
	assert.NoError(t, err)

	// get user by ID
	tx, user, err := bc.UserByID(tests.AliceID)

	assert.NoError(t, err)
	assert.Equal(t, "alice", user.Nick)
	assert.Equal(t, tests.AliceID, user.ID())
	assert.Equal(t, tests.AliceAddr, user.Address())
	assert.Equal(t, uint64(0x100000000), tx.UID())
}

func TestBlockchainStorage_UserByNick(t *testing.T) {
	st, bc := newTestBC()
	defer bc.Drop()

	// register users
	b := blockchain.GenesisBlock().NewBlock()
	b.AddTx(st, object.NewUser(tests.AliceKey, "alice", config.MasterPublicKey.ID(), nil))
	b.AddTx(st, object.NewUser(tests.BobKey, "bob", tests.AliceID, nil))
	b.AddTx(st, object.NewUser(tests.CatKey, "cat", tests.AliceID, nil))
	b.SetSign(tests.MasterKey)
	bc.PutBlock(b)

	// get user by name
	tx, user, err := bc.UserByNick("bob")

	assert.NoError(t, err)
	assert.Equal(t, "bob", user.Nick)
	assert.Equal(t, tests.BobID, user.ID())
	assert.Equal(t, tests.BobAddr, user.Address())
	assert.Equal(t, uint64(0x100000001), tx.UID())
}
