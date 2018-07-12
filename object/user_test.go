package object

import (
	"encoding/json"
	"testing"

	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/tests"
	"github.com/stretchr/testify/assert"
)

func TestUser_Verify(t *testing.T) {
	tx := NewUser(tests.AliceKey, "alice", config.MasterPublicKey.ID(), nil)

	err := tx.Verify()

	assert.NoError(t, err)
}

func TestUser_Verify_fail(t *testing.T) {
	tx := NewUser(tests.AliceKey, "alice", config.MasterPublicKey.ID(), nil)
	tx.Sig[3]++ // corrupt signature

	err := tx.Verify()

	assert.Error(t, err)
}

func TestUser_JSONMarshal(t *testing.T) {
	tx := NewUser(tests.AliceKey, "alice", config.MasterPublicKey.ID(), nil)

	data, err := json.Marshal(tx.TxObject())

	assert.NoError(t, err)
	assert.JSONEq(t, `{
	  "nick":     "alice",
	  "referrer": "e4a962c20df12faf",
	  "data":     null
	}`, string(data))
}

func TestUser_JSONUnmarshal(t *testing.T) {
	data := []byte(`{
	  "nick":      "alice",
	  "referrer":  "e4a962c20df12faf",
	  "data":      null
	}`)

	var obj = new(User)
	err := json.Unmarshal(data, obj)

	assert.NoError(t, err)
	assert.JSONEq(t, string(data), enc.JSON(obj))
}
