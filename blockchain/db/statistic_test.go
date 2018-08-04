package db

import (
	"testing"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/commons/bignum"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/stretchr/testify/assert"
)

func TestStatistic_Decode(t *testing.T) {
	data := bin.Encode(&Statistic{
		Blocks: 123456,
		Txs:    123456000,
		Users:  8888,
		Coins: []CoinStatistic{
			{assets.Default, 1e6, bignum.NewInt(1e9), bignum.NewInt(1e12)},
		},
	})

	var v *Statistic
	err := bin.Decode(data, &v)

	assert.NoError(t, err)
	assert.JSONEq(t, `{
	  "blocks": 123456,
	  "txs": 123456000,
	  "users": 8888,
	  "coins": [
		{
		  "asset": "0x0001",
		  "likes": 1000000,
		  "rate": 1000000000,
		  "supply": 1000000000000
		}
	  ]
	}`, enc.JSON(v))
}
