package state

import (
	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/crypto"
)

type Key struct {
	Asset   assets.Asset
	Address crypto.Address
}

func NewKey(asset assets.Asset, addr crypto.Address) Key {
	return Key{asset, addr}
}

func NewCounterKey(coin assets.Asset, counterID string) Key {
	if !coin.IsCoin() {
		panic("Asset is not a coin")
	}
	return NewKey(coin.CoinCounter(counterID), crypto.NilAddress)
}

func (k Key) strKey() string {
	return string(k.Asset) + string(k.Address[:])
}

func (k Key) String() string {
	return k.Asset.String() + ":" + k.Address.String()
}

func (k Key) Encode() []byte {
	return bin.Encode(
		k.Asset,
		k.Address,
	)
}

func (k *Key) Decode(data []byte) error {
	return bin.Decode(data,
		&k.Asset,
		&k.Address,
	)
}
