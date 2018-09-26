package db

import (
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/commons/bignum"
	"github.com/likecoin-pro/likecoin/commons/hex"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/likecoin-pro/likecoin/object"
)

type AddressInfo struct {
	Address       string       `json:"address"`     // original address
	AddressHex    string       `json:"addr_hex"`    //
	TaggedAddress string       `json:"address_tag"` // address+tag
	Tag           string       `json:"tag"`         // tag in hex
	Balance       bignum.Int   `json:"balance"`     // balance on address (not tagged address)
	Asset         assets.Asset `json:"asset"`       //
	LastTx        hex.Bytes    `json:"last_tx"`     // last tx of tagged_address
	User          *object.User `json:"user"`        // user associated with address
}

func (s *BlockchainStorage) AddressInfo(addr crypto.Address, tag uint64, asset assets.Asset) (inf AddressInfo, err error) {
	inf.TaggedAddress = addr.MemoString(tag)
	inf.Address = addr.String()
	inf.AddressHex = addr.Hex()
	if tag != 0 {
		inf.Tag = "0x" + hex.EncodeUint(tag)
	}
	inf.Asset = asset
	bal, tx, err := s.GetBalance(addr, asset)
	if err != nil {
		return
	}
	inf.Balance = bal
	if tag != 0 {
		if tx, err = s.LastTx(addr, tag, asset); err != nil {
			return
		}
	}
	if tx != nil {
		inf.LastTx = tx.Hash()
	}
	if _, inf.User, err = s.UserByID(addr.ID()); err != nil {
		return
	}

	return
}
