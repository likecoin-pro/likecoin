package crypto

import (
	"encoding/hex"
	"encoding/json"
)

type Asset []byte

func (a Asset) String() string {
	return hex.EncodeToString(a)
}

func (a Asset) IsCounter() bool {
	return len(a) > 1 && a[0] == 0
}

func (a Asset) CounterAsset() Asset {
	c := make(Asset, len(a)+1)
	copy(c[1:], a)
	return c
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

func (a *Asset) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err = json.Unmarshal(data, &s); err != nil {
		return err
	}
	*a, err = hex.DecodeString(s)
	return
}

func ParseAsset(s string) (Asset, error) {
	data, err := hex.DecodeString(s)
	return Asset(data), err
}
