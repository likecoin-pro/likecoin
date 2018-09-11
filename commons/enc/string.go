package enc

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// String returns object as string (encode to json)
func String(v interface{}) string {
	switch s := v.(type) {
	case string:
		return s
	case fmt.Stringer:
		return s.String()
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

// JSON returns object as json-string
func JSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func IndentJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

var sizeSfxs = []string{" B", " KB", " MB", " GB", " TB", " PB", " EB"}

func BinarySizeStr(size int64) string {
	f, sfx, pfx := float64(size), 0, ""
	if f < 0 {
		f, pfx = -f, "-"
	}
	for f >= 1000 {
		f /= 1024
		sfx++
	}
	v := strconv.FormatFloat(f, 'f', 3, 64)
	if len(v) > 4 {
		v = v[:4]
	}
	if strings.IndexByte(v, '.') > 0 {
		for v[len(v)-1] == '0' { // trim right '0'
			v = v[:len(v)-1]
		}
	}
	if v[len(v)-1] == '.' { // trim right '.'
		v = v[:len(v)-1]
	}
	return pfx + v + sizeSfxs[sfx]
}
