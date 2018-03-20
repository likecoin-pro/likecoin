package patricia

import "github.com/likecoin-pro/likecoin/crypto/merkle"

type node struct {
	val    []byte
	hashes [16][]byte
}

func (nd *node) hash() []byte {
	if nd.val != nil {
		return nd.val
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
	_, proof := merkle.Proof(hh, idx)
	return proof
}

func (nd *node) encode() (data []byte) {
	data = make([]byte, 2)
	var ff uint32
	for i, b := range nd.hashes {
		if len(b) != 0 {
			ff |= 1 << uint32(i)
			data = append(data, b...)
		}
	}
	if ff == 0 {
		data = append(data, nd.val...)
	} else {
		data[0] = uint8(ff >> 8)
		data[1] = uint8(ff)
	}
	return
}

func (nd *node) decode(data []byte) (n int, err error) {
	if len(data) < 2 {
		err = errInvalidNodeData
		return
	}
	ff, data := (uint32(data[0])<<8)|uint32(data[1]), data[2:] // flags
	n += 2
	if ff == 0 {
		nd.val = data
		n += len(data)
		return
	}
	nd.val = nil
	for i := range nd.hashes {
		if ff&(1<<uint32(i)) != 0 {
			if len(data) < merkle.HashSize {
				err = errInvalidNodeData
				return
			}
			nd.hashes[i], data = data[:merkle.HashSize], data[merkle.HashSize:]
			n += merkle.HashSize
		}
	}
	return
}
