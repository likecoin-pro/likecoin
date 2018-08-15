package object

import (
	"encoding/json"
	"testing"

	"github.com/likecoin-pro/likecoin/commons/bignum"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/stretchr/testify/assert"
)

func TestTransfer_Verify(t *testing.T) {
	tx := NewSimpleTransfer(aliceKey, bobAddr, bignum.NewInt(100), coin, "transfer to Bob", 123, 1456)

	err := tx.Verify()

	assert.NoError(t, err)
}

func TestTransfer_Verify_fail(t *testing.T) {
	tx := NewSimpleTransfer(aliceKey, bobAddr, bignum.NewInt(100), coin, "transfer to Bob", 123, 456)

	tx.Sig[3]++ // corrupt sign

	err := tx.Verify()

	assert.Error(t, err)
}

func TestTransfer_JSONMarshal(t *testing.T) {
	tx := NewSimpleTransfer(aliceKey, bobAddr, bignum.NewInt(1.5e9), coin, "transfer to Bob", 123, 456)

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
		  "to_tag": 456,
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
		  "to_tag": 456,
		  "to_chain": 1
		}
	  ]
	}`)

	var obj = new(Transfer)
	err := json.Unmarshal(data, obj)

	assert.NoError(t, err)
	assert.JSONEq(t, string(data), enc.JSON(obj))
}
