package transaction

import (
	"errors"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/commons/hex"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
)

type Header struct {
	Type    Type      `json:"type"`    // tx type
	Version int       `json:"version"` // tx version
	Network int       `json:"network"` //
	ChainID uint64    `json:"chain"`   //
	Nonce   uint64    `json:"nonce"`   //
	Sig     hex.Bytes `json:"sig"`     // tx signature

	// not imported fields
	Sender *crypto.PublicKey `json:"sender"` // sender of tx (get from signature)
}

var ErrTxIncorrectSig = errors.New("transaction: Incorrect signature")

func NewHeader(txType Type, txVersion int) Header {
	return Header{
		Type:    txType,
		Version: txVersion,
		Network: config.NetworkID,
		ChainID: config.ChainID,
	}
}

func (h *Header) Sign(tx Transaction, prv *crypto.PrivateKey) {
	//h.Nonce = 0
	h.Sender = prv.PublicKey
	h.Sig = prv.Sign(tx.Hash())
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
		h.Nonce,
	)
}

func (h Header) Encode() []byte {
	return bin.Encode(
		h.Type,
		h.Version,
		h.Network,
		h.ChainID,
		h.Nonce,
		h.Sig,
	)
}
func (h *Header) Decode(data []byte) error {
	return bin.Decode(data,
		&h.Type,
		&h.Version,
		&h.Network,
		&h.ChainID,
		&h.Nonce,
		&h.Sig,
	)
}

// recoverSender recovers sender`s pubkey from signature
func (h *Header) recoverSender(tx Transaction) (err error) {
	h.Sender, err = crypto.RecoverPublicKey(tx.Hash(), h.Sig)
	return
}

func (h *Header) VerifySign(tx Transaction) (err error) {
	if h.Sender == nil {
		if err = h.recoverSender(tx); err != nil {
			return
		}
	}
	if !h.Sender.Verify(tx.Hash(), h.Sig) {
		return ErrTxIncorrectSig
	}
	return nil
}
