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
	YotubeCoin = NewCoin(1)
	InstaCoin  = NewCoin(2)
	YoukuCoin  = NewCoin(3)
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

var labels = map[string]string{
	LikeCoin.String():  "LIKE",
	InstaCoin.String(): "InLIKE",
	YoukuCoin.String(): "YkLIKE",
}

var mediaSrcURLs = map[uint8]string{
	1: "https://www.youtube.com/watch?v={ID}",
	2: "https://www.instagram.com/p/{ID}/",
	3: "https://v.youku.com/v_show/id_{ID}",
}
