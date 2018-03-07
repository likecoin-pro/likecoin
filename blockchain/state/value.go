package state

import (
	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/crypto"
)

type Value struct {
	Asset   assets.Asset   `json:"asset"`
	Address crypto.Address `json:"address"`
	Value   Number         `json:"value"`
	Tag     int64          `json:"tag"`
}

func (v *Value) Equal(b *Value) bool {
	return v.Asset.Equal(b.Asset) &&
		v.Address == b.Address &&
		v.Value.Cmp(b.Value) == 0 &&
		v.Tag == b.Tag
}

func (v *Value) Hash() []byte {
	return crypto.Hash256(
		v.Asset,
		v.Address,
		v.Value,
		v.Tag,
	)
}

func (v *Value) Encode() []byte {
	return bin.Encode(
		v.Asset,
		v.Address,
		v.Value,
		v.Tag,
	)
}

func (v *Value) Decode(data []byte) error {
	return bin.Decode(data,
		&v.Asset,
		&v.Address,
		&v.Value,
		&v.Tag,
	)
}
