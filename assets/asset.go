package assets

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
)

type Asset []byte

func (a Asset) String() string {
	return hex.EncodeToString(a)
}

func (a Asset) Type() uint8 {
	return a[0]
}

func (a Asset) IsCoin() bool {
	return a.Type() == CoinType
}

func (a Asset) IsCounter() bool {
	return a.Type() == CounterType
}

func (a Asset) IsName() bool {
	return a.Type() == NameType
}

func (a Asset) CoinCounter(counterID string) Asset {
	return NewCounter(a[1], counterID)
}

func (a Asset) CounterID() string {
	// if !a.IsCounter() || len(a) < 2 panic()
	return string(a[2:])
}

func (a Asset) CounterType() uint8 {
	// if !a.IsCounter() || len(a) < 2 panic()
	return a[1]
}

func (a Asset) Equal(b Asset) bool {
	return bytes.Equal(a, b)
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
