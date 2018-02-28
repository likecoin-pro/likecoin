package assets

// Asset types
const (
	CoinType    = 0
	CounterType = 1
	NameType    = 2
)

var (
	LikeCoin = YotubeCoin

	// coins
	YotubeCoin    = NewCoin(1)
	InstagramCoin = NewCoin(2)
	YoukuCoin     = NewCoin(3)
)

func NewCoin(id uint8) Asset {
	return Asset{CoinType, id}
}

func NewCounter(typ uint8, id string) Asset {
	return append(Asset{CounterType, typ}, []byte(id)...)
}

func NewName(name string) Asset {
	return append(Asset{NameType}, []byte(name)...)
}
