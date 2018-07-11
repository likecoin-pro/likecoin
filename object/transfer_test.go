package object

import (
	"encoding/json"
	"testing"

	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/likecoin-pro/likecoin/tests"
	"github.com/stretchr/testify/assert"
)

func TestTransfer_Verify(t *testing.T) {
	tx := NewSimpleTransfer(tests.AliceKey, tests.BobAddr, state.Int(100), tests.Coin, "transfer to Bob", 123)

	err := tx.Verify()

	assert.NoError(t, err)
}

func TestTransfer_Verify_fail(t *testing.T) {
	tx := NewSimpleTransfer(tests.AliceKey, tests.BobAddr, state.Int(100), tests.Coin, "transfer to Bob", 123)

	tx.Sig[3]++ // corrupt sign

	err := tx.Verify()

	assert.Error(t, err)
}

func TestTransfer_JSONMarshal(t *testing.T) {
	tx := NewSimpleTransfer(tests.AliceKey, tests.BobAddr, state.Int(1.5e9), tests.Coin, "transfer to Bob", 123)

	data, err := json.Marshal(tx.TxObject())

	assert.NoError(t, err)
	assert.NoError(t, tx.Verify())
	assert.JSONEq(t, `{
	  "comment": "transfer to Bob",
	  "outs": [
		{
		  "asset": "0x0001",
		  "amount": 1500000000,
		  "tag": 123,
		  "to": "Like4ujgQHL98BH21cPowptBCCTtHbAoygbjEU4iYmi",
		  "to_tag": 123,
		  "to_chain": 1
		}
	  ]
	}`, string(data))
}

func TestTransfer_JSONUnmarshal(t *testing.T) {
	data := []byte(`{
	  "comment": "transfer to Bob",
	  "outs": [
		{
		  "asset": "0x0001",
		  "amount": 1500000000,
		  "tag": 123,
		  "to": "Like4ujgQHL98BH21cPowptBCCTtHbAoygbjEU4iYmi",
		  "to_tag": 123,
		  "to_chain": 1
		}
	  ]
	}`)

	var obj = new(Transfer)
	err := json.Unmarshal(data, obj)

	assert.NoError(t, err)
	assert.JSONEq(t, string(data), enc.JSON(obj))
}
