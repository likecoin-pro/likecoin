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
	YotubeCoin = newCoin(1, "LIKE", "https://www.youtube.com/watch?v={ID}", 10*Coin, 2e-7)
	//InstaCoin  = newCoin(2, "INSTA","https://www.instagram.com/p/{ID}/", 0,0)
	//YoukuCoin  = newCoin(3,"YOUKU","https://v.youku.com/v_show/id_{ID}", 0,0)
)

func NewCounter(typ uint8, id string) Asset {
	return append(Asset{CounterType, typ}, []byte(id)...)
}

func NewName(name string) Asset {
	return append(Asset{NameType}, []byte(name)...)
}
