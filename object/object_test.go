package object

import (
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
)

var (
	coin = assets.Likecoin

	emissionKey = crypto.NewPrivateKeyBySecret("Test master key")
	aliceKey    = crypto.NewPrivateKeyBySecret("alice::Alice secret")
	bobKey      = crypto.NewPrivateKeyBySecret("bob::Bob secret")

	aliceAddr = aliceKey.PublicKey.Address()
	bobAddr   = bobKey.PublicKey.Address()
	bobID     = bobKey.PublicKey.ID()

	testCfg = &blockchain.Config{
		NetworkID: blockchain.NetworkTest,
		ChainID:   1,
	}
)

func init() {
	config.EmissionPublicKey = emissionKey.PublicKey
}
