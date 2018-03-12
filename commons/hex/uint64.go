package hex

import (
	"encoding/json"
	"strconv"

	"strings"

	"github.com/denisskin/bin"
)

type Uint64 uint64

func (i Uint64) String() string {
	return EncodeUint(uint64(i))
}

func (i Uint64) BinWrite(w *bin.Writer) {
	w.WriteVarUint64(uint64(i))
}

func (i *Uint64) BinRead(r *bin.Reader) {
	num, _ := r.ReadVarUint64()
	*i = Uint64(num)
}

func (i Uint64) MarshalJSON() ([]byte, error) {
	if i == 0 {
		return []byte(`""`), nil
	}
	return []byte(`"` + i.String() + `"`), nil
}

func (i *Uint64) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v := v.(type) {
	case float64:
		*i = Uint64(v)
	case string:
		if v == "" {
			return nil
		}
		v = strings.TrimPrefix(v, "0x")
		u64, err := strconv.ParseUint(v, 16, 64)
		if err != nil {
			return err
		}
		*i = Uint64(u64)
	}
	return nil
}
