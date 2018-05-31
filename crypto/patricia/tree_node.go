package patricia

import (
	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/crypto/merkle"
)

type node struct {
	key    []byte
	value  []byte
	hashes [16][]byte
}

func (nd *node) hash() []byte {
	if nd.key != nil || nd.value != nil {
		return merkle.Root(nd.key, nd.value)
	}
	var hh [][]byte // no empty hashes
	for _, b := range nd.hashes {
		if b != nil {
			hh = append(hh, b)
		}
	}
	return merkle.Root(hh...)
}

func (nd *node) proof(iHash int) []byte {
	var hh [][]byte
	var idx int
	for i, b := range nd.hashes {
		if b != nil {
			if iHash == i {
				idx = len(hh)
			}
			hh = append(hh, b)
		}
	}
	proof, _ := merkle.Proof(hh, idx)
	return proof
}

func (nd *node) encode() (data []byte) {
	w := bin.NewBuffer(nil)
	w.WriteBytes(nd.key)
	w.WriteBytes(nd.value)

	var ff uint16
	for i, b := range nd.hashes {
		if len(b) != 0 {
			ff |= 1 << uint16(i)
		}
	}
	w.WriteUint16(ff)
	for _, b := range nd.hashes {
		if len(b) != 0 {
			w.Write(b)
		}
	}
	return w.Bytes()
}

func (nd *node) decode(data []byte) error {
	r := bin.NewBuffer(data)
	nd.key, _ = r.ReadBytes()
	nd.value, _ = r.ReadBytes()
	ff, _ := r.ReadUint16()
	for i := range nd.hashes {
		if ff&(1<<uint16(i)) != 0 {
			nd.hashes[i] = make([]byte, merkle.HashSize)
			r.Read(nd.hashes[i])
		}
	}
	if r.Error() != nil {
		return errInvalidNodeData
	}
	return nil
}
