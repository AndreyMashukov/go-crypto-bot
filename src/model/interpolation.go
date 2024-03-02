package model

type Interpolation struct {
	Asset                string  `json:"asset"`
	BtcInterpolationUsdt float64 `json:"btcInterpolationUsdt"`
	EthInterpolationUsdt float64 `json:"ethInterpolationUsdt"`
}

func (i *Interpolation) HasBoth() bool {
	return i.HasBtc() && i.HasBtc()
}

func (i *Interpolation) HasBtc() bool {
	return i.BtcInterpolationUsdt > 0.00
}

func (i *Interpolation) HasEth() bool {
	return i.EthInterpolationUsdt > 0.00
}
