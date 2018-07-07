package object

import (
	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
)

type Transfer struct {
	Outs    []*TransferOut `json:"outs"`
	Comment string         `json:"comment"`
}

type TransferOut struct {
	Asset     assets.Asset   `json:"asset"`
	Amount    state.Number   `json:"amount"`
	Tag       uint64         `json:"tag"`
	To        crypto.Address `json:"to"`
	ToTag     uint64         `json:"to_tag"`
	ToChainID uint64         `json:"to_chain"`
}

var _ = blockchain.RegisterTxObject(TxTypeTransfer, &Transfer{})

func NewSimpleTransfer(
	from *crypto.PrivateKey,
	toAddr crypto.Address,
	amount state.Number,
	asset assets.Asset,
	comment string,
	tag uint64,
) *blockchain.Transaction {
	tr := &Transfer{
		Comment: comment,
	}
	tr.AddOut(asset, amount, tag, toAddr, tag, config.ChainID)
	return blockchain.NewTx(from, tr)
}

func (obj *Transfer) AddOut(
	asset assets.Asset,
	amount state.Number,
	tag uint64,
	to crypto.Address,
	toTag uint64,
	toChainID uint64,
) {
	obj.Outs = append(obj.Outs, &TransferOut{
		Asset:     asset,
		Amount:    amount,
		Tag:       tag,
		To:        to,
		ToTag:     toTag,
		ToChainID: toChainID,
	})
}

func (obj *Transfer) Encode() []byte {
	return bin.Encode(
		obj.Outs,
		obj.Comment,
	)
}

func (obj *Transfer) Decode(data []byte) error {
	return bin.Decode(data,
		&obj.Outs,
		&obj.Comment,
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

func (obj *Transfer) Totals() map[string]state.Number {
	vv := map[string]state.Number{}
	for _, out := range obj.Outs {
		sAsset := out.Asset.String()
		s, ok := vv[sAsset]
		if !ok {
			s = state.Int(0)
		}
		vv[sAsset] = s.Add(s, out.Amount)
	}
	return vv
}

func (obj *Transfer) Verify(tx *blockchain.Transaction) error {
	for _, out := range obj.Outs {
		if out.Amount.Sign() <= 0 {
			return ErrTxIncorrectAmount
		}
	}
	return nil
}

func (obj *Transfer) Execute(tx *blockchain.Transaction, st *state.State) {
	for _, out := range obj.Outs {

		// decrement amount from address; panic if not enough funds
		st.Decrement(out.Asset, tx.SenderAddress(), out.Amount, out.Tag)

		// increment amount to new address
		if tx.ChainID == out.ToChainID {
			st.Increment(out.Asset, out.To, out.Amount, out.ToTag)
		} else {
			st.CrossChainSet(out.ToChainID, out.Asset, out.To, out.Amount, out.ToTag)
		}
	}
}
