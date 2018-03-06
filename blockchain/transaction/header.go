package transaction

import (
	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/config"
)

type Header struct {
	Type    Type   `json:"type"`    // tx type
	Version int    `json:"version"` // tx version
	Network int    `json:"network"` //
	ChainID uint64 `json:"chain"`   //
}

func NewHeader(txType Type, txVersion int) Header {
	return Header{
		Type:    txType,
		Version: txVersion,
		Network: config.NetworkID,
		ChainID: config.ChainID,
	}
}

func (h Header) GetHeader() Header {
	return h
}

func (h Header) Encode() []byte {
	return bin.Encode(
		h.Type,
		h.Version,
		h.Network,
		h.ChainID,
	)
}
func (h *Header) Decode(data []byte) error {
	return bin.Decode(data,
		&h.Type,
		&h.Version,
		&h.Network,
		&h.ChainID,
	)
}
