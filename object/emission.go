package object

import (
	"errors"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/commons/bignum"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
)

type Emission struct {
	Asset   assets.Asset   `json:"asset"`   // coin
	Rate    bignum.Int     `json:"rate"`    // current like rate (in coins)
	Comment string         `json:"comment"` //
	Outs    []*EmissionOut `json:"outs"`    //
}

type EmissionOut struct {
	Address     crypto.Address `json:"address"` // address associated with the media-source
	Delta       int64          `json:"delta"`   //
	SourceID    string         `json:"srcID"`   // unique media-source ID
	SourceValue int64          `json:"srcVal"`  // current count likes for media-source
}

var _ = blockchain.RegisterTxObject(TxTypeEmission, &Emission{})

var (
	ErrEmissionTxEmptyAddr      = errors.New("emission-tx: empty address")
	ErrEmissionTxEmptySourceID  = errors.New("emission-tx: empty source ID")
	ErrEmissionTxIncorrectRate  = errors.New("emission-tx: incorrect rate")
	ErrEmissionTxIncorrectDelta = errors.New("emission-tx: incorrect delta")
)

func NewEmission(
	emissionKey *crypto.PrivateKey,
	asset assets.Asset,
	rate bignum.Int,
	comment string,
	vv []*EmissionOut,
) *blockchain.Transaction {
	return blockchain.NewTx(emissionKey, 0, &Emission{
		Asset:   asset,
		Rate:    rate,
		Comment: comment,
		Outs:    vv,
	})
}

func (obj *Emission) Encode() []byte {
	return bin.Encode(
		0, // ver
		obj.Asset,
		obj.Rate,
		obj.Comment,
		obj.Outs,
	)
}

func (obj *Emission) Decode(data []byte) error {
	return bin.Decode(data,
		new(int),
		&obj.Asset,
		&obj.Rate,
		&obj.Comment,
		&obj.Outs,
	)
}

func (out *EmissionOut) Encode() []byte {
	return bin.Encode(
		out.Address,
		out.Delta,
		out.SourceID,
		out.SourceValue,
	)
}

func (out *EmissionOut) Decode(data []byte) error {
	return bin.Decode(data,
		&out.Address,
		&out.Delta,
		&out.SourceID,
		&out.SourceValue,
	)
}

func (obj *Emission) Amount(delta int64) bignum.Int {
	return bignum.NewInt(delta).Mul(obj.Rate)
}

func (obj *Emission) TotalDelta() (likes int64) {
	for _, out := range obj.Outs {
		likes += out.Delta
	}
	return
}

func (obj *Emission) TotalAmount() bignum.Int {
	return obj.Amount(obj.TotalDelta())
}

func (obj *Emission) OutBySrc(srcID string) *EmissionOut {
	for _, out := range obj.Outs {
		if out.SourceID == srcID {
			return out
		}
	}
	return nil
}

func (obj *Emission) Verify(tx *blockchain.Transaction) error {

	if !tx.Sender.Equal(config.EmissionPublicKey) { // Sender of emission-tx must be EmissionPublicKey
		return ErrTxIncorrectSender
	}
	if obj.Rate.Sign() < 0 {
		return ErrEmissionTxIncorrectRate
	}

	for _, out := range obj.Outs {
		if out.Delta < 0 || out.SourceValue < 0 || out.Delta > out.SourceValue {
			return ErrEmissionTxIncorrectDelta
		}
		if out.SourceID == "" {
			return ErrEmissionTxEmptySourceID
		}
		if out.Address.Empty() && out.Delta > 0 {
			return ErrEmissionTxEmptyAddr
		}
	}
	return nil
}

func (obj *Emission) Execute(tx *blockchain.Transaction, st *state.State) {

	coin := obj.Asset

	// change state
	for _, out := range obj.Outs {
		// add coins to attached address
		if !out.Address.Empty() && out.Delta > 0 {
			st.Increment(coin, out.Address, obj.Amount(out.Delta), 0)
		}
	}
}
