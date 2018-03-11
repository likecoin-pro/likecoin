package object

import (
	"encoding/json"
	"testing"

	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain/tests"
	"github.com/likecoin-pro/likecoin/blockchain/transaction"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/stretchr/testify/assert"
)

func TestEmission_Verify(t *testing.T) {
	tx := NewEmission(tests.MasterKey, assets.LikeCoin, "emission#1",
		&EmissionOut{tests.AliceAddr, +200, "GDHxcxjUGAs", 200},
		&EmissionOut{tests.BobAddr, +100, "tDxrdnk4e5g", 100},
	)

	err := tx.Verify()

	assert.NoError(t, err)
}

func TestEmission_Verify_fail(t *testing.T) {
	tx := NewEmission(tests.MasterKey, assets.LikeCoin, "emission#1",
		&EmissionOut{tests.AliceAddr, +200, "GDHxcxjUGAs", 200},
		&EmissionOut{tests.BobAddr, +100, "tDxrdnk4e5g", 100},
	)
	tx.Sign[0] = tx.Sign[0] - 1 // fail sign

	err := tx.Verify()

	assert.Error(t, err)
}

func TestEmission_DecodeAsUnknownTx(t *testing.T) {
	tests.InitRand()
	txEmission := NewEmission(tests.MasterKey, assets.LikeCoin, "emission#1",
		&EmissionOut{tests.AliceAddr, +200, "GDHxcxjUGAs", 200},
		&EmissionOut{tests.BobAddr, +100, "tDxrdnk4e5g", 100},
	)
	data := txEmission.Encode()

	var txUnknown = new(transaction.UnknownTransaction)
	err := txUnknown.Decode(data)

	assert.NoError(t, err)
	assert.Equal(t, txEmission.GetHeader(), txUnknown.GetHeader())
	assert.Equal(t, txEmission.Encode(), txUnknown.Encode())
	assert.Equal(t, transaction.Hash(txEmission), transaction.Hash(txUnknown))
}

func TestEmission_Decode(t *testing.T) {
	tests.InitRand()
	tx := NewEmission(tests.MasterKey, assets.LikeCoin, "emission#1",
		&EmissionOut{tests.AliceAddr, +200, "GDHxcxjUGAs", 200},
		&EmissionOut{tests.BobAddr, +100, "tDxrdnk4e5g", 100},
	)
	data := tx.Encode()

	var v Emission
	err := v.Decode(data)

	assert.NoError(t, err)
	assert.NoError(t, tx.Verify())
	assert.JSONEq(t, `
	{
	  "type": 0,
	  "version": 0,
	  "network": 1,
	  "chain": 1,
	  "asset": "0001",
	  "comment": "emission#1",
	  "outs": [
		{
		  "address": "Like5A2PEu6eCHQzy1tMsa6b3kc1xXS7ywj2NQZr8xL",
		  "amount": 200,
		  "media_uid": "GDHxcxjUGAs",
		  "media_val": 200
		},
		{
		  "address": "Like4fGoCMKi9LNqBbAdG3ppFuWRmGDM5bqSsQq9b37",
		  "amount": 100,
		  "media_uid": "tDxrdnk4e5g",
		  "media_val": 100
		}
	  ],
	  "signature": "8e71d29b331b8f4bc48da180a3777f4efa88b1435f5a8f7381af9dd4d73f9f96e29f9e28ec7065b63bc406db4b2563c7277aa23772ced2530422fb98c9b87d9d"
	}`, enc.JSON(v))
}

func TestEmission_JSONMarshal(t *testing.T) {
	tests.InitRand()
	tx := NewEmission(tests.MasterKey, assets.LikeCoin, "emission#1",
		&EmissionOut{tests.AliceAddr, +200, "GDHxcxjUGAs", 200},
		&EmissionOut{tests.BobAddr, +100, "tDxrdnk4e5g", 100},
	)
	data, err := json.Marshal(tx)

	assert.NoError(t, err)
	assert.JSONEq(t, `
	{
	  "type": 0,
	  "version": 0,
      "network": 1,
	  "chain": 1,
	  "asset": "0001",
	  "comment": "emission#1",
	  "outs": [
		{
		  "address": "Like5A2PEu6eCHQzy1tMsa6b3kc1xXS7ywj2NQZr8xL",
		  "amount": 200,
		  "media_uid": "GDHxcxjUGAs",
		  "media_val": 200
		},
		{
		  "address": "Like4fGoCMKi9LNqBbAdG3ppFuWRmGDM5bqSsQq9b37",
		  "amount": 100,
		  "media_uid": "tDxrdnk4e5g",
		  "media_val": 100
		}
	  ],
	  "signature": "8e71d29b331b8f4bc48da180a3777f4efa88b1435f5a8f7381af9dd4d73f9f96e29f9e28ec7065b63bc406db4b2563c7277aa23772ced2530422fb98c9b87d9d"
	}`, string(data))
}

func TestEmission_JSONUnmarshal(t *testing.T) {
	data := []byte(`
	{
	  "type": 0,
	  "version": 0,
      "network": 1,
	  "chain": 1,
	  "asset": "0001",
	  "comment": "emission#1",
	  "outs": [
		{
		  "address": "Like5A2PEu6eCHQzy1tMsa6b3kc1xXS7ywj2NQZr8xL",
		  "amount": 200,
		  "media_uid": "GDHxcxjUGAs",
		  "media_val": 200
		},
		{
		  "address": "Like4fGoCMKi9LNqBbAdG3ppFuWRmGDM5bqSsQq9b37",
		  "amount": 100,
		  "media_uid": "tDxrdnk4e5g",
		  "media_val": 100
		}
	  ],
	  "signature": "8e71d29b331b8f4bc48da180a3777f4efa88b1435f5a8f7381af9dd4d73f9f96e29f9e28ec7065b63bc406db4b2563c7277aa23772ced2530422fb98c9b87d9d"
	}`)

	var tx = new(Emission)
	err := json.Unmarshal(data, tx)

	assert.NoError(t, err)
	assert.NoError(t, tx.Verify())
	assert.JSONEq(t, string(data), enc.JSON(tx))
}
