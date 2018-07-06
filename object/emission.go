package object

import (
	"errors"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
)

type Emission struct {
	Asset   assets.Asset   `json:"asset"`   // coin
	Rate    string         `json:"rate"`    // new like-rate
	Comment string         `json:"comment"` //
	Outs    []*EmissionOut `json:"outs"`    //
}

type EmissionOut struct {
	Address     crypto.Address `json:"address"` // address linked with media-source
	Amount      state.Number   `json:"amount"`  // emission amount (in coin)
	SourceID    string         `json:"srcID"`   // unique media-source ID
	SourceValue int64          `json:"srcVal"`  // current count likes for media-source
}

var _ = blockchain.RegisterTxObject(TxTypeEmission, &Emission{})

var (
	ErrTxEmptyMediaAddr = errors.New("emission-tx: Empty media address")
)

func NewEmission(
	emissionKey *crypto.PrivateKey,
	asset assets.Asset,
	comment string,
	vv []*EmissionOut,
) *blockchain.Transaction {
	return blockchain.NewTx(emissionKey, &Emission{
		Asset:   asset,
		Comment: comment,
		Outs:    vv,
	})
}

func (obj *Emission) Encode() []byte {
	return bin.Encode(
		obj.Asset,
		//obj.Rate,
		obj.Comment,
		obj.Outs,
	)
}

func (obj *Emission) Decode(data []byte) error {
	return bin.Decode(data,
		&obj.Asset,
		//&obj.Rate,
		&obj.Comment,
		&obj.Outs,
	)
}

func (tx *EmissionOut) Encode() []byte {
	return bin.Encode(
		tx.Address,
		tx.Amount,
		tx.SourceID,
		tx.SourceValue,
	)
}

func (out *EmissionOut) Decode(data []byte) error {
	return bin.Decode(data,
		&out.Address,
		&out.Amount,
		&out.SourceID,
		&out.SourceValue,
	)
}

func (obj *Emission) TotalAmount() (s state.Number) {
	s = state.Int(0)
	for _, v := range obj.Outs {
		s.Add(s, v.Amount)
	}
	return
}

func (obj *Emission) OutBySrc(srcID string) *EmissionOut {
	for _, o := range obj.Outs {
		if o.SourceID == srcID {
			return o
		}
	}
	return nil
}

func (obj *Emission) Verify(tx *blockchain.Transaction) error {

	if !tx.Sender.Equal(config.EmissionPublicKey) { // Sender of emission-tx must be EmissionPublicKey
		return ErrTxIncorrectIssuer
	}

	for _, v := range obj.Outs {
		if v.SourceID == "" {
			return ErrTxEmptyMediaAddr
		}
	}
	return nil
}

func (obj *Emission) Execute(tx *blockchain.Transaction, st *state.State) {

	coin := obj.Asset

	// change state
	for _, v := range obj.Outs {

		// set counter of media source
		//st.Set(coin.SourceCounter(v.SourceID), crypto.NilAddress, state.Int(v.SourceValue), 0)

		// add coins to attached address
		st.Increment(coin, v.Address, v.Amount, 0)
	}

	// refresh total supply
	st.Increment(coin, crypto.NilAddress, obj.TotalAmount(), 0)
}
