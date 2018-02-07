package state

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/denisskin/bin"
)

type State struct {
	vals   map[string]int64    //
	sets   map[string]struct{} //
	keys   []Key               //
	getter func(key Key) int64 //
	setter func(Key, int64)    //
	mx     sync.Mutex          // ??
}

var (
	//errIncrement = errors.New("blockchain/state-error: increment error")
	ErrDecrement  = errors.New("blockchain/state-error: decrement error")
	ErrInvalidKey = errors.New("blockchain/state-error: invalid key")
)

func NewState() *State {
	return NewStateEx(nil, nil)
}

func NewStateEx(
	getter func(Key) int64,
	setter func(Key, int64),
) *State {
	return &State{
		getter: getter,
		setter: setter,
		vals:   map[string]int64{},
		sets:   map[string]struct{}{},
	}
}

func (s *State) NewSubState() *State {
	return NewStateEx(s.Get, s.Set)
}

func (s *State) Copy() *State {
	a := NewState()
	for _, key := range s.Keys() {
		a.Set(key, s.Get(key))
	}
	return a
}

func (s *State) Keys() []Key {
	return s.keys
}

func (s *State) Get(key Key) int64 {
	sKey := key.str()
	if val, ok := s.vals[sKey]; ok {
		return val
	}
	if s.getter != nil {
		val := s.getter(key)
		s.vals[sKey] = val
		return val
	} else {
		return 0
	}
}

func (s *State) Set(key Key, v int64) {
	sKey := key.str()
	s.vals[sKey] = v
	if _, ok := s.sets[sKey]; !ok {
		s.keys = append(s.keys, key)
		s.sets[sKey] = struct{}{}
	}
	if s.setter != nil {
		s.setter(key, v)
	}
}

func (s *State) Inc(key Key, v int64) {
	v = s.Get(key) + v
	if v < 0 {
		panic(ErrDecrement)
	}
	s.Set(key, v)
}

func (s *State) Equal(a *State) bool {
	if len(s.keys) != len(a.keys) {
		return false
	}
	for _, key := range s.keys {
		if s.Get(key) != a.Get(key) {
			return false
		}
	}
	return true
}

func (s *State) Encode() []byte {
	w := bin.NewBuffer(nil)
	w.WriteVarInt(len(s.keys))
	for _, key := range s.keys {
		w.WriteVar(key)
		w.WriteVarInt64(s.Get(key))
	}
	return w.Bytes()
}

func (s *State) Decode(data []byte) error {
	s.vals = map[string]int64{}
	s.sets = map[string]struct{}{}

	r := bin.NewBuffer(data)
	var key Key
	for n, _ := r.ReadVarInt(); n > 0 && r.Error() == nil; n-- {
		r.ReadVar(&key)
		v, _ := r.ReadVarInt64()
		s.Set(key, v)
	}
	return r.Error()
}

type stateValue struct {
	Addr  string `json:"address"`
	Asset string `json:"asset"`
	Value int64  `json:"value"`
}

func (s *State) MarshalJSON() ([]byte, error) {
	var vv []stateValue
	for _, key := range s.keys {
		vv = append(vv, stateValue{
			Addr:  key.Address.String(),
			Asset: key.Asset.String(),
			Value: s.Get(key),
		})
	}
	return json.Marshal(vv)
}

func (s *State) Fail(err error) {
	panic(err)
}

func (s *State) Execute(fn func()) (err error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	defer func() {
		err, _ = recover().(error)
	}()

	fn()

	return
}
