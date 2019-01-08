package crypto

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"encoding/hex"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/crypto/base58"
	"golang.org/x/crypto/sha3"
)

const (
	AddressLength = 24 // 192 bits

	addressPrefix = "Like"
	addressVer    = 0x01
	checksumLen   = 3
)

type Address [AddressLength]byte

var NilAddress Address

var (
	errAddrInvalidLength        = errors.New("crypto.Address: Invalid address length")
	errParseAddrInvalid         = errors.New("crypto.ParseAddress: Invalid address")
	errParseAddrUnknownVer      = errors.New("crypto.ParseAddress: Unknown address version")
	errParseAddrInvalidCheckSum = errors.New("crypto.ParseAddress: Invalid check-sum")
)

func newAddress(data []byte) (addr Address) {
	if len(data) != AddressLength {
		panic(errAddrInvalidLength)
	}
	copy(addr[:], data[:AddressLength])
	return
}

func (addr Address) String() string {
	return addr.MemoString(0)
}

func (addr Address) Empty() bool {
	return addr == NilAddress
}

func (addr Address) Equal(a Address) bool {
	return addr == a
}

func (addr Address) ID() uint64 {
	return bin.BytesToUint64(addr[:8])
}

func (addr Address) Hex() string {
	return "0x" + hex.EncodeToString(addr[:])
}

func (addr Address) Encode() []byte {
	if addr.Empty() {
		return nil
	}
	return addr[:]
}

func (addr *Address) Decode(data []byte) error {
	if len(data) == 0 { // is nil address
		*addr = Address{}
		return nil
	}
	if len(data) != AddressLength {
		return errAddrInvalidLength
	}
	copy(addr[:], data)
	return nil
}

func addrCheckSum(addr []byte, memo uint64) []byte {
	sha := sha3.NewShake256()
	sha.Write([]byte(addressPrefix))
	sha.Write([]byte{addressVer})
	sha.Write(addr)
	sha.Write(bin.Uint64ToBytes(memo))
	var h [checksumLen]byte
	sha.Read(h[:])
	return h[:]
}

func (addr Address) MemoString(memo uint64) string {
	if addr.Empty() {
		return ""
	}
	w := bin.NewBuffer(nil)
	w.WriteByte(addressVer) // first byte have to > 0
	w.Write(addr[:])
	w.WriteVarUint64(memo)
	w.Write(addrCheckSum(addr[:], memo))
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
	if s == "" { // nil addr
		*addr = NilAddress
		return nil
	}
	if a, _, err := ParseAddress(s); err != nil {
		return err
	} else {
		copy(addr[:], a[:])
	}
	return
}

func ParseAddress(strAddr string) (addr Address, memo uint64, err error) {
	defer func() {
		if err != nil {
			addr = NilAddress
		}
	}()

	// hex format of address
	if strings.HasPrefix(strAddr, "0x") {
		if buf, e := hex.DecodeString(strAddr[2:]); e == nil && (len(buf) == AddressLength || len(buf) == AddressLength+8) {
			copy(addr[:], buf[:AddressLength])
			memo = bin.BytesToUint64(buf[AddressLength:])
		} else {
			err = errParseAddrInvalid
		}
		return
	}

	if i := strings.Index(strAddr, "0"); i > 0 {
		if addr, _, err = ParseAddress(strAddr[:i]); err != nil {
			return
		}
		memo, err = strconv.ParseUint(strAddr[i:], 16, 64)
		return
	}

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
	if memo, err = r.ReadVarUint64(); err != nil {
		err = errParseAddrInvalid
		return
	}
	chSum := make([]byte, checksumLen)
	_, err = r.Read(chSum[:])
	if err != nil || !bytes.Equal(chSum, addrCheckSum(addr[:], memo)) {
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
