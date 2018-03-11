package object

import (
	"encoding/json"
	"testing"

	"github.com/likecoin-pro/likecoin/blockchain/tests"
	"github.com/likecoin-pro/likecoin/blockchain/transaction"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/stretchr/testify/assert"
)

func TestUser_Verify(t *testing.T) {
	tx := NewUser(tests.AliceKey, "alice", config.MasterPublicKey.ID(), nil)

	err := tx.Verify()

	assert.NoError(t, err)
}

func TestUser_Verify_fail(t *testing.T) {
	tx := NewUser(tests.AliceKey, "alice", config.MasterPublicKey.ID(), nil)
	tx.Sign[0] = tx.Sign[0] - 1 // fail sign

	err := tx.Verify()

	assert.Error(t, err)
}

func TestUser_DecodeAsUnknownTx(t *testing.T) {
	tests.InitRand()
	tx := NewUser(tests.AliceKey, "alice", config.MasterPublicKey.ID(), nil)
	data := tx.Encode()

	var txUnknown = new(transaction.UnknownTransaction)
	err := txUnknown.Decode(data)

	assert.NoError(t, err)
	assert.Equal(t, tx.GetHeader(), txUnknown.GetHeader())
	assert.Equal(t, tx.Encode(), txUnknown.Encode())
	assert.Equal(t, transaction.Hash(tx), transaction.Hash(txUnknown))
}

func TestUser_Decode(t *testing.T) {
	tests.InitRand()
	tx := NewUser(tests.AliceKey, "alice", config.MasterPublicKey.ID(), nil)
	data := tx.Encode()

	var v User
	err := v.Decode(data)

	assert.NoError(t, err)
	assert.JSONEq(t, `
	{
	  "type":    2,
      "chain":   1, 
	  "version": 0,
	  "network": 1,
	  "chain":   1,
	  "nick":    "alice",
	  "pubkey":  "3mMBU5JPL3GT3e2iXr7VN1taMy4R1xL6Cw9gspfrRxjh9SJY5cfPaTorY6y9qcYWvQQmJeUC2qv2xmwASD2mehrH",
	  "referrer":"7ad1d904654415d0",
	  "data": "",
	  "signature": "8e71d29b331b8f4bc48da180a3777f4efa88b1435f5a8f7381af9dd4d73f9f96fbc93363088fce777667a07d9287f64eff16260ad69602ebb4ca57ac75ae1fe1"
	}`, enc.JSON(v))
}

func TestUser_JSONMarshal(t *testing.T) {
	tests.InitRand()
	tx := NewUser(tests.AliceKey, "alice", config.MasterPublicKey.ID(), nil)
	data, err := json.Marshal(tx)

	assert.NoError(t, err)
	assert.JSONEq(t, `
	{
	  "type": 2,
      "chain": 1, 
	  "version": 0,
	  "network": 1,
	  "chain": 1,
	  "nick": "alice",
	  "pubkey": "3mMBU5JPL3GT3e2iXr7VN1taMy4R1xL6Cw9gspfrRxjh9SJY5cfPaTorY6y9qcYWvQQmJeUC2qv2xmwASD2mehrH",
	  "referrer": "7ad1d904654415d0",
	  "data": null,
	  "signature": "8e71d29b331b8f4bc48da180a3777f4efa88b1435f5a8f7381af9dd4d73f9f96fbc93363088fce777667a07d9287f64eff16260ad69602ebb4ca57ac75ae1fe1"
	}`, string(data))
}

func TestUser_JSONUnmarshal(t *testing.T) {
	data := []byte(`
	{
	  "type": 2,
      "chain": 1,
	  "version": 0,
	  "network": 1,
	  "chain": 1,
	  "nick": "alice",
	  "pubkey": "3mMBU5JPL3GT3e2iXr7VN1taMy4R1xL6Cw9gspfrRxjh9SJY5cfPaTorY6y9qcYWvQQmJeUC2qv2xmwASD2mehrH",
	  "referrer": "7ad1d904654415d0",
	  "data": null,
	  "signature": "8e71d29b331b8f4bc48da180a3777f4efa88b1435f5a8f7381af9dd4d73f9f96fbc93363088fce777667a07d9287f64eff16260ad69602ebb4ca57ac75ae1fe1"
	}`)

	var tx = new(User)
	err := json.Unmarshal(data, tx)

	assert.NoError(t, err)
	assert.NoError(t, tx.Verify())
	assert.JSONEq(t, string(data), enc.JSON(tx))
}
