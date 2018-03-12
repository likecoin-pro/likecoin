package hex

import (
	"encoding/hex"
	"fmt"
	"strconv"
)

func EncodeUint(num uint64) (s string) {
	s = strconv.FormatUint(num, 16)
	if n := len(s); n < 16 {
		const s0 = "00000000000000000000"
		s = s0[:16-n] + s
	}
	return
}

func Decode(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

func Encode(v interface{}, params ...int) string {
	defer func() {
		recover()
	}()
	var param int
	if len(params) > 0 {
		param = params[0]
	}
	switch v := v.(type) {
	case []byte:
		if param > 0 {
			if n := len(v); n > param {
				v = v[:param]
			} else if n < param {
				v = append(make([]byte, param-n, param), v...)
			}
		}
		return hex.EncodeToString(v)

	case int32:
		return EncodeUint(uint64(v))[8:]
	case uint32:
		return EncodeUint(uint64(v))[8:]
	case int:
		return EncodeUint(uint64(v))
	case uint:
		return EncodeUint(uint64(v))
	case int64:
		return EncodeUint(uint64(v))
	case uint64:
		return EncodeUint(uint64(v))

	}
	return fmt.Sprintf("%s", v)
}
