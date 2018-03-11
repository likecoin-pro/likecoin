package object

import (
	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/blockchain/transaction"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
)

type Emission struct {
	transaction.Header
	Issuer  *crypto.PublicKey `json:"issuer"`    //
	Asset   assets.Asset      `json:"asset"`     // coin
	Comment string            `json:"comment"`   //
	Outs    []*EmissionOut    `json:"outs"`      //
	Sign    bin.Bytes         `json:"signature"` // master signature
}

type EmissionOut struct {
	Address    crypto.Address `json:"address"`   //
	Amount     int64          `json:"amount"`    // todo: ?? func(preMediaValue(MediaAddr), MediaValue)
	MediaAddr  string         `json:"media_uid"` // unique media ID
	MediaValue int64          `json:"media_val"` // current count likes on media
}

var _ = transaction.Register(TxTypeEmission, &Emission{})

func NewEmission(
	prv *crypto.PrivateKey,
	asset assets.Asset,
	comment string,
	vv ...*EmissionOut,
) *Emission {
	tx := &Emission{
		Header:  transaction.NewHeader(TxTypeEmission, 0),
		Issuer:  prv.PublicKey,
		Asset:   asset,
		Comment: comment,
		Outs:    vv,
	}
	tx.SetSign(prv)
	return tx
}

func (tx *Emission) SetSign(prv *crypto.PrivateKey) {
	tx.Sign = prv.Sign(tx.hash())
}

func (tx *Emission) hash() []byte {
	return crypto.Hash256(
		tx.Header,
		tx.Issuer,
		tx.Asset,
		tx.Comment,
		tx.Outs,
	)
}

func (tx *Emission) Encode() []byte {
	return bin.Encode(
		tx.Header,
		tx.Issuer,
		tx.Asset,
		tx.Comment,
		tx.Outs,
		tx.Sign,
	)
}

func (tx *Emission) Decode(data []byte) error {
	return bin.Decode(data,
		&tx.Header,
		&tx.Issuer,
		&tx.Asset,
		&tx.Comment,
		&tx.Outs,
		&tx.Sign,
	)
}

func (tx *EmissionOut) Encode() []byte {
	return bin.Encode(
		tx.Address,
		tx.Amount,
		tx.MediaAddr,
		tx.MediaValue,
	)
}

func (tx *EmissionOut) Decode(data []byte) error {
	return bin.Decode(data,
		&tx.Address,
		&tx.Amount,
		&tx.MediaAddr,
		&tx.MediaValue,
	)
}

func (tx *Emission) Verify() error {
	if !tx.Issuer.Equal(config.EmissionPublicKey) {
		return ErrTxIncorrectIssuer
	}
	if !tx.Issuer.Verify(tx.hash(), tx.Sign) {
		return ErrTxIncorrectSign
	}
	return nil
}

func (tx *Emission) Execute(st *state.State) {

	coin := tx.Asset

	// change state
	for _, v := range tx.Outs {
		// set counter of media source
		st.Set(coin.CoinCounter(v.MediaAddr), crypto.NilAddress, state.Int(v.MediaValue), 0)

		// add coins to attached address
		st.Increment(coin, v.Address, state.Int(v.Amount), 0)
	}
}
