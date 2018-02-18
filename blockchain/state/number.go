package state

import "math/big"

type Number = *big.Int

func Int(v int64) Number {
	return big.NewInt(v)
}
