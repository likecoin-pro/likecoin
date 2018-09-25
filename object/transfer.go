package object

import (
	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/commons/bignum"
	"github.com/likecoin-pro/likecoin/crypto"
)

type Transfer struct {
	Outs    []*TransferOut `json:"outs"`
	Comment string         `json:"comment"`
}

type TransferOut struct {
	Asset     assets.Asset   `json:"asset"`
	Amount    bignum.Int     `json:"amount"`
	Tag       uint64         `json:"tag"`
	To        crypto.Address `json:"to"`
	ToTag     uint64         `json:"to_tag"`
	ToChainID uint64         `json:"to_chain"`
}

var _ = blockchain.RegisterTxObject(TxTypeTransfer, &Transfer{})

func NewSimpleTransfer(
	cfg *blockchain.Config,
	from *crypto.PrivateKey,
	toAddr crypto.Address,
	amount bignum.Int,
	asset assets.Asset,
	comment string,
	tag uint64,
	toTag uint64,
) *blockchain.Transaction {
	tr := &Transfer{
		Comment: comment,
	}
	tr.AddOut(asset, amount, tag, toAddr, toTag, cfg.ChainID)
	return blockchain.NewTx(cfg, from, 0, tr)
}

func (obj *Transfer) AddOut(
	asset assets.Asset,
	amount bignum.Int,
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
		0, // ver
		obj.Outs,
		obj.Comment,
	)
}

func (obj *Transfer) Decode(data []byte) error {
	return bin.Decode(data,
		new(int),
		&obj.Outs,
		&obj.Comment,
	)
}

func (out *TransferOut) Encode() []byte {
	return bin.Encode(
		out.Asset,
		out.Amount,
		out.Tag,
		out.To,
		out.ToTag,
		out.ToChainID,
	)
}

func (out *TransferOut) Decode(data []byte) error {
	return bin.Decode(data,
		&out.Asset,
		&out.Amount,
		&out.Tag,
		&out.To,
		&out.ToTag,
		&out.ToChainID,
	)
}

func (obj *Transfer) Totals() map[string]bignum.Int {
	vv := map[string]bignum.Int{}
	for _, out := range obj.Outs {
		sAsset := out.Asset.String()
		vv[sAsset] = vv[sAsset].Add(out.Amount)
	}
	return vv
}

func (obj *Transfer) Verify(tx *blockchain.Transaction) error {
	sender := tx.SenderAddress()
	for _, out := range obj.Outs {
		if out.To.Empty() || out.To.Equal(sender) {
			return ErrTxIncorrectOutAddress
		}
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
