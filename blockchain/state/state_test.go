package state

import (
	"testing"

	"errors"

	"github.com/stretchr/testify/assert"
)

func TestState_Get(t *testing.T) {
	storage := TestStorage{"a": 10}
	st := NewState(storage)

	v0, err0 := st.Get([]byte("0"))
	v1, err1 := st.Get([]byte("a"))
	_, err2 := st.Get([]byte("err"))

	assert.NoError(t, err0)
	assert.NoError(t, err1)
	assert.Error(t, err2)
	assert.Equal(t, int64(0), v0)
	assert.Equal(t, int64(10), v1)
}

func TestState_Increment(t *testing.T) {
	storage := TestStorage{"a": 10}
	st := NewState(storage)

	err0 := st.Increment([]byte("0"), 1)
	err1 := st.Increment([]byte("a"), 1)
	err2 := st.Increment([]byte("a"), -2)
	v0, _ := st.Get([]byte("0"))
	vA, _ := st.Get([]byte("a"))

	assert.NoError(t, err0)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, int64(1), v0)
	assert.Equal(t, int64(9), vA)
}

func TestState_Decrement(t *testing.T) {
	storage := TestStorage{"a": 10}
	st := NewState(storage)

	err0 := st.Decrement([]byte("0"), 1)
	err1 := st.Decrement([]byte("a"), 1)
	err2 := st.Decrement([]byte("a"), 10)
	v, _ := st.Get([]byte("a"))

	assert.Error(t, err0)
	assert.NoError(t, err1)
	assert.Error(t, err2)
	assert.Equal(t, int64(9), v)
}

func TestState_Decode(t *testing.T) {
	storage := TestStorage{"a": 10}
	s := NewState(storage)
	s.Increment([]byte("0"), 1)
	s.Increment([]byte("a"), -10)
	data := s.Encode()

	st := NewState(nil)
	err := st.Decode(data)

	v0, _ := st.Get([]byte("0"))
	vA, _ := st.Get([]byte("a"))

	assert.NoError(t, err)
	assert.Equal(t, int64(1), v0)
	assert.Equal(t, int64(0), vA)
}

//-----------------------------------------------
type TestStorage map[string]int64

func (s TestStorage) GetState(key []byte) (int64, error) {
	if string(key) == "err" {
		return 0, errors.New("test-error")
	}
	return s[strKey(key)], nil
}
