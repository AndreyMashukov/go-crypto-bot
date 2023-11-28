package model

import (
	"errors"
	"fmt"
)

type Frame struct {
	AvgHigh float64 `json:"avgHigh"`
	AvgLow  float64 `json:"avgLow"`
}

func (f *Frame) GetBestFramePrice(limit TradeLimit, marketDepth Depth) ([2]float64, error) {
	openPrice := 0.00
	closePrice := 0.00

	for _, bid := range marketDepth.GetBids() {
		potentialOpenPrice := bid[0].Value
		closePrice = potentialOpenPrice * (100 + limit.MinProfitPercent) / 100

		if potentialOpenPrice >= f.AvgHigh {
			continue
		}

		//if potentialOpenPrice <= f.AvgLow {
		//	break
		//}

		if closePrice < f.AvgHigh {
			openPrice = potentialOpenPrice
			break
		}
	}

	if openPrice == 0.00 {
		return [2]float64{0.00, 0.00}, errors.New(fmt.Sprintf(
			"Bad time to buy! Frame %f - %f [must close = %f]",
			f.AvgLow,
			f.AvgHigh,
			closePrice,
		))
	}

	return [2]float64{f.AvgLow, openPrice}, nil
}
