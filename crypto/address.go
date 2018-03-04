package crypto

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/crypto/base58"
)

const (
	AddressSize = 24 // 192 bits

	addressPrefix = "Like"
	addressVer    = 0x01
	checksumLen   = 3
)

type Address [AddressSize]byte

var NilAddress Address

var (
	errAddrInvalidLength        = errors.New("crypto.Address: Invalid address length")
	errParseAddrInvalid         = errors.New("crypto.ParseAddress: Invalid address")
	errParseAddrUnknownVer      = errors.New("crypto.ParseAddress: Unknown address version")
	errParseAddrInvalidCheckSum = errors.New("crypto.ParseAddress: Invalid check-sum")
)

func newAddress(data []byte) (addr Address) {
	if len(data) != AddressSize {
		panic(errAddrInvalidLength)
	}
	copy(addr[:], data[:AddressSize])
	return
}

func (addr Address) String() string {
	return addr.TaggedString(0)
}

func (addr Address) IsNil() bool {
	return addr == NilAddress
}

func (addr Address) Equal(a Address) bool {
	return addr == a
}

func (addr Address) ID() uint64 {
	return bin.BytesToUint64(addr[:8])
}

func (addr Address) Encode() []byte {
	if addr.IsNil() {
		return nil
	}
	return addr[:]
}

func (addr *Address) Decode(data []byte) error {
	if len(data) == 0 { // is nil address
		*addr = Address{}
		return nil
	}
	if len(data) != AddressSize {
		return errAddrInvalidLength
	}
	copy(addr[:], data[:AddressSize])
	return nil
}

func addrCheckSum(addr []byte, tag int64) []byte {
	h := newHash256()
	h.Write([]byte(addressPrefix))
	h.Write([]byte{addressVer})
	h.Write(addr)
	h.Write(bin.Uint64ToBytes(uint64(tag)))
	return hash256(h.Sum(nil))[:checksumLen]
}

func (addr Address) TaggedString(tag int64) string {
	if addr.IsNil() {
		return ""
	}
	w := bin.NewBuffer(nil)
	w.WriteByte(addressVer) // first byte have to > 0
	w.Write(addr[:])
	w.WriteVarInt64(tag)
	w.Write(addrCheckSum(addr[:], tag))
	return addressPrefix + base58.Encode(w.Bytes())
}

func (addr Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(addr.String())
}

func (addr *Address) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err = json.Unmarshal(data, &s); err != nil {
		return
	}
	if a, _, err := ParseAddress(s); err != nil {
		return err
	} else {
		copy(addr[:], a[:])
	}
	return
}

func ParseAddress(strAddr string) (addr Address, tag int64, err error) {
	if !strings.HasPrefix(strAddr, addressPrefix) {
		err = errParseAddrUnknownVer
		return
	}
	bb, err := base58.Decode(strAddr[len(addressPrefix):])
	if err != nil {
		return
	}
	r := bin.NewBuffer(bb)
	if b, e := r.ReadByte(); b != addressVer || e != nil {
		err = errParseAddrInvalid
		return
	}
	if _, err = r.Read(addr[:]); err != nil {
		err = errParseAddrInvalid
		return
	}
	if tag, err = r.ReadVarInt64(); err != nil {
		err = errParseAddrInvalid
		return
	}
	chSum := make([]byte, checksumLen)
	_, err = r.Read(chSum[:])
	if err != nil || !bytes.Equal(chSum, addrCheckSum(addr[:], tag)) {
		err = errParseAddrInvalidCheckSum
		return
	}
	return
}

func MustParseAddress(strAddr string) Address {
	addr, _, err := ParseAddress(strAddr)
	if err != nil {
		panic(err)
	}
	return addr
}
