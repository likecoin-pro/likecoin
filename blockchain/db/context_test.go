package db

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/blockchain/transaction"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/likecoin-pro/likecoin/crypto/patricia"
	"github.com/likecoin-pro/likecoin/object"
	"github.com/likecoin-pro/likecoin/tests"
)

type Context struct {
	*BlockchainStorage
	state     *state.State
	stateTree *patricia.Tree
	chainTree *patricia.Tree
	block     *blockchain.Block
}

func newContext() *Context {
	config.VerifyTransactions = true
	config.NetworkID = 1 // test
	config.ChainID = 1   //
	bc := NewBlockchainStorage(config.ChainID, fmt.Sprintf("%s/test-db-%x", os.TempDir(), rand.Uint64()))
	return &Context{
		BlockchainStorage: bc,
		state:             state.NewState(config.ChainID, nil),
		stateTree:         bc.StateTree(),
		chainTree:         bc.ChainTree(),
		block:             blockchain.GenesisBlock(),
	}
}

func (c *Context) Drop() error {
	return c.BlockchainStorage.Drop()
}

func (c *Context) NewBlock(txs ...transaction.Transaction) (block *blockchain.Block, err error) {
	block = c.block.NewBlock()
	c.block = block
	for _, tx := range txs {
		if _, err = block.AddTx(tx, c.state, c.stateTree); err != nil {
			return
		}
	}
	block.SetSign(tests.MasterKey, c.chainTree)
	return
}

func (c *Context) AddBlock(txs ...transaction.Transaction) (block *blockchain.Block, err error) {
	if block, err = c.NewBlock(txs...); err != nil {
		return
	}
	err = c.BlockchainStorage.PutBlock(block, true)
	return
}

//------------------------------------------------------
var coin = tests.Coin

func newTestTransfer(
	from *crypto.PrivateKey,
	toAddr crypto.Address,
	amount int64,
	asset assets.Asset,
	comment string,
	tag int64,
) (tx *object.Transfer) {
	return object.NewSimpleTransfer(from, toAddr, tag, asset, state.Int(amount), comment)
}

func newTestEmission(toAddr crypto.Address, amount int64, asset assets.Asset) (tx *object.Emission) {
	return object.NewEmission(
		tests.MasterKey,
		asset,
		"Test emission",
		&object.EmissionOut{toAddr, amount, "testMediaAddr", 0},
	)
}
