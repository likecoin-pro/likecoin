package state

import (
	"encoding/json"
	"testing"

	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/stretchr/testify/assert"
)

var (
	coin = assets.LikeCoin

	addr0 = crypto.MustParseAddress("Like3m1UbktLcKpr2uLihHakhREPX23xUgdChrZnWcK")
	addrA = crypto.MustParseAddress("Like5eBiwK1JXRTsfNAAPN5GD6zwUWjdvu5y8JXRiLJ")
	addrB = crypto.MustParseAddress("Like5T98kZKvq49awa7awjHWvD25wkJKMQD7g6Q5X9r")
	addrC = crypto.MustParseAddress("Like5LtpZeGE5Ve5NbjCeraFVY5GCcTSFv9FYDzb7nm")
)

func exec(fn func()) (err error) {
	defer func() { err, _ = recover().(error) }()
	fn()
	return
}

func (s *State) init(addr crypto.Address, v int64) *State {
	s.vals[strKey(coin, addr)] = Int(v)
	return s
}

func TestState_Get(t *testing.T) {

	st := NewState(0, nil).init(addrA, 10)

	v0 := st.Get(coin, addr0)
	v1 := st.Get(coin, addrA)

	assert.Equal(t, Int(0), v0)
	assert.Equal(t, Int(10), v1)
}

func TestState_Get_(t *testing.T) {
	a := NewState(0, nil).init(addrA, 666)
	a.Get(coin, addrA)
	a.Increment(coin, addr0, Int(123), 0)

	b := NewState(0, nil).init(addr0, 100).init(addrA, 333)
	b.Get(coin, addrB)
	b.Increment(coin, addr0, Int(23), 0)
	b.Get(coin, addrC)

	c := NewState(0, nil).init(addr0, 123)
	c.Get(coin, addr0)

	assert.True(t, a.Equal(b))
	assert.True(t, b.Equal(a))
	assert.False(t, c.Equal(a))
}

func TestState_Equal(t *testing.T) {

	a := NewState(0, nil)
	b := NewState(0, nil)
	c := NewState(0, nil)

	a.Increment(coin, addr0, Int(11), 0)
	b.Increment(coin, addr0, Int(11), 0)
	c.Increment(coin, addr0, Int(11), 22)

	a.Increment(coin, addrA, Int(22), 0)
	b.Increment(coin, addrA, Int(22), 0)
	c.Increment(coin, addrA, Int(22), 0)

	assert.True(t, a.Equal(b))
	assert.True(t, b.Equal(a))
	assert.False(t, c.Equal(a))
}

func TestState_Increment(t *testing.T) {
	st := NewState(0, nil).init(addrA, 10)

	err := exec(func() {
		st.Increment(coin, addr0, Int(1), 0)
		st.Increment(coin, addrA, Int(1), 0)
		st.Decrement(coin, addrA, Int(2), 0)
	})

	v0 := st.Get(coin, addr0)
	vA := st.Get(coin, addrA)

	assert.NoError(t, err)
	assert.Equal(t, Int(1), v0)
	assert.Equal(t, Int(9), vA)
}

func TestState_Decrement_fail(t *testing.T) {
	st := NewState(0, nil).init(addrA, 10)

	err0 := exec(func() { st.Decrement(coin, addr0, Int(1), 0) })
	err1 := exec(func() { st.Decrement(coin, addrA, Int(1), 0) })
	err2 := exec(func() { st.Decrement(coin, addrA, Int(10), 0) })
	v0 := st.Get(coin, addr0)
	vA := st.Get(coin, addrA)

	assert.Error(t, err0)
	assert.NoError(t, err1)
	assert.Error(t, err2)
	assert.Equal(t, Int(0), v0)
	assert.Equal(t, Int(9), vA)
}

func TestState_Encode(t *testing.T) {
	s1 := NewState(0, nil).init(addr0, 12)
	s1.Increment(coin, addrA, Int(34), 0)
	s1.Increment(coin, addrB, Int(56), 0)
	data1 := s1.Encode()

	var s2 = new(State)
	err2 := s2.Decode(data1)
	data2 := s2.Encode()

	assert.NoError(t, err2)
	assert.Equal(t, data1, data2)
}

func TestState_Decode(t *testing.T) {
	s := NewState(0, nil).init(addrA, 10).init(addrB, 10)
	s.Increment(coin, addr0, Int(1), 0)
	s.Decrement(coin, addrA, Int(10), 0)
	data := s.Encode() // encode only changed values

	st := NewState(0, nil)
	err := st.Decode(data)
	v0 := st.Get(coin, addr0)
	vA := st.Get(coin, addrA)
	vB := st.Get(coin, addrB)

	assert.NoError(t, err)
	assert.Equal(t, Int(1), v0)
	assert.Equal(t, Int(0), vA)
	assert.Equal(t, Int(0), vB) // 0 - because keyB is not imported
}

func TestState_MarshalJSON(t *testing.T) {
	st := NewState(0, nil).init(addrA, 123)
	st.Increment(coin, addr0, Int(1), 111)
	st.Get(coin, addrC)
	st.Increment(coin, addrB, Int(100), 222)
	st.Get(coin, addrA)

	data, err := json.Marshal(st)

	assert.NoError(t, err)
	assert.JSONEq(t, `[
	  {
		"chain": 0,
		"address": "Like3m1UbktLcKpr2uLihHakhREPX23xUgdChrZnWcK",
		"asset":   "0001",
		"total":   1,
		"tag":   111
	  },
	  {
		"chain": 0,
		"address": "Like5T98kZKvq49awa7awjHWvD25wkJKMQD7g6Q5X9r",
		"asset":   "0001",
		"total":   100,
		"tag":   222
	  }
	]`, string(data))
}

func TestState_Values(t *testing.T) {
	st := NewState(0, nil).init(addrA, 10).init(addrB, 5).init(addrC, 1)

	err := exec(func() {
		st.Increment(coin, addr0, Int(1), 111)
		st.Get(coin, addrA)
		st.Decrement(coin, addrB, Int(2), 222)
		st.Decrement(coin, addrB, Int(3), 333)
		st.Get(coin, addrC)
		st.Increment(coin, addr0, Int(3), 333)
	})
	values := st.Values()

	assert.NoError(t, err)
	assert.JSONEq(t, `[
	  {
		"chain": 0,
		"asset": "0001",
		"address": "Like3m1UbktLcKpr2uLihHakhREPX23xUgdChrZnWcK",
		"total": 1,
		"tag": 111
	  },
	  {
		"chain": 0,
		"asset": "0001",
		"address": "Like5T98kZKvq49awa7awjHWvD25wkJKMQD7g6Q5X9r",
		"total": 3,
		"tag": 222
	  },
	  {
		"chain": 0,
		"asset": "0001",
		"address": "Like5T98kZKvq49awa7awjHWvD25wkJKMQD7g6Q5X9r",
		"total": 0,
		"tag": 333
	  },
	  {
		"chain": 0,
		"asset": "0001",
		"address": "Like3m1UbktLcKpr2uLihHakhREPX23xUgdChrZnWcK",
		"total": 4,
		"tag": 333
	  }
	]`, enc.JSON(values))
}
