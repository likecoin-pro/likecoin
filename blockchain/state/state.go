package state

import (
	"errors"
	"math/big"
	"sync"

	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/crypto"
)

type State struct {
	chainID uint64
	getter  func(assets.Asset, crypto.Address) Number //

	vals map[string]Number //
	sets []*Value          //
	mx   sync.Mutex        // execution mutex
}

var (
	//errIncrement = errors.New("blockchain/state-error: increment error")
	ErrNegativeValue = errors.New("blockchain/state-error: not enough funds")
	ErrInvalidKey    = errors.New("blockchain/state-error: invalid key")
)

func NewState(chainID uint64, getter func(assets.Asset, crypto.Address) Number) *State {
	return &State{
		chainID: chainID,
		getter:  getter,
		vals:    map[string]Number{},
	}
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

func (s *State) Set(asset assets.Asset, addr crypto.Address, v Number, tag int64) {
	s.set(&Value{s.chainID, asset, addr, tag, v})
}

func (s *State) CrossChainSet(chainID uint64, asset assets.Asset, addr crypto.Address, v Number, tag int64) {
	s.set(&Value{chainID, asset, addr, tag, v})
}

func (s *State) Increment(asset assets.Asset, addr crypto.Address, delta Number, tag int64) {
	if delta.Sign() == 0 {
		return
	}
	v := s.Get(asset, addr)
	v = v.Add(v, delta)
	s.Set(asset, addr, v, tag)
}

func (s *State) Decrement(asset assets.Asset, addr crypto.Address, delta Number, tag int64) {
	v := new(big.Int).Neg(delta)
	s.Increment(asset, addr, v, tag)
}

func (s *State) Fail(err error) {
	panic(err)
}

type Transaction interface {
	Execute(*State)
}

func (s *State) Execute(tx Transaction) (newState *State, err error) {

	s.mx.Lock()
	defer s.mx.Unlock()
	defer func() {
		err, _ = recover().(error)
	}()

	newState = NewState(s.chainID, s.Get)

	tx.Execute(newState)

	// set on success
	for _, v := range newState.sets {
		s.set(v)
	}

	return
}
