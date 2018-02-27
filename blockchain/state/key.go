package state

import (
	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/crypto"
)

type Key struct {
	Asset   crypto.Asset
	Address []byte
}

func NewKey(addr crypto.Address, asset crypto.Asset) Key {
	return Key{asset, addr.Encode()}
}

func NewCounterKey(addr string, asset crypto.Asset) Key {
	if !asset.IsCounter() {
		panic("Asset is not a counter")
	}
	return Key{asset, []byte(addr)}
}

func (k Key) strKey() string {
	return string(k.Asset) + ":" + string(k.Address)
}

func (k Key) String() string {
	return k.Asset.String() + ":" + k.StrAddress()
}

func (k Key) StrAddress() string {
	if k.Asset.IsCounter() { // external counter address
		return string(k.Address) // raw data to string
	}
	return crypto.StringAddress(k.Address)
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
