package crypto

import (
	"encoding/hex"
	"encoding/json"
)

type Asset []byte

func (a Asset) String() string {
	return hex.EncodeToString(a)
}

func (a Asset) Encode() []byte {
	return a
}

func (a *Asset) Decode(data []byte) (err error) {
	*a = data
	return nil
}

func (a Asset) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func ParseAsset(s string) (Asset, error) {
	data, err := hex.DecodeString(s)
	return Asset(data), err
}

//func DecodeAddressAsset(strAddr, strAsset string) (addr Address, asset []byte, err error) {
//	if addr, _, err = ParseAddress(strAddr); err == nil {
//		asset, err = ParseAsset(strAsset)
//	}
//	return
//}
