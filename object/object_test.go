package object

import (
	"github.com/likecoin-pro/likecoin/assets"
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
)

func init() {
	config.NetworkID = 1 // test network
	config.ChainID = 1
	config.EmissionPublicKey = emissionKey.PublicKey
}
