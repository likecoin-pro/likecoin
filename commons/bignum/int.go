package bignum

import (
	"encoding/json"
	"math/big"

	"github.com/denisskin/bin"
)

type Int struct {
	i *big.Int
}

func NewInt(i int64) Int {
	if i == 0 {
		return Int{}
	}
	return Int{big.NewInt(i)}
}

func (x Int) Int64() int64 {
	if x.i == nil {
		return 0
	}
	return x.i.Int64()
}

func (x Int) String() string {
	if x.i == nil {
		return "0"
	}
	return x.i.String()
}

func (x Int) Hex() string {
	if x.i == nil {
		return "0"
	}
	return x.i.Text(16)
}

func (x Int) Bytes() []byte {
	if x.i == nil {
		return nil
	}
	return x.i.Bytes()
}

func (x Int) BinWrite(w *bin.Writer) {
	w.WriteBigInt(x.i)
}

func (x *Int) BinRead(r *bin.Reader) {
	x.i, _ = r.ReadBigInt()
}

func (x Int) MarshalJSON() ([]byte, error) {
	return []byte(x.String()), nil
}

func (x *Int) UnmarshalJSON(data []byte) (err error) {
	return json.Unmarshal(data, &x.i)
}

func (x Int) IsZero() bool {
	return x.i == nil || x.i.Sign() == 0
}

func (x Int) BigInt() *big.Int {
	if x.i == nil {
		return new(big.Int)
	}
	return new(big.Int).Set(x.i)
}

func (x Int) Sign() int {
	if x.i == nil {
		return 0
	}
	return x.i.Sign()
}

func (x Int) Equal(y Int) bool {
	if x.IsZero() {
		return y.IsZero()
	}
	return y.i != nil && x.i.Cmp(y.i) == 0
}

func (x Int) Cmp(y Int) int {
	if x.i == nil {
		x.i = new(big.Int)
	}
	if y.i == nil {
		y.i = new(big.Int)
	}
	return x.i.Cmp(y.i)
}

func (x Int) Neg() Int {
	if x.i == nil {
		return Int{}
	}
	return Int{new(big.Int).Neg(x.i)}
}

func (x Int) Min(y Int) Int {
	if x.Cmp(y) <= 0 {
		return x
	}
	return y
}

func (x Int) Max(y Int) Int {
	if x.Cmp(y) >= 0 {
		return x
	}
	return y
}

func (x Int) Add(y Int) Int {
	if x.i == nil {
		return y
	} else if y.i == nil {
		return x
	}
	return Int{new(big.Int).Add(x.i, y.i)}
}

func (x Int) Sub(y Int) Int {
	if y.IsZero() {
		return x
	} else if x.IsZero() {
		return y.Neg()
	}
	return Int{new(big.Int).Sub(x.i, y.i)}
}

func (x Int) Mul(y Int) Int {
	if x.IsZero() || y.IsZero() {
		return Int{}
	}
	return Int{new(big.Int).Mul(x.i, y.i)}
}

func (x Int) Div(y Int) Int {
	if x.IsZero() {
		return Int{}
	}
	if y.IsZero() {
		panic("big: division by zero")
	}
	return Int{new(big.Int).Div(x.i, y.i)}
}

func (x Int) Mod(y Int) Int {
	if x.IsZero() {
		return Int{}
	}
	if y.IsZero() {
		panic("big: division by zero")
	}
	return Int{new(big.Int).Mod(x.i, y.i)}
}
