package state

import (
	"errors"

	"github.com/denisskin/bin"
)

type State struct {
	keys    [][]byte
	vals    map[string]int64 // key: value
	storage StateStorage
}

type StateStorage interface {
	GetState(key []byte) (int64, error)
}

var (
	errIncrement = errors.New("state increment error")
	errDecrement = errors.New("state decrement error")
)

func NewState(storage StateStorage) *State {
	return &State{
		vals:    map[string]int64{},
		storage: storage,
	}
}

func (s *State) Get(key []byte) (v int64, err error) {
	v, ok := s.vals[strKey(key)]
	if !ok && s.storage != nil {
		if v, err = s.storage.GetState(key); err != nil {
			s.set(key, v)
		}
	}
	return
}

func strKey(key []byte) string {
	return string(key)
}

func (s *State) set(key []byte, v int64) {
	sKey := strKey(key)
	if _, ok := s.vals[sKey]; !ok {
		s.keys = append(s.keys, key)
	}
	s.vals[sKey] = v
}

func (s *State) Increment(key []byte, v int64) (err error) {
	if v == 0 {
		return errIncrement
	}
	if v < 0 {
		return s.Decrement(key, -v)
	}
	cur, err := s.Get(key)
	if err != nil {
		return
	}
	s.set(key, cur+v)
	return
}

func (s *State) Decrement(key []byte, v int64) (err error) {
	if v <= 0 {
		return errDecrement
	}
	cur, err := s.Get(key)
	if err != nil {
		return
	}
	if cur < v {
		return errDecrement
	}
	s.set(key, cur-v)
	return
}

func (s *State) Encode() []byte {
	n := 0 // count no zero values
	for _, v := range s.vals {
		if v != 0 {
			n++
		}
	}
	w := bin.NewBuffer(nil)
	w.WriteVarInt(n)
	for _, key := range s.keys {
		if v := s.vals[strKey(key)]; v > 0 {
			w.WriteBytes(key)
			w.WriteVarInt64(v)
		}
	}
	return w.Bytes()
}

func (s *State) Decode(data []byte) error {
	r := bin.NewBuffer(data)
	for n, _ := r.ReadVarInt(); n > 0 && r.Error() == nil; n-- {
		key, _ := r.ReadBytes()
		v, _ := r.ReadVarInt64()
		s.set(key, v)
	}
	return r.Error()
}
