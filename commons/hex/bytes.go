package hex

import (
	"encoding/hex"
	"encoding/json"
	"io"

	"github.com/denisskin/bin"
)

type Bytes []byte

func (b Bytes) String() string {
	return hex.EncodeToString(b)
}

func (b Bytes) BinaryEncode(w io.Writer) error {
	return bin.NewWriter(w).WriteBytes([]byte(b))
}

func (b *Bytes) BinaryDecode(r io.Reader) (err error) {
	*b, err = bin.NewReader(r).ReadBytes()
	return
}

func (b Bytes) MarshalJSON() ([]byte, error) {
	return []byte(`"` + b.String() + `"`), nil
}

func (b *Bytes) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err = json.Unmarshal(data, &s); err != nil {
		return err
	}
	*b, err = hex.DecodeString(s)
	return
}
