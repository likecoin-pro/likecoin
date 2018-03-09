package state

import (
	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/crypto"
)

type Value struct {
	ChainID uint64         `json:"chain"`
	Asset   assets.Asset   `json:"asset"`
	Address crypto.Address `json:"address"`
	Tag     int64          `json:"tag"`
	Value   Number         `json:"total"`
}

func (v *Value) Equal(b *Value) bool {
	return v.Asset.Equal(b.Asset) &&
		v.Address == b.Address &&
		v.Value.Cmp(b.Value) == 0 &&
		v.Tag == b.Tag
}

func (v *Value) Hash() []byte {
	return crypto.Hash256(
		v.ChainID,
		v.Asset,
		v.Address,
		v.Tag,
		v.Value,
	)
}

func (v *Value) Encode() []byte {
	return bin.Encode(
		v.ChainID,
		v.Asset,
		v.Address,
		v.Tag,
		v.Value,
	)
}

func (v *Value) Decode(data []byte) error {
	return bin.Decode(data,
		&v.ChainID,
		&v.Asset,
		&v.Address,
		&v.Tag,
		&v.Value,
	)
}
