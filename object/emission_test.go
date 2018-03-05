package object

import (
	"encoding/json"
	"testing"

	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain/tests"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/stretchr/testify/assert"
)

func TestEmission_Decode(t *testing.T) {
	tests.InitRand()
	tx := &Emission{
		Asset:   assets.LikeCoin,
		Comment: "emission#1",
		Outs: []*EmissionOut{
			{tests.AliceAddr, +200, "GDHxcxjUGAs", 200},
			{tests.BobAddr, +100, "tDxrdnk4e5g", 100},
		},
	}
	tx.SetSign(tests.MasterKey)
	data := tx.Encode()

	var v Emission
	err := v.Decode(data)

	assert.NoError(t, err)
	assert.JSONEq(t, `
	{
	  "version": 0,
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
	  "signature": "8e71d29b331b8f4bc48da180a3777f4efa88b1435f5a8f7381af9dd4d73f9f96113a960f45c589b3a3c5ad210f999ff991552c24103c41b0c577b5829ce3fa7f"
	}`, enc.JSON(v))
}

func TestEmission_JSONMarshal(t *testing.T) {
	tests.InitRand()
	tx := &Emission{
		Asset:   assets.LikeCoin,
		Comment: "emission#1",
		Outs: []*EmissionOut{
			{tests.AliceAddr, +200, "GDHxcxjUGAs", 200},
			{tests.BobAddr, +100, "tDxrdnk4e5g", 100},
		},
	}
	tx.SetSign(tests.MasterKey)

	data, err := json.Marshal(tx)

	assert.NoError(t, err)
	assert.JSONEq(t, `
	{
	  "version": 0,
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
	  "signature": "8e71d29b331b8f4bc48da180a3777f4efa88b1435f5a8f7381af9dd4d73f9f96113a960f45c589b3a3c5ad210f999ff991552c24103c41b0c577b5829ce3fa7f"
	}`, string(data))
}
