package state

import (
	"encoding/json"
	"testing"

	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/stretchr/testify/assert"
)

var (
	testCoin = crypto.Asset{127}

	key0 = NewKey(crypto.MustParseAddress("Like3m1UbktLcKpr2uLihHakhREPX23xUgdChrZnWcK"), testCoin)
	keyA = NewKey(crypto.MustParseAddress("Like5eBiwK1JXRTsfNAAPN5GD6zwUWjdvu5y8JXRiLJ"), testCoin)
	keyB = NewKey(crypto.MustParseAddress("Like5T98kZKvq49awa7awjHWvD25wkJKMQD7g6Q5X9r"), testCoin)
	keyC = NewKey(crypto.MustParseAddress("Like5LtpZeGE5Ve5NbjCeraFVY5GCcTSFv9FYDzb7nm"), testCoin)
)

func (s *State) init(k Key, v int64) *State {
	s.vals[k.str()] = v
	return s
}

func TestState_Get(t *testing.T) {

	st := NewState().init(keyA, 10)

	v0 := st.Get(key0)
	v1 := st.Get(keyA)

	assert.Equal(t, int64(0), v0)
	assert.Equal(t, int64(10), v1)
}

func TestState_Keys(t *testing.T) {
	st := NewState().init(keyA, 10).init(keyB, 5).init(keyC, 1)

	err := st.Execute(func() {
		st.Inc(key0, 1)
		st.Get(keyA)
		st.Inc(keyB, -5)
		st.Get(keyC)
	})
	changedKeys := st.Keys()

	assert.NoError(t, err)
	assert.Equal(t, []Key{key0, keyB}, changedKeys)
}

func TestState_Equal(t *testing.T) {
	a := NewState().init(keyA, 666)
	a.Get(keyA)
	a.Inc(key0, 123)

	b := NewState().init(key0, 100).init(keyA, 333)
	b.Get(keyB)
	b.Inc(key0, 23)
	b.Get(keyC)

	c := NewState().init(key0, 123)
	c.Get(key0)

	assert.True(t, a.Equal(b))
	assert.True(t, b.Equal(a))
	assert.False(t, c.Equal(a))
}

func TestState_Inc(t *testing.T) {
	st := NewState().init(keyA, 10)

	err := st.Execute(func() {
		st.Inc(key0, 1)
		st.Inc(keyA, 1)
		st.Inc(keyA, -2)
	})

	v0 := st.Get(key0)
	vA := st.Get(keyA)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), v0)
	assert.Equal(t, int64(9), vA)
}

func TestState_Inc_fail(t *testing.T) {
	st := NewState().init(keyA, 10)

	err0 := st.Execute(func() { st.Inc(key0, -1) })
	err1 := st.Execute(func() { st.Inc(keyA, -1) })
	err2 := st.Execute(func() { st.Inc(keyA, -10) })
	v0 := st.Get(key0)
	vA := st.Get(keyA)

	assert.Error(t, err0)
	assert.NoError(t, err1)
	assert.Error(t, err2)
	assert.Equal(t, int64(0), v0)
	assert.Equal(t, int64(9), vA)
}

func TestState_Encode(t *testing.T) {
	s1 := NewState().init(key0, 12)
	s1.Inc(keyA, 34)
	s1.Inc(keyB, 56)
	data1 := s1.Encode()

	var s2 = new(State)
	err2 := s2.Decode(data1)
	data2 := s2.Encode()

	assert.NoError(t, err2)
	assert.Equal(t, data1, data2)
}

func TestState_Decode(t *testing.T) {
	s := NewState().init(keyA, 10).init(keyB, 10)
	s.Inc(key0, 1)
	s.Inc(keyA, -10)
	data := s.Encode() // encode only changed values

	st := NewState()
	err := st.Decode(data)
	v0 := st.Get(key0)
	vA := st.Get(keyA)
	vB := st.Get(keyB)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), v0)
	assert.Equal(t, int64(0), vA)
	assert.Equal(t, int64(0), vB) // 0 - because keyB is not imported
}

func TestState_MarshalJSON(t *testing.T) {
	st := NewState().init(keyA, 123)
	st.Inc(key0, 1)
	st.Get(keyC)
	st.Inc(keyB, 100)
	st.Get(keyA)

	data, err := json.Marshal(st)

	assert.NoError(t, err)
	assert.JSONEq(t, `[
	  {
		"address": "Like3m1UbktLcKpr2uLihHakhREPX23xUgdChrZnWcK",
		"asset":   "ff",
		"value":   1
	  },
	  {
		"address": "Like5T98kZKvq49awa7awjHWvD25wkJKMQD7g6Q5X9r",
		"asset":   "ff",
		"value":   100
	  }
	]`, string(data))
}
