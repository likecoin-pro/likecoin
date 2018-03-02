package tests

import (
	"fmt"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
)

type TestTransaction struct {
	Timestamp int64             `json:"timestamp"`
	From      *crypto.PublicKey `json:"from"`
	To        crypto.Address    `json:"to"`
	Asset     assets.Asset      `json:"asset"`
	Value     int64             `json:"value"`
	Comment   string            `json:"comment"`
	Tag       int64             `json:"tag"`
	Sign      bin.Bytes         `json:"sign"`
}

const TestTransactionType blockchain.TxType = 255

var _ = blockchain.RegisterTransactionType(&TestTransaction{})

func NewTestTransaction(
	from *crypto.PrivateKey,
	toAddr crypto.Address,
	value int64,
	asset assets.Asset,
	comment string,
	tag int64,
) (tx *TestTransaction) {
	tx = &TestTransaction{
		Timestamp: timestamp(),
		From:      from.PublicKey,
		To:        toAddr,
		Asset:     asset,
		Value:     value,
		Comment:   comment,
		Tag:       tag,
	}
	tx.Sign = from.Sign(tx.Hash()) // set sign
	return
}

func NewTestEmission(toAddr crypto.Address, value int64, asset assets.Asset) (tx *TestTransaction) {
	return NewTestTransaction(MasterKey, toAddr, value, asset, "Emission", 0)
}

func (tx *TestTransaction) Type() blockchain.TxType {
	return TestTransactionType
}

func (tx *TestTransaction) String() string {
	return fmt.Sprintf("%s:%s:%+-d (%s)", tx.To, tx.Asset, tx.Value, tx.Comment)
}

func (tx *TestTransaction) Hash() []byte {
	return bin.Hash256(
		tx.Timestamp,
		tx.From,
		tx.To,
		tx.Asset,
		tx.Value,
		tx.Comment,
		tx.Tag,
	)
}

func (tx *TestTransaction) Encode() []byte {
	return bin.Encode(
		tx.Timestamp,
		tx.From,
		tx.To,
		tx.Asset,
		tx.Value,
		tx.Comment,
		tx.Tag,
		tx.Sign,
	)
}

func (tx *TestTransaction) Decode(data []byte) error {
	return bin.Decode(data,
		&tx.Timestamp,
		&tx.From,
		&tx.To,
		&tx.Asset,
		&tx.Value,
		&tx.Comment,
		&tx.Tag,
		&tx.Sign,
	)
}

func (tx *TestTransaction) Execute(c *state.State) {

	// verify tx
	if !tx.From.Verify(tx.Hash(), tx.Sign) {
		c.Fail(blockchain.ErrInvalidSign)
	}
	if tx.Value <= 0 {
		c.Fail(blockchain.ErrInvalidNum)
	}

	// exec transaction
	if !tx.From.Equal(config.MasterPublicKey) {
		c.Decrement(state.NewKey(tx.Asset, tx.From.Address()), state.Int(tx.Value), tx.Tag)
	}
	c.Increment(state.NewKey(tx.Asset, tx.To), state.Int(tx.Value), tx.Tag)

	// get state value by random address
	randomAddress := crypto.NewPrivateKey().PublicKey.Address()
	c.Get(state.NewKey(tx.Asset, randomAddress))
}
