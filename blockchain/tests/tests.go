package tests

import (
	cryptorand "crypto/rand"
	mathrand "math/rand"
	"time"

	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
)

var (
	Coin = assets.LikeCoin

	MasterKey = crypto.NewPrivateKeyBySecret("Test master key")

	AliceKey = crypto.NewPrivateKeyBySecret("Alice secret")
	BobKey   = crypto.NewPrivateKeyBySecret("Bob secret")
	CatKey   = crypto.NewPrivateKeyBySecret("Cat secret")

	AliceAddr = AliceKey.PublicKey.Address() // Like5A2PEu6eCHQzy1tMsa6b3kc1xXS7ywj2NQZr8xL
	BobAddr   = BobKey.PublicKey.Address()   // Like4fGoCMKi9LNqBbAdG3ppFuWRmGDM5bqSsQq9b37
	CatAddr   = CatKey.PublicKey.Address()   // Like3ssGc6gvkhbpveJRSPLxK8ZnKP7HEbhDXfwJfvk
)

func init() {
	config.MasterPublicKey = MasterKey.PublicKey
	config.EmissionPublicKey = MasterKey.PublicKey

	InitRand()
}

func InitRand() {
	cryptorand.Reader = mathrand.New(mathrand.NewSource(0))
}

func timestamp() int64 { // set test timestamp func
	return time.Now().UnixNano() / 1e3
}
