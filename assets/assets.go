package assets

// Asset types
const (
	CoinType    = 0
	CounterType = 1
	NameType    = 2
)

var (
	Likecoin = Asset{CoinType, 1} // is synonym of "YotubeCoin"

	Default = Likecoin
)

func NewCounter(typ uint8, id string) Asset {
	return append(Asset{CounterType, typ}, []byte(id)...)
}

func NewName(name string) Asset {
	return append(Asset{NameType}, []byte(name)...)
}
