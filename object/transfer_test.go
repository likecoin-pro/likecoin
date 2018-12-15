package object

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/likecoin-pro/likecoin/commons/bignum"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/stretchr/testify/assert"
)

func TestTransfer_Verify(t *testing.T) {
	tx := NewSimpleTransfer(testCfg, aliceKey, bobAddr, bignum.NewInt(100), coin, "transfer to Bob", 123, 1456)

	err := tx.Verify(testCfg)

	assert.NoError(t, err)
}

func TestTransfer_Encode(t *testing.T) {
	tx := NewSimpleTransfer(testCfg, aliceKey, bobAddr, bignum.NewInt(100), coin, "Test", 123, 456)

	data := tx.Encode()

	assert.Equal(t, `010001018705543df729c0062a000122020001647b187c14e6734f55d6d594d5af08c142120d38d44a49421311748201c8010454657374000021034093cdf68e4fbeea9307530b20138fd56675f386a4eb0daa1f8067435e4eef9a4142782952b94b22a04518d9303ddf8292b2ecd6e349a275090d34c6810fc805b67acd0359ea2e6f6f2aa21965b3dfb39e9a4f41bc07697dfc4a74156b2277b7a20100`, hex.EncodeToString(data))
}

func TestTransfer_Verify_fail(t *testing.T) {
	tx := NewSimpleTransfer(testCfg, aliceKey, bobAddr, bignum.NewInt(100), coin, "transfer to Bob", 123, 456)

	tx.Sig[3]++ // corrupt sign

	err := tx.Verify(testCfg)

	assert.Error(t, err)
}

func TestTransfer_JSONMarshal(t *testing.T) {
	tx := NewSimpleTransfer(testCfg, aliceKey, bobAddr, bignum.NewInt(1.5e9), coin, "transfer to Bob", 123, 456)

	data, err := json.Marshal(tx.TxObject())

	assert.NoError(t, err)
	assert.NoError(t, tx.Verify(testCfg))
	assert.JSONEq(t, `{
	  "comment": "transfer to Bob",
	  "outs": [
		{
		  "asset": "0x0001",
		  "amount": 1500000000,
		  "tag": 123,
		  "to": "Like4ujgQHL98BH21cPowptBCCTtHbAoygbjEU4iYmi",
		  "to_memo": 456,
		  "to_memo_address": "Like2KAAshsYmxGiMapWx3v6k4TZzLfZma6asRZmyfmhF5",
		  "to_chain": 1,
		  "to_nick": ""
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
		  "to_memo": 456,
		  "to_memo_address": "Like2KAAshsYmxGiMapWx3v6k4TZzLfZma6asRZmyfmhF5",
		  "to_chain": 1,
		  "to_nick": ""
		}
	  ]
	}`)

	var obj = new(Transfer)
	err := json.Unmarshal(data, obj)

	assert.NoError(t, err)
	assert.JSONEq(t, string(data), enc.JSON(obj))
}
