package state

import (
	"encoding/json"
	"testing"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/commons/bignum"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/stretchr/testify/assert"
)

var (
	coin = assets.Likecoin

	addr0 = crypto.MustParseAddress("Like5QU4UiVeh7459hiRPJZfh9AhoFN29Pn9bwwSGvJ")
	addrA = crypto.MustParseAddress("Like46j9pghNTUaxd2zRbJfruQt8U9AWMd3Rb5bJxLj")
	addrB = crypto.MustParseAddress("Like3mJ34d7obc37tS1Rx4aqtZdPew3n61xwTeWV9Am")
	addrC = crypto.MustParseAddress("Like52qhrYBtJp5k8dDhYJw1p34E97SQicrjDrRp3Fj")
)

func exec(fn func()) (err error) {
	defer func() { err, _ = recover().(error) }()
	fn()
	return
}

func (s *State) init(addr crypto.Address, v int64) *State {
	s.vals[strKey(coin, addr)] = bignum.NewInt(v)
	return s
}

func TestState_Get(t *testing.T) {

	st := NewState(0, nil).init(addrA, 10)

	v0 := st.Get(coin, addr0)
	v1 := st.Get(coin, addrA)

	assert.EqualValues(t, 0, v0.Int64())
	assert.EqualValues(t, 10, v1.Int64())
}

func TestState_Get_(t *testing.T) {
	a := NewState(0, nil).init(addrA, 666)
	a.Get(coin, addrA)
	a.Increment(coin, addr0, bignum.NewInt(123), 0)

	b := NewState(0, nil).init(addr0, 100).init(addrA, 333)
	b.Get(coin, addrB)
	b.Increment(coin, addr0, bignum.NewInt(23), 0)
	b.Get(coin, addrC)

	c := NewState(0, nil).init(addr0, 123)
	c.Get(coin, addr0)

	assert.True(t, a.Values().Equal(b.Values()))
	assert.True(t, b.Values().Equal(a.Values()))
	assert.False(t, c.Values().Equal(a.Values()))
}

func TestValues_Equal(t *testing.T) {

	a := NewState(0, nil)
	b := NewState(0, nil)
	c := NewState(0, nil)

	a.Increment(coin, addr0, bignum.NewInt(11), 0)
	b.Increment(coin, addr0, bignum.NewInt(11), 0)
	c.Increment(coin, addr0, bignum.NewInt(11), 22)

	a.Increment(coin, addrA, bignum.NewInt(22), 0)
	b.Increment(coin, addrA, bignum.NewInt(22), 0)
	c.Increment(coin, addrA, bignum.NewInt(22), 0)

	assert.True(t, a.Values().Equal(b.Values()))
	assert.True(t, b.Values().Equal(a.Values()))
	assert.False(t, c.Values().Equal(a.Values()))
}

func TestState_Increment(t *testing.T) {
	st := NewState(0, nil).init(addrA, 10)

	err := exec(func() {
		st.Increment(coin, addr0, bignum.NewInt(1), 0)
		st.Increment(coin, addrA, bignum.NewInt(1), 0)
		st.Decrement(coin, addrA, bignum.NewInt(2), 0)
	})

	v0 := st.Get(coin, addr0)
	vA := st.Get(coin, addrA)

	assert.NoError(t, err)
	assert.EqualValues(t, 1, v0.Int64())
	assert.EqualValues(t, 9, vA.Int64())
}

func TestState_Decrement_fail(t *testing.T) {
	st := NewState(0, nil).init(addrA, 10)

	err0 := exec(func() { st.Decrement(coin, addr0, bignum.NewInt(1), 0) })
	err1 := exec(func() { st.Decrement(coin, addrA, bignum.NewInt(1), 0) })
	err2 := exec(func() { st.Decrement(coin, addrA, bignum.NewInt(10), 0) })
	v0 := st.Get(coin, addr0)
	vA := st.Get(coin, addrA)

	assert.Error(t, err0)
	assert.NoError(t, err1)
	assert.Error(t, err2)
	assert.Equal(t, bignum.NewInt(0), v0)
	assert.Equal(t, bignum.NewInt(9), vA)
}

