package object

import (
	"encoding/json"
	"testing"

	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/commons/bignum"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/stretchr/testify/assert"
)

func TestEmission_Verify(t *testing.T) {

	tx := NewEmission(testCfg, emissionKey, coin, bignum.NewInt(1e9), "emission#1", []*EmissionOut{
		{aliceAddr, +200, "GDHxcxjUGAs", 888},
		{bobAddr, +100, "tDxrdnk4e5g", 777},
	})

	err := tx.Verify(testCfg)

	assert.NoError(t, err)
}

func TestEmission_Verify_fail(t *testing.T) {
	tx := NewEmission(testCfg, emissionKey, coin, bignum.NewInt(1e9), "emission#1", []*EmissionOut{
		{aliceAddr, +200, "GDHxcxjUGAs", 888},
		{bobAddr, +100, "tDxrdnk4e5g", 777},
	})
	tx.Sig[3]++ // corrupt signature

	err := tx.Verify(testCfg)

	assert.Error(t, err)
}

func TestEmission_Decode(t *testing.T) {
	data := NewEmission(testCfg, emissionKey, coin, bignum.NewInt(1e9), "emission#1", []*EmissionOut{
		{aliceAddr, +200, "GDHxcxjUGAs", 888},
		{bobAddr, +100, "tDxrdnk4e5g", 777},
	}).Encode()

	var tx blockchain.Transaction
	err := tx.Decode(data)

	assert.NoError(t, err)
	assert.NoError(t, tx.Verify(testCfg))
	assert.JSONEq(t, `
	{
	  "asset":   "0x0001",
	  "rate":    1000000000,
	  "comment": "emission#1",
	  "outs": [
		{
		  "address": "Like62D4Rq3s8D4Y5Q92YBoRiVpcMFEcXTyGLbqrtAv",
		  "delta":   200,
		  "srcID":   "GDHxcxjUGAs",
		  "srcVal":  888
		},
		{
		  "address": "Like4ujgQHL98BH21cPowptBCCTtHbAoygbjEU4iYmi",
		  "delta":   100,
		  "srcID":   "tDxrdnk4e5g",
		  "srcVal":  777
		}
	  ]
  
	}`, enc.JSON(tx.TxObject()))
}

func TestEmission_JSONMarshal(t *testing.T) {

	tx := NewEmission(testCfg, emissionKey, coin, bignum.NewInt(1e9), "emission#1", []*EmissionOut{
		{aliceAddr, +200, "GDHxcxjUGAs", 777},
		{bobAddr, +100, "tDxrdnk4e5g", 888},
	})

	data, err := json.Marshal(tx.TxObject())

	assert.NoError(t, err)
	assert.JSONEq(t, `
	{
      "asset": 	 "0x0001",
      "rate":    1000000000,
	  "comment": "emission#1",
	  "outs": [
		{
		  "address": "Like62D4Rq3s8D4Y5Q92YBoRiVpcMFEcXTyGLbqrtAv",
		  "delta":   200,
		  "srcID":   "GDHxcxjUGAs",
		  "srcVal":  777
		},
		{
		  "address": "Like4ujgQHL98BH21cPowptBCCTtHbAoygbjEU4iYmi",
		  "delta":   100,
		  "srcID":   "tDxrdnk4e5g",
		  "srcVal":  888
		}
	  ]
	}`, string(data))
}

func TestEmission_JSONUnmarshal(t *testing.T) {
	data := []byte(`
	{
      "asset": 	 "0x0001",
      "rate":    1000000000,
	  "comment": "emission#1",
	  "outs": [
		{
		  "address": "Like62D4Rq3s8D4Y5Q92YBoRiVpcMFEcXTyGLbqrtAv",
		  "delta":   200,
		  "srcID":   "GDHxcxjUGAs",
		  "srcVal":  777
		},
		{
		  "address": "Like4ujgQHL98BH21cPowptBCCTtHbAoygbjEU4iYmi",
		  "delta":   100,
		  "srcID":   "tDxrdnk4e5g",
		  "srcVal":  888
		}
	  ]
	}`)

	var obj = new(Emission)
	err := json.Unmarshal(data, obj)

	assert.NoError(t, err)
	assert.JSONEq(t, string(data), enc.JSON(obj))
}
