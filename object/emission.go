package object

import (
	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/commons/hex"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
)

type Emission struct {
	Version int            `json:"version"`   //
	Asset   assets.Asset   `json:"asset"`     // coin
	Comment string         `json:"comment"`   //
	Outs    []*EmissionOut `json:"outs"`      //
	Sign    hex.Bytes      `json:"signature"` // master signature
}

type EmissionOut struct {
	Addr       crypto.Address `json:"address"`   //
	Amount     int64          `json:"amount"`    // todo: ?? func(preMediaValue(MediaAddr), MediaValue)
	MediaAddr  string         `json:"media_uid"` // unique media ID
	MediaValue int64          `json:"media_val"` // current count likes on media
}

var _ = blockchain.RegisterTransactionType(&Emission{})

func (tx *Emission) Type() blockchain.TxType {
	return TxTypeEmission
}

func (tx *Emission) SetSign(prv *crypto.PrivateKey) {
	tx.Sign = prv.Sign(tx.hash())
}

func (tx *Emission) hash() []byte {
	return bin.Hash256(
		tx.Version,
		tx.Asset,
		tx.Comment,
		tx.Outs,
	)
}

func (tx *Emission) Encode() []byte {
	return bin.Encode(
		tx.Version,
		tx.Asset,
		tx.Comment,
		tx.Outs,
		tx.Sign,
	)
}

func (tx *Emission) Decode(data []byte) error {
	return bin.Decode(data,
		&tx.Version,
		&tx.Asset,
		&tx.Comment,
		&tx.Outs,
		&tx.Sign,
	)
}

func (tx *EmissionOut) Encode() []byte {
	return bin.Encode(
		tx.Addr,
		tx.Amount,
		tx.MediaAddr,
		tx.MediaValue,
	)
}

func (tx *EmissionOut) Decode(data []byte) error {
	return bin.Decode(data,
		&tx.Addr,
		&tx.Amount,
		&tx.MediaAddr,
		&tx.MediaValue,
	)
}

func (tx *Emission) Execute(st *state.State) {
	// verify transaction
	if !config.EmissionPublicKey.Verify(tx.hash(), tx.Sign) {
		st.Fail(ErrTxIncorrectSign)
	}

	coin := tx.Asset

	// change state
	for _, v := range tx.Outs {
		// set counter of media source
		st.Set(coin.CoinCounter(v.MediaAddr), crypto.NilAddress, state.Int(v.MediaValue), 0)

		// add coins to attached address
		st.Increment(coin, v.Addr, state.Int(v.Amount), 0)
	}
}
