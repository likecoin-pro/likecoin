package state

import (
	"errors"
	"math/big"

	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/crypto"
)

type State struct {
	chainID uint64
	getter  func(assets.Asset, crypto.Address) Number //

	vals map[string]Number //
	sets Values            //
}

var (
	ErrNegativeValue = errors.New("blockchain/state: not enough funds")
	ErrInvalidKey    = errors.New("blockchain/state: invalid key")
)

func NewState(chainID uint64, getter func(assets.Asset, crypto.Address) Number) *State {
	return &State{
		chainID: chainID,
		getter:  getter,
		vals:    map[string]Number{},
	}
}

func (s *State) NewSubState() *State {
	return NewState(s.chainID, s.Get)
}

func (s *State) Copy() *State {
	a := NewState(s.chainID, nil)
	for _, v := range s.sets {
		a.set(v)
	}
	return a
}

func strKey(a assets.Asset, addr crypto.Address) string {
	return string(a) + string(addr[:])
}

func (s *State) Get(asset assets.Asset, addr crypto.Address) Number {
	sKey := strKey(asset, addr)
	val, ok := s.vals[sKey]
	if !ok {
		if s.getter != nil {
			val = s.getter(asset, addr)
		}
		if val == nil {
			val = Int(0)
		}
		s.vals[sKey] = val
	}
	return new(big.Int).Set(val)
}

func (s *State) Values() Values {
	return s.sets
}

func (s *State) set(v *Value) {
	if v.Balance.Sign() < 0 {
		s.Fail(ErrNegativeValue)
		return
	}
	if v.ChainID == s.chainID {
		s.vals[strKey(v.Asset, v.Address)] = v.Balance
	}
	s.sets = append(s.sets, v)
}

func (s *State) Apply(vv Values) {
	for _, v := range vv {
		s.set(v)
	}
}

func (s *State) Set(asset assets.Asset, addr crypto.Address, v Number, tag uint64) {
	s.set(&Value{s.chainID, asset, addr, v, tag})
}

func (s *State) CrossChainSet(chainID uint64, asset assets.Asset, addr crypto.Address, v Number, tag uint64) {
	s.set(&Value{chainID, asset, addr, v, tag})
}

func (s *State) Increment(asset assets.Asset, addr crypto.Address, delta Number, tag uint64) {
	if delta.Sign() == 0 {
		return
	}
	v := s.Get(asset, addr)
	v = v.Add(v, delta)
	s.Set(asset, addr, v, tag)
}

func (s *State) Decrement(asset assets.Asset, addr crypto.Address, delta Number, tag uint64) {
	v := new(big.Int).Neg(delta)
	s.Increment(asset, addr, v, tag)
}

func (s *State) Fail(err error) {
	panic(err)
}
