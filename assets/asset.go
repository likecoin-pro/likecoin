package assets

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"strings"
)

type Asset []byte

func (a Asset) String() string {
	if len(a) > 0 {
		return "0x" + hex.EncodeToString(a)
	}
	return ""
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

func (a Asset) ID() uint8 {
	return a[1]
}

func (a Asset) SourceCounter(counterID string) Asset {
	return NewCounter(a[1], counterID)
}

func (a Asset) Empty() bool {
	return len(a) == 0
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
	*a, err = ParseAsset(s)
	return
}

func ParseAsset(s string) (Asset, error) {
	s = strings.TrimPrefix(s, "0x")
	data, err := hex.DecodeString(s)
	return Asset(data), err
}
