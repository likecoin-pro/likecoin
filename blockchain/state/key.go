package state

import "github.com/likecoin-pro/likecoin/crypto"

type Key struct {
	crypto.Address
	crypto.Asset
}

func NewKey(addr crypto.Address, a crypto.Asset) Key {
	return Key{addr, a}
}

func (k Key) str() string {
	return string(k.Address[:]) + string(k.Asset)
}

func (k Key) String() string {
	return k.Address.String() + ":" + k.Asset.String()
}

func (k Key) Encode() []byte {
	b := make([]byte, 0, crypto.AddressSize+len(k.Asset))
	b = append(b, k.Address.Encode()...)
	b = append(b, k.Asset.Encode()...)
	return b
}

func (k *Key) Decode(data []byte) error {
	if len(data) < crypto.AddressSize+1 {
		return ErrInvalidKey
	}
	if err := k.Address.Decode(data[:crypto.AddressSize]); err != nil {
		return err
	}
	if err := k.Asset.Decode(data[crypto.AddressSize:]); err != nil {
		return err
	}
	return nil
}
