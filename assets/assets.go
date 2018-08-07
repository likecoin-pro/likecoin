package assets

// Asset types
const (
	CoinType = 0
	NameType = 1
)

var (
	Likecoin = Asset{CoinType, 1} // is synonym of "YotubeCoin"

	Default = Likecoin
)

func NewName(name string) Asset {
	return append(Asset{NameType}, []byte(name)...)
}
