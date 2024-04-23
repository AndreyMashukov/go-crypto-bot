package model

import (
	"math"
	"time"
)

type PriceChangeSpeed struct {
	CloseTime       int64   `json:"closeTime"`
	FromPrice       float64 `json:"fromPrice"`
	ToPrice         float64 `json:"toPrice"`
	FromTime        int64   `json:"fromTime"`
	ToTime          int64   `json:"ToTime"`
	PointsPerSecond float64 `json:"pointsPerSecond"`
}

type KLine struct {
	Symbol              string             `json:"s"`
	Open                float64            `json:"o,string"`
	Close               float64            `json:"c,string"`
	Low                 float64            `json:"l,string"`
	High                float64            `json:"h,string"`
	Interval            string             `json:"i"`
	Timestamp           int64              `json:"T,int"`
	OpenTime            int64              `json:"t,int"`
	Volume              float64            `json:"v,string"`
	UpdatedAt           int64              `json:"updatedAt"`
	PriceChangeSpeed    []PriceChangeSpeed `json:"priceChangeSpeed"`
	PriceChangeSpeedMax *float64           `json:"priceChangeSpeedMax"`
	PriceChangeSpeedMin *float64           `json:"priceChangeSpeedMin"`
}

func (k *KLine) GetPriceChangeSpeedMax() float64 {
	if k.PriceChangeSpeedMax != nil {
		return *k.PriceChangeSpeedMax
	}

	return 0.00
}

func (k *KLine) GetPriceChangeSpeedMin() float64 {
	if k.PriceChangeSpeedMin != nil {
		return *k.PriceChangeSpeedMin
	}

	return 0.00
}

func (k *KLine) GetPriceChangeSpeed() []PriceChangeSpeed {
	priceChangeSpeed := make([]PriceChangeSpeed, 0)

	if k.PriceChangeSpeed != nil {
		return k.PriceChangeSpeed
	}

	return priceChangeSpeed
}

func (k *KLine) GetPriceChangeSpeedAvg() float64 {
	avgValue := 0.00
	changes := k.GetPriceChangeSpeed()

	if len(changes) > 0 {
		valueSum := 0.00
		for _, change := range changes {
			valueSum += change.PointsPerSecond
		}

		return valueSum / float64(len(changes))
	}

	return avgValue
}

func (k *KLine) IsNegative() bool {
	return k.Close < k.Open
}

func (k *KLine) IsPositive() bool {
	return k.Close > k.Open
}

func (k *KLine) GetLowPercent(percent float64) float64 {
	return k.Low + (k.Low * percent / 100)
}

const PriceNotActualSeconds = 5
const PriceValidSeconds = 30

func (k *KLine) IsPriceExpired() bool {
	return (time.Now().Unix() - (k.UpdatedAt)) > PriceValidSeconds
}

func (k *KLine) IsPriceNotActual() bool {
	return (time.Now().Unix() - (k.UpdatedAt)) > PriceNotActualSeconds
}

func (k *KLine) Includes(ticker MiniTicker) bool {
	return k.OpenTime <= ticker.EventTime && ticker.EventTime < k.Timestamp
}

func (k *KLine) Update(ticker MiniTicker) {
	k.UpdatedAt = time.Now().Unix()
	k.Close = ticker.Close
	k.High = math.Max(k.High, ticker.Close)
	k.Low = math.Min(k.Low, ticker.Close)
}

type KlineBatch struct {
	Items []KLine `json:"items"`
}
