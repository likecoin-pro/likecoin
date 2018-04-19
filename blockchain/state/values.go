package state

import "github.com/likecoin-pro/likecoin/crypto/merkle"

type Values []*Value

func (vv Values) String() string {
	s := "[\n"
	for _, v := range vv {
		s += "  " + v.String() + ",\n"
	}
	return s + "]"
}

func (vv Values) Equal(b Values) bool {
	if len(vv) != len(b) {
		return false
	}
	for i, v := range vv {
		if !v.Equal(b[i]) {
			return false
		}
	}
	return true
}

func (vv Values) Hash() []byte {
	hh := make([][]byte, len(vv))
	for i, v := range vv {
		hh[i] = v.Hash()
	}
	return merkle.Root(hh...)
}
