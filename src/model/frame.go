package model

import (
	"errors"
	"fmt"
)

type Frame struct {
	High    float64 `json:"high"`
	Low     float64 `json:"low"`
	AvgHigh float64 `json:"avgHigh"`
	AvgLow  float64 `json:"avgLow"`
}

func (f *Frame) GetBestFrameSell(marketDepth OrderBookModel) ([2]float64, error) {
	closePrice := 0.00

	for _, ask := range marketDepth.GetAsksReversed() {
		if ask[0].Value >= f.AvgHigh {
			continue
		}

		closePrice = ask[0].Value
		break
	}

	if closePrice == 0.00 {
		return [2]float64{0.00, 0.00}, errors.New(fmt.Sprintf(
			"Order OrderBookModel is out of Frame [low:%f - high:%f]",
			f.AvgLow,
			f.AvgHigh,
		))
	}

	return [2]float64{closePrice, f.AvgHigh}, nil
}

func (f *Frame) GetMediumVolatilityPercent() float64 {
	return (f.AvgHigh * 100.00 / f.AvgLow) - 100.00
}

func (f *Frame) GetVolatilityPercent() float64 {
	return (f.High * 100.00 / f.Low) - 100.00
}
