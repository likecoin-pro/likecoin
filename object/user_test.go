package object

import (
	"encoding/json"
	"testing"

	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/stretchr/testify/assert"
)

func TestUser_Decode(t *testing.T) {
	data := NewUser(testCfg, aliceKey, "alice", bobID, nil).Encode()

	var tx blockchain.Transaction
	err := tx.Decode(data)

	user := tx.TxObject().(*User)

	assert.NoError(t, err)
	assert.Equal(t, "alice", user.Nick)
	assert.Equal(t, uint64(0x7c14e6734f55d6d5), uint64(user.ReferrerID))

	//assert.JSONEq(t, `{
	//  "nick":     "alice",
	//  "referrer": "7c14e6734f55d6d5",
	//  "data":     null
	//}`, string(data))
}

func TestUser_Verify(t *testing.T) {
	tx := NewUser(testCfg, aliceKey, "alice", bobID, nil)

	err := tx.Verify(testCfg)

	assert.NoError(t, err)
}

func TestUser_Verify_fail(t *testing.T) {
	tx := NewUser(testCfg, aliceKey, "alice", bobID, nil)
	tx.Sig[3]++ // corrupt signature

	err := tx.Verify(testCfg)

	assert.Error(t, err)
}

func TestUser_JSONMarshal(t *testing.T) {
	tx := NewUser(testCfg, aliceKey, "alice", bobID, nil)

	println(bobID)
	//[]byte(enc.JSON(map[string]int{"abc": 123}))

	data, err := json.Marshal(tx.TxObject())

	assert.NoError(t, err)
	assert.JSONEq(t, `{
	  "nick":     "alice",
	  "referrer": "7c14e6734f55d6d5",
	  "data":     null
	}`, string(data))
}

func TestUser_JSONUnmarshal(t *testing.T) {
	data := []byte(`{
	  "nick":      "alice",
	  "referrer":  "7c14e6734f55d6d5",
	  "data":      null
	}`)

	var obj = new(User)
	err := json.Unmarshal(data, obj)

	assert.NoError(t, err)
	assert.JSONEq(t, string(data), enc.JSON(obj))
}
