package object

import (
	"encoding/json"
	"testing"

	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/blockchain/tests"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/stretchr/testify/assert"
)

func TestTransfer_Verify(t *testing.T) {
	tx := NewSimpleTransfer(tests.AliceKey, tests.BobAddr, 123, tests.Coin, state.Int(100), "transfer to Bob")

	err := tx.Verify()

	assert.NoError(t, err)
}

func TestTransfer_Verify_fail(t *testing.T) {
	tx := NewSimpleTransfer(tests.AliceKey, tests.BobAddr, 123, tests.Coin, state.Int(100), "transfer to Bob")
	tx.Sign[0] = tx.Sign[0] - 1 // fail sign

	err := tx.Verify()

	assert.Error(t, err)
}

func TestTransfer_JSONMarshal(t *testing.T) {
	tests.InitRand()
	tx := NewSimpleTransfer(tests.AliceKey, tests.BobAddr, 123, tests.Coin, state.Int(100), "transfer to Bob")

	data, err := json.Marshal(tx)

	assert.NoError(t, err)
	assert.JSONEq(t, `{
	  "type": 1,
	  "version": 0,
      "network": 1,
	  "chain": 1,
	  "from": "3mMBU5JPL3GT3e2iXr7VN1taMy4R1xL6Cw9gspfrRxjh9SJY5cfPaTorY6y9qcYWvQQmJeUC2qv2xmwASD2mehrH",
	  "outs": [
		{
		  "asset": "0001",
		  "amount": 100,
		  "tag": 123,
		  "to": "Like4fGoCMKi9LNqBbAdG3ppFuWRmGDM5bqSsQq9b37",
		  "to_tag": 123,
		  "to_chain": 1
		}
	  ],
	  "comment": "transfer to Bob",
	  "signature": "8e71d29b331b8f4bc48da180a3777f4efa88b1435f5a8f7381af9dd4d73f9f96047759ec1ecc23911c25888c04fa222b5f1d5a1892cf1aa808a88fba969131aa"
	}`, string(data))
}

func TestTransfer_JSONUnmarshal(t *testing.T) {
	data := []byte(`{
	  "type": 1,
	  "version": 0,
      "network": 1,
	  "chain": 1,
	  "from": "3mMBU5JPL3GT3e2iXr7VN1taMy4R1xL6Cw9gspfrRxjh9SJY5cfPaTorY6y9qcYWvQQmJeUC2qv2xmwASD2mehrH",
	  "outs": [
		{
		  "asset": "0001",
		  "amount": 100,
		  "tag": 123,
		  "to": "Like4fGoCMKi9LNqBbAdG3ppFuWRmGDM5bqSsQq9b37",
		  "to_tag": 123,
		  "to_chain": 1
		}
	  ],
	  "comment": "transfer to Bob",
	  "signature": "8e71d29b331b8f4bc48da180a3777f4efa88b1435f5a8f7381af9dd4d73f9f96047759ec1ecc23911c25888c04fa222b5f1d5a1892cf1aa808a88fba969131aa"
	}`)

	var tx Transfer
	err := json.Unmarshal(data, &tx)

	assert.NoError(t, err)
	assert.NoError(t, tx.Verify())
	assert.JSONEq(t, string(data), enc.JSON(tx))
}
