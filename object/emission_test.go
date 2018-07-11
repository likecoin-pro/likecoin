package object

import (
	"encoding/json"
	"testing"

	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/likecoin-pro/likecoin/tests"
	"github.com/stretchr/testify/assert"
)

func TestEmission_Verify(t *testing.T) {
	tx := NewEmission(tests.MasterKey, assets.Likecoin, "emission#1", []*EmissionOut{
		{tests.AliceAddr, state.Int(+200), "GDHxcxjUGAs", 200},
		{tests.BobAddr, state.Int(+100), "tDxrdnk4e5g", 100},
	})

	err := tx.Verify()

	assert.NoError(t, err)
}

func TestEmission_Verify_fail(t *testing.T) {
	tx := NewEmission(tests.MasterKey, assets.Likecoin, "emission#1", []*EmissionOut{
		{tests.AliceAddr, state.Int(+200), "GDHxcxjUGAs", 200},
		{tests.BobAddr, state.Int(+100), "tDxrdnk4e5g", 100},
	})
	tx.Sig[3]++ // corrupt signature

	err := tx.Verify()

	assert.Error(t, err)
}

func TestEmission_Decode(t *testing.T) {
	data := NewEmission(tests.MasterKey, assets.Likecoin, "emission#1", []*EmissionOut{
		{tests.AliceAddr, state.Int(+2e9), "GDHxcxjUGAs", 200},
		{tests.BobAddr, state.Int(+1e9), "tDxrdnk4e5g", 100},
	}).Encode()

	var tx blockchain.Transaction
	err := tx.Decode(data)

	assert.NoError(t, err)
	assert.NoError(t, tx.Verify())
	assert.JSONEq(t, `
	{
	  "asset":   "0x0001",
	  "comment": "emission#1",
	  "outs": [
		{
		  "address": "Like62D4Rq3s8D4Y5Q92YBoRiVpcMFEcXTyGLbqrtAv",
		  "amount":  2000000000,
		  "srcID":   "GDHxcxjUGAs",
		  "srcVal":  200
		},
		{
		  "address": "Like4ujgQHL98BH21cPowptBCCTtHbAoygbjEU4iYmi",
		  "amount":  1000000000,
		  "srcID":   "tDxrdnk4e5g",
		  "srcVal":  100
		}
	  ]
  
	}`, enc.JSON(tx.TxObject()))
}

func TestEmission_JSONMarshal(t *testing.T) {

	tx := NewEmission(tests.MasterKey, assets.Likecoin, "emission#1", []*EmissionOut{
		{tests.AliceAddr, state.Int(+2e9), "GDHxcxjUGAs", 200},
		{tests.BobAddr, state.Int(+1e9), "tDxrdnk4e5g", 100},
	})

	data, err := json.Marshal(tx.TxObject())

	assert.NoError(t, err)
	assert.JSONEq(t, `
	{
      "asset": 	 "0x0001",
	  "comment": "emission#1",
	  "outs": [
		{
		  "address": "Like62D4Rq3s8D4Y5Q92YBoRiVpcMFEcXTyGLbqrtAv",
		  "amount":  2000000000,
		  "srcID":   "GDHxcxjUGAs",
		  "srcVal":  200
		},
		{
		  "address": "Like4ujgQHL98BH21cPowptBCCTtHbAoygbjEU4iYmi",
		  "amount":  1000000000,
		  "srcID":   "tDxrdnk4e5g",
		  "srcVal":  100
		}
	  ]
	}`, string(data))
}

func TestEmission_JSONUnmarshal(t *testing.T) {
	data := []byte(`
	{
      "asset": 	 "0x0001",
	  "comment": "emission#1",
	  "outs": [
		{
		  "address": "Like62D4Rq3s8D4Y5Q92YBoRiVpcMFEcXTyGLbqrtAv",
		  "amount":  2000000000,
		  "srcID":   "GDHxcxjUGAs",
		  "srcVal":  200
		},
		{
		  "address": "Like4ujgQHL98BH21cPowptBCCTtHbAoygbjEU4iYmi",
		  "amount":  1000000000,
		  "srcID":   "tDxrdnk4e5g",
		  "srcVal":  100
		}
	  ]
	}`)

	var obj = new(Emission)
	err := json.Unmarshal(data, obj)

	assert.NoError(t, err)
	assert.JSONEq(t, string(data), enc.JSON(obj))
}
