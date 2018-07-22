package bignum

import (
	"encoding/json"
	"testing"

	"github.com/denisskin/bin"
	"github.com/stretchr/testify/assert"
)

func TestInt_MarshalJSON(t *testing.T) {
	var x = NewInt(0x7fffffffffffffff).Mul(NewInt(0x7fffffffffffffff))

	data, err := json.Marshal(x)

	assert.NoError(t, err)
	assert.Equal(t, "85070591730234615847396907784232501249", string(data))
}

func TestInt_UnmarshalJSON(t *testing.T) {
	data := []byte(`85070591730234615847396907784232501249`)

	var x Int
	err := json.Unmarshal(data, &x)

	assert.NoError(t, err)
	assert.Equal(t, "85070591730234615847396907784232501249", x.String())
}

func TestInt_IsZero(t *testing.T) {
	var x Int
	var y = NewInt(2).Sub(NewInt(2))
	var z = NewInt(0)

	assert.True(t, x.IsZero())
	assert.True(t, y.IsZero())
	assert.True(t, z.IsZero())
}

func TestInt_Equal(t *testing.T) {
	var x = NewInt(123)
	var y = NewInt(123)
	var z Int

	assert.True(t, x.Equal(x))
	assert.True(t, x.Equal(y))
	assert.True(t, y.Equal(x))
	assert.True(t, z.Equal(Int{}))
	assert.True(t, z.Equal(NewInt(0)))
	assert.True(t, x.Sub(y).Equal(z))
	assert.Equal(t, x, y)
	assert.Equal(t, y, x)
}

func TestInt_Neg(t *testing.T) {
	var x = NewInt(123)

	y := x.Neg()

	assert.Equal(t, NewInt(123), x)
	assert.Equal(t, NewInt(-123), y)
}

func TestInt_Sign(t *testing.T) {
	var x Int
	var y = NewInt(42)
	var z = NewInt(-42)

	assert.Equal(t, 0, x.Sign())
	assert.Equal(t, +1, y.Sign())
	assert.Equal(t, -1, z.Sign())
}

func TestInt_Add(t *testing.T) {
	var x = NewInt(2)
	var y = NewInt(3)

	z := x.Add(y)

	assert.Equal(t, NewInt(2), x)
	assert.Equal(t, NewInt(3), y)
	assert.Equal(t, NewInt(5), z)
}

func TestInt_Mul(t *testing.T) {
	var x = NewInt(2)
	var y = NewInt(-3)

	z := x.Mul(y)

	assert.Equal(t, NewInt(2), x)
	assert.Equal(t, NewInt(-3), y)
	assert.Equal(t, NewInt(-6), z)
}

func TestInt_BinWrite(t *testing.T) {
	var x Int
	var y = NewInt(123456)

	data0 := bin.Encode(x)
	data1 := bin.Encode(y)

	assert.Equal(t, []byte{0}, data0)
	assert.Equal(t, bin.Encode(123456), data1)
}

func TestInt_BinRead(t *testing.T) {
	data := bin.Encode(-0x7fffffffffffffff)

	var x Int
	err := bin.Decode(data, &x)

	assert.NoError(t, err)
	assert.EqualValues(t, -0x7fffffffffffffff, x.Int64())
}
