package hex

import (
	"encoding/hex"
	"encoding/json"

	"github.com/denisskin/bin"
)

type Bytes []byte

func (b Bytes) String() string {
	return hex.EncodeToString(b)
}

func (b Bytes) BinWrite(w *bin.Writer) {
	w.WriteBytes([]byte(b))
}

func (b *Bytes) BinRead(r *bin.Reader) {
	*b, _ = r.ReadBytes()
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
