package transaction

import (
	"errors"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/commons/hex"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
)

type Header struct {
	Type    Type              `json:"type"`      // tx type
	Version int               `json:"version"`   // tx version
	Network int               `json:"network"`   //
	ChainID uint64            `json:"chain"`     //
	Sender  *crypto.PublicKey `json:"sender"`    //
	Nonce   uint64            `json:"nonce"`     //
	Sign    hex.Bytes         `json:"signature"` //
}

var ErrTxIncorrectSign = errors.New("transaction: Incorrect signature")

func (h *Header) InitHeader(txType Type, txVersion int) {
	h.Type = txType
	h.Version = txVersion
	h.Network = config.NetworkID
	h.ChainID = config.ChainID
}

func (h *Header) SetSign(tx Transaction, prv *crypto.PrivateKey) {
	//h.Nonce = 0
	h.Sender = prv.PublicKey
	h.Sign = prv.Sign(tx.Hash())
}

func (h *Header) GetHeader() *Header {
	return h
}

func (h *Header) Address() crypto.Address {
	return h.Sender.Address()
}

func (h *Header) StrType() string {
	s, ok := txTypeNames[h.Type]
	if !ok {
		return "UntypedTx"
	}
	return s
}

func (h *Header) HeaderHash() []byte {
	return crypto.Hash256(
		h.Type,
		h.Version,
		h.Network,
		h.ChainID,
		h.Sender,
		h.Nonce,
	)
}

func (h Header) Encode() []byte {
	return bin.Encode(
		h.Type,
		h.Version,
		h.Network,
		h.ChainID,
		h.Sender,
		h.Nonce,
		h.Sign,
	)
}
func (h *Header) Decode(data []byte) error {
	return bin.Decode(data,
		&h.Type,
		&h.Version,
		&h.Network,
		&h.ChainID,
		&h.Sender,
		&h.Nonce,
		&h.Sign,
	)
}

func (h *Header) VerifySign(tx Transaction) error {
	if !h.Sender.Verify(tx.Hash(), h.Sign) {
		return ErrTxIncorrectSign
	}
	return nil
}
