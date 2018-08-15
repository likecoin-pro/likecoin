package object

import (
	"encoding/json"
	"testing"

	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/stretchr/testify/assert"
)

func TestUser_Verify(t *testing.T) {
	tx := NewUser(aliceKey, "alice", bobID, nil)

	err := tx.Verify()

	assert.NoError(t, err)
}

func TestUser_Verify_fail(t *testing.T) {
	tx := NewUser(aliceKey, "alice", bobID, nil)
	tx.Sig[3]++ // corrupt signature

	err := tx.Verify()

	assert.Error(t, err)
}

func TestUser_JSONMarshal(t *testing.T) {
	tx := NewUser(aliceKey, "alice", bobID, nil)

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
