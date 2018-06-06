package db

import (
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/commons/hex"
	"github.com/likecoin-pro/likecoin/crypto"
)

type AddressInfo struct {
	Address       string       `json:"address"`  // original address
	AddressHex    string       `json:"addr_hex"` //
	TaggedAddress string       `json:"addr+tag"` // address+tag
	Tag           uint64       `json:"tag"`      //
	Balance       state.Number `json:"balance"`  // balance on address (not tagged address)
	Asset         assets.Asset `json:"asset"`    //
	LastTx        hex.Bytes    `json:"last_tx"`  // last tx of tagged_address
}

func (s *BlockchainStorage) AddressInfo(addr crypto.Address, tag uint64, asset assets.Asset) (inf AddressInfo, err error) {
	inf.TaggedAddress = addr.TaggedString(tag)
	inf.Address = addr.String()
	inf.AddressHex = addr.Hex()
	inf.Tag = tag
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
		inf.LastTx = tx.Tx.Hash()
	}
	return
}
