package state

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/likecoin-pro/likecoin/assets"
)

type Number = *big.Int

func Int(v int64) Number {
	return big.NewInt(v)
}

func StrAmount(amount interface{}, a assets.Asset) string {
	if len(a) == 0 {
		a = assets.Default
	}
	s := fmt.Sprint(amount)
	if a.IsCoin() {
		_s := "000000000" + s
		s, _s = strings.TrimLeft(_s[:len(_s)-9], "0"), strings.TrimRight(_s[len(_s)-9:], "0")
		if s == "" {
			s = "0"
		}
		if _s != "" {
			s += "." + _s
		}
	}
	return s
}
