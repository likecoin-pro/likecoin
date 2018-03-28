package assets

// coin units
const (
	Coin      = 1e9
	MilliCoin = Coin / 1e3
	MicroCoin = Coin / 1e6
	NanoCoin  = Coin / 1e9
)

type coinConfig struct {
	Label     string
	SrcURL    string
	EmissionA float64
	EmissionK float64
}

var coinsCFG = map[uint8]*coinConfig{}

func newCoin(id uint8, label, srcURL string, emissionA int64, emissionK float64) Asset {
	coinsCFG[id] = &coinConfig{
		Label:     label,
		SrcURL:    srcURL,
		EmissionA: float64(emissionA),
		EmissionK: emissionK,
	}
	return Asset{CoinType, id}
}
