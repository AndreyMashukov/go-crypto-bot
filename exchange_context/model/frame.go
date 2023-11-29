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

func (f *Frame) GetBestFrameBuy(limit TradeLimit, marketDepth Depth) ([2]float64, error) {
	openPrice := 0.00
	closePrice := 0.00
	potentialOpenPrice := 0.00

	for _, bid := range marketDepth.GetBids() {
		potentialOpenPrice = bid[0].Value
		closePrice = potentialOpenPrice * (100 + limit.MinProfitPercent) / 100

		if potentialOpenPrice <= f.Low {
			break
		}

		if closePrice <= f.AvgHigh {
			openPrice = potentialOpenPrice
			break
		}
	}

	if openPrice == 0.00 {
		return [2]float64{0.00, 0.00}, errors.New(fmt.Sprintf(
			"Order Depth is out of Frame [low:%f - high:%f] [must close = %f, if open = %f]",
			f.AvgLow,
			f.AvgHigh,
			closePrice,
			potentialOpenPrice,
		))
	}

	return [2]float64{f.AvgLow, openPrice}, nil
}

func (f *Frame) GetBestFrameSell(marketDepth Depth) ([2]float64, error) {
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
			"Order Depth is out of Frame [low:%f - high:%f]",
			f.AvgLow,
			f.AvgHigh,
		))
	}

	return [2]float64{closePrice, f.AvgHigh}, nil
}

func (f *Frame) GetMediumVolatilityPercent() float64 {
	return (f.AvgHigh * 100 / f.AvgLow) - 100
}

func (f *Frame) GetVolatilityPercent() float64 {
	return (f.High * 100 / f.Low) - 100
}
