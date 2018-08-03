package db

import (
	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/assets"
	"github.com/likecoin-pro/likecoin/commons/bignum"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/likecoin-pro/likecoin/object"
)

type Statistic struct {
	Blocks uint64          `json:"blocks"` //
	Txs    int64           `json:"txs"`    //
	Users  int64           `json:"users"`  //
	Coins  []CoinStatistic `json:"coins"`  //
}

type CoinStatistic struct {
	Asset  assets.Asset `json:"asset"`  //
	Likes  int64        `json:"likes"`  //
	Rate   bignum.Int   `json:"rate"`   //
	Supply bignum.Int   `json:"supply"` //
}

func (s *Statistic) New(blockNum uint64, blockTxs int) *Statistic {
	c := s.Clone()
	c.Blocks = blockNum
	c.Txs += int64(blockTxs)
	return c
}

func (s *Statistic) Clone() *Statistic {
	var c = *s
	c.Coins = make([]CoinStatistic, len(s.Coins))
	copy(c.Coins, s.Coins)
	return &c
}

func (s *Statistic) String() string {
	return enc.JSON(s)
}

func (s *Statistic) Encode() []byte {
	return bin.Encode(
		0,
		s.Blocks,
		s.Txs,
		s.Users,
		s.Coins,
	)
}

func (s *Statistic) Decode(data []byte) error {
	return bin.Decode(data,
		new(int), // version
		&s.Blocks,
		&s.Txs,
		&s.Users,
		&s.Coins,
	)
}

func (c CoinStatistic) Encode() []byte {
	return bin.Encode(
		c.Asset,
		c.Likes,
		c.Rate,
		c.Supply,
	)
}

func (c *CoinStatistic) Decode(data []byte) error {
	return bin.Decode(data,
		&c.Asset,
		&c.Likes,
		&c.Rate,
		&c.Supply,
	)
}

func (s *Statistic) CoinStat(a assets.Asset) CoinStatistic {
	for _, c := range s.Coins {
		if c.Asset.Equal(a) {
			return c
		}
	}
	return CoinStatistic{Asset: a}
}

func (s *Statistic) setCoinStat(v CoinStatistic) {
	for i, c := range s.Coins {
		if c.Asset.Equal(v.Asset) {
			s.Coins[i] = v
			return
		}
	}
	s.Coins = append(s.Coins, v)
}

func (s *Statistic) Refresh(emission *object.Emission) {
	c := s.CoinStat(emission.Asset)
	c.Rate = emission.Rate
	c.Likes += emission.TotalDelta()
	c.Supply = c.Supply.Add(emission.TotalAmount())
	s.setCoinStat(c)
}
