package enc

import (
	"encoding/json"
	"fmt"
	"xnet/std/consts"
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

func DataSizeToString(size int64) string {
	switch {
	case size >= consts.EiB:
		return fmt.Sprintf("%.2f EiB", float64(size)/float64(consts.EiB))

	case size >= consts.PiB:
		return fmt.Sprintf("%.2f PiB", float64(size)/float64(consts.PiB))

	case size >= consts.TiB:
		return fmt.Sprintf("%.2f TiB", float64(size)/float64(consts.TiB))

	case size >= consts.GiB:
		return fmt.Sprintf("%.2f GiB", float64(size)/float64(consts.GiB))

	case size >= consts.MiB:
		return fmt.Sprintf("%.2f MiB", float64(size)/float64(consts.MiB))

	case size >= consts.KiB:
		return fmt.Sprintf("%.2f KiB", float64(size)/float64(consts.KiB))

	default:
		return fmt.Sprintf("%d B", size)
	}
}