func TestValue_Encode(t *testing.T) {
	s1 := NewState(0, nil).init(addr0, 12)
	s1.Increment(coin, addrA, bignum.NewInt(34), 0)
	s1.Increment(coin, addrB, bignum.NewInt(56), 0)
	data1 := bin.Encode(s1.Values())

	var s2 Values
	err2 := bin.Decode(data1, &s2)
	data2 := bin.Encode(s2)

	assert.NoError(t, err2)
	assert.Equal(t, data1, data2)
}

func TestValue_Decode(t *testing.T) {
	s := NewState(0, nil).init(addrA, 10).init(addrB, 10)
	s.Increment(coin, addr0, bignum.NewInt(1), 0)
	s.Decrement(coin, addrA, bignum.NewInt(10), 0)
	data := bin.Encode(s.Values())

	var vv Values
	err := bin.Decode(data, &vv)

	assert.NoError(t, err)
	assert.JSONEq(t, enc.JSON(Values{
		{Address: addr0, Asset: coin, Balance: bignum.NewInt(1)},
		{Address: addrA, Asset: coin, Balance: bignum.NewInt(0)},
	}), enc.JSON(vv))
}

func TestValue_MarshalJSON(t *testing.T) {
	st := NewState(0, nil).init(addrA, 123)
	st.Increment(coin, addr0, bignum.NewInt(1), 111)
	st.Get(coin, addrC)
	st.Increment(coin, addrB, bignum.NewInt(100), 222)
	st.Get(coin, addrA)

	data, err := json.Marshal(st.Values())

	assert.NoError(t, err)
	assert.JSONEq(t, `[
	  {
		"chain":   0,
		"address": "Like5QU4UiVeh7459hiRPJZfh9AhoFN29Pn9bwwSGvJ",
		"asset":   "0x0001",
		"balance": 1,
		"memo":    111
	  },
	  {
		"chain":   0,
		"address": "Like3mJ34d7obc37tS1Rx4aqtZdPew3n61xwTeWV9Am",
		"asset":   "0x0001",
		"balance": 100,
		"memo":    222
	  }
	]`, string(data))
}

func TestState_Values(t *testing.T) {
	st := NewState(0, nil).init(addrA, 10).init(addrB, 5).init(addrC, 1)

	err := exec(func() {
		st.Increment(coin, addr0, bignum.NewInt(1), 111)
		st.Get(coin, addrA)
		st.Decrement(coin, addrB, bignum.NewInt(2), 222)
		st.Decrement(coin, addrB, bignum.NewInt(3), 333)
		st.Get(coin, addrC)
		st.Increment(coin, addr0, bignum.NewInt(3), 333)
	})
	values := st.Values()

	assert.NoError(t, err)
	assert.JSONEq(t, `[
	  {
		"chain":   0,
		"asset":   "0x0001",
		"address": "Like5QU4UiVeh7459hiRPJZfh9AhoFN29Pn9bwwSGvJ",
		"balance": 1,
		"memo":    111
	  },
	  {
		"chain":   0,
		"asset":   "0x0001",
		"address": "Like3mJ34d7obc37tS1Rx4aqtZdPew3n61xwTeWV9Am",
		"balance": 3,
		"memo":    222
	  },
	  {
		"chain":   0,
		"asset":   "0x0001",
		"address": "Like3mJ34d7obc37tS1Rx4aqtZdPew3n61xwTeWV9Am",
		"balance": 0,
		"memo":    333
	  },
	  {
		"chain":   0,
		"asset":   "0x0001",
		"address": "Like5QU4UiVeh7459hiRPJZfh9AhoFN29Pn9bwwSGvJ",
		"balance": 4,
		"memo":    333
	  }
	]`, enc.JSON(values))
}
