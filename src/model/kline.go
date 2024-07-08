package model

import (
	"math"
	"time"
)

const KLineSourceTickerStream = "ticker_stream"
const KLineSourceKLineStream = "kline_stream"
const KLineSourceKLineFetch = "kline_fetch"

type PriceChange struct {
	CloseTime       TimestampMilli `json:"closeTime"`
	FromPrice       float64        `json:"fromPrice"`
	ToPrice         float64        `json:"toPrice"`
	FromTime        TimestampMilli `json:"fromTime"`
	ToTime          TimestampMilli `json:"ToTime"`
	PointsPerSecond float64        `json:"pointsPerSecond"`
}

type PriceChangeSpeed struct {
	Symbol    string         `json:"symbol"`
	Timestamp TimestampMilli `json:"timestamp"`
	Changes   []PriceChange  `json:"changes"`
	MaxChange float64        `json:"maxChange"`
	MinChange float64        `json:"minChange"`
}

type KLine struct {
	Symbol           string            `json:"s"`
	Open             Price             `json:"o"`
	Close            Price             `json:"c"`
	Low              Price             `json:"l"`
	High             Price             `json:"h"`
	Interval         string            `json:"i"`
	Timestamp        TimestampMilli    `json:"T,int"`
	OpenTime         TimestampMilli    `json:"t,int"`
	Volume           Volume            `json:"v"`
	UpdatedAt        int64             `json:"updatedAt"`
	PriceChangeSpeed *PriceChangeSpeed `json:"priceChangeSpeed"`
	TradeVolume      *TradeVolume      `json:"tradeVolume"`
	Source           string            `json:"source"`
}

func (k *KLine) GetTradeVolumeSell() float64 {
	if k.TradeVolume != nil {
		return k.TradeVolume.SellQty
	}

	return 0.00
}

func (k *KLine) GetTradeVolumeBuy() float64 {
	if k.TradeVolume != nil {
		return k.TradeVolume.BuyQty
	}

	return 0.00
}

func (k *KLine) GetPriceChangeSpeedMax() float64 {
	if k.PriceChangeSpeed != nil {
		return k.PriceChangeSpeed.MaxChange
	}

	return 0.00
}

func (k *KLine) GetPriceChangeSpeedMin() float64 {
	if k.PriceChangeSpeed != nil {
		return k.PriceChangeSpeed.MinChange
	}

	return 0.00
}

func (k *KLine) GetPriceChangeSpeed() []PriceChange {
	priceChangeSpeed := make([]PriceChange, 0)

	if k.PriceChangeSpeed != nil {
		return k.PriceChangeSpeed.Changes
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
	return k.Close.Value() < k.Open.Value()
}

func (k *KLine) IsPositive() bool {
	return k.Close > k.Open
}

func (k *KLine) GetLowPercent(percent float64) float64 {
	return k.Low.Value() + (k.Low.Value() * percent / 100)
}

const PriceNotActualSeconds = 5
const PriceValidSeconds = 30

func (k *KLine) IsPriceWrongTimestamp() bool {
	return TimestampMilli(time.Now().UnixMilli()).GetPeriodToMinute() > k.Timestamp.GetPeriodToMinute()
}

func (k *KLine) IsPriceExpired() bool {
	return (time.Now().Unix() - (k.UpdatedAt)) > PriceValidSeconds
}

func (k *KLine) IsPriceNotActual() bool {
	return (time.Now().Unix() - (k.UpdatedAt)) > PriceNotActualSeconds
}

func (k *KLine) Includes(ticker MiniTicker) bool {
	return k.OpenTime <= ticker.EventTime && ticker.EventTime < k.Timestamp
}

func (k *KLine) Update(ticker MiniTicker, source string) KLine {
	// WARNING!!!
	// This is daily ticker price, we can use only `ticker.Close` for minute KLines!
	currentInterval := TimestampMilli(time.Now().UnixMilli()).GetPeriodToMinute()
	if k.Timestamp.GetPeriodToMinute() < currentInterval {
		return KLine{
			Timestamp: TimestampMilli(currentInterval),
			Symbol:    ticker.Symbol,
			Open:      ticker.Close,
			Close:     ticker.Close,
			High:      ticker.Close,
			Low:       ticker.Close,
			Interval:  "1m",
			OpenTime:  TimestampMilli(TimestampMilli(currentInterval).GetPeriodFromMinute()),
			UpdatedAt: time.Now().Unix(),
			Source:    source,
		}
	}

	return KLine{
		Timestamp:        TimestampMilli(currentInterval),
		Symbol:           ticker.Symbol,
		Open:             k.Open,
		Close:            ticker.Close,
		High:             Price(math.Max(k.High.Value(), ticker.Close.Value())),
		Low:              Price(math.Min(k.Low.Value(), ticker.Close.Value())),
		Interval:         "1m",
		OpenTime:         TimestampMilli(TimestampMilli(currentInterval).GetPeriodFromMinute()),
		UpdatedAt:        time.Now().Unix(),
		PriceChangeSpeed: k.PriceChangeSpeed,
		TradeVolume:      k.TradeVolume,
		Source:           source,
	}
}

type KlineBatch struct {
	Items []KLine `json:"items"`
}
