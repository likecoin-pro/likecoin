package db

import (
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/commons/bignum"
	"github.com/likecoin-pro/likecoin/commons/hex"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/likecoin-pro/likecoin/object"
)

type AddressInfo struct {
	Address     string       `json:"address"`      // original address
	AddressHex  string       `json:"addr_hex"`     //
	MemoAddress string       `json:"address_memo"` // address+memo
	Memo        string       `json:"memo"`         // memo in hex
	Balance     bignum.Int   `json:"balance"`      // balance on address (not memo address)
	Asset       assets.Asset `json:"asset"`        //
	LastTx      hex.Bytes    `json:"last_tx"`      // last tx of memo address
	User        *object.User `json:"user"`         // user associated with address
}

func (s *BlockchainStorage) AddressInfo(addr crypto.Address, memo uint64, asset assets.Asset) (inf AddressInfo, err error) {
	inf.MemoAddress = addr.MemoString(memo)
	inf.Address = addr.String()
	inf.AddressHex = addr.Hex()
	if memo != 0 {
		inf.Memo = "0x" + hex.EncodeUint(memo)
	}
	inf.Asset = asset
	bal, tx, err := s.GetBalance(addr, asset)
	if err != nil {
		return
	}
	inf.Balance = bal
	if memo != 0 {
		if tx, err = s.LastTx(addr, memo, asset); err != nil {
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
