package crypto

import (
	"bytes"
	"errors"
	"strings"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/crypto/base58"
)

const (
	AddressSize = 24 // 192 bits

	addressPrefix = "Like"
	addressVer    = '\xf3'
	checksumLen   = 3
)

type Address [AddressSize]byte

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
	return addr.ExtendedString(0)
}

func (addr Address) Encode() []byte {
	return addr[:]
}

func (addr *Address) Decode(data []byte) error {
	if len(data) != AddressSize {
		return errAddrInvalidLength
	}
	copy(addr[:], data[:AddressSize])
	return nil
}

func addrCheckSum(addr []byte, mgNum int64) []byte {
	h := newHash256()
	h.Write([]byte(addressPrefix))
	h.Write([]byte{addressVer})
	h.Write(addr)
	h.Write(bin.Uint64ToBytes(uint64(mgNum)))
	return hash256(h.Sum(nil))[:checksumLen]
}

func (addr Address) ExtendedString(magicNum int64) string {
	w := bin.NewBuffer(nil)
	w.WriteByte(addressVer) // first byte have to > 0
	w.Write(addr[:])
	w.WriteVarInt64(magicNum)
	w.Write(addrCheckSum(addr[:], magicNum))
	return addressPrefix + base58.Encode(w.Bytes())
}

func ParseAddress(strAddr string) (addr Address, magicNum int64, err error) {
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
	if magicNum, err = r.ReadVarInt64(); err != nil {
		err = errParseAddrInvalid
		return
	}
	chSum := make([]byte, checksumLen)
	_, err = r.Read(chSum[:])
	if err != nil || !bytes.Equal(chSum, addrCheckSum(addr[:], magicNum)) {
		err = errParseAddrInvalidCheckSum
		return
	}
	return
}
