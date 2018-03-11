package object

import (
	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/blockchain/transaction"
	"github.com/likecoin-pro/likecoin/crypto"
)

type Transfer struct {
	transaction.Header
	From    *crypto.PublicKey `json:"from"`
	Outs    []*TransferOut    `json:"outs"`
	Comment string            `json:"comment"`
	Sign    bin.Bytes         `json:"signature"`
}

type TransferOut struct {
	Asset     assets.Asset   `json:"asset"`
	Amount    state.Number   `json:"amount"`
	Tag       int64          `json:"tag"`
	To        crypto.Address `json:"to"`
	ToTag     int64          `json:"to_tag"`
	ToChainID uint64         `json:"to_chain"`
}

var _ = transaction.Register(TxTypeTransfer, &Transfer{})

func NewSimpleTransfer(
	prv *crypto.PrivateKey,
	to crypto.Address,
	tag int64,
	asset assets.Asset,
	amount state.Number,
	comment string,
) (t *Transfer) {
	t = &Transfer{
		Header:  transaction.NewHeader(TxTypeTransfer, 0),
		Comment: comment,
	}
	t.AddOut(asset, amount, tag, to, tag, t.ChainID)
	t.SetSign(prv)
	return
}

func (tx *Transfer) AddOut(
	asset assets.Asset,
	amount state.Number,
	tag int64,
	to crypto.Address,
	toTag int64,
	toChainID uint64,
) {
	tx.Outs = append(tx.Outs, &TransferOut{
		Asset:     asset,
		Amount:    amount,
		Tag:       tag,
		To:        to,
		ToTag:     toTag,
		ToChainID: toChainID,
	})
}

func (tx *Transfer) hash() []byte {
	return crypto.Hash256(
		tx.Header,
		tx.From,
		tx.Outs,
		tx.Comment,
	)
}

func (tx *Transfer) SetSign(prv *crypto.PrivateKey) {
	tx.From = prv.PublicKey
	tx.Sign = prv.Sign(tx.hash())
}

func (tx *Transfer) Encode() []byte {
	return bin.Encode(
		tx.Header,
		tx.From,
		tx.Outs,
		tx.Comment,
		tx.Sign,
	)
}

func (tx *Transfer) Decode(data []byte) error {
	return bin.Decode(data,
		&tx.Header,
		&tx.From,
		&tx.Outs,
		&tx.Comment,
		&tx.Sign,
	)
}
func (t *TransferOut) Encode() []byte {
	return bin.Encode(
		t.Asset,
		t.Amount,
		t.Tag,
		t.To,
		t.ToTag,
		t.ToChainID,
	)
}

func (t *TransferOut) Decode(data []byte) error {
	return bin.Decode(data,
		&t.Asset,
		&t.Amount,
		&t.Tag,
		&t.To,
		&t.ToTag,
		&t.ToChainID,
	)
}

func (tx *Transfer) Totals() map[string]state.Number {
	vv := map[string]state.Number{}
	for _, out := range tx.Outs {
		sAsset := out.Asset.String()
		s, ok := vv[sAsset]
		if !ok {
			s = state.Int(0)
		}
		vv[sAsset] = s.Add(s, out.Amount)
	}
	return vv
}

func (tx *Transfer) Verify() error {
	if !tx.From.Verify(tx.hash(), tx.Sign) {
		return ErrTxIncorrectSign
	}
	for _, out := range tx.Outs {
		if out.Amount.Sign() <= 0 {
			return ErrTxIncorrectAmount
		}
	}
	return nil
}

func (tx *Transfer) Execute(st *state.State) {

	fromAddr := tx.From.Address()
	for _, out := range tx.Outs {

		// decrement amount from address; panic if not enough funds
		st.Decrement(out.Asset, fromAddr, out.Amount, out.Tag)

		// increment amount to new address
		if tx.ChainID == out.ToChainID {
			st.Increment(out.Asset, out.To, out.Amount, out.ToTag)
		} else {
			st.CrossChainSet(out.ToChainID, out.Asset, out.To, out.Amount, out.ToTag)
		}
	}
}
