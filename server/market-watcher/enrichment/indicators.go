package enrichment

import (
	"github.com/shopspring/decimal"

	"github.com/AndreyMashukov/go-crypto-bot/server/shared/event"
)

// Window sizes the strategy reads off the tick. Held as constants so
// tests can assert against the same numbers the watcher uses.
const (
	RSIPeriod   = 14
	SMA20Period = 20
	SMA50Period = 50
	VWMA20Size  = 20
)

// ComputeIndicators returns the indicator block over the provided
// candle series. Missing periods (e.g. RSI when fewer than RSIPeriod+1
// candles are buffered) yield zero values; the strategy reads zero
// values as "not yet primed" and falls back to its kill-switch path.
func ComputeIndicators(candles []event.Candle) event.Indicators {
	return event.Indicators{
		RSI14:  rsi(candles, RSIPeriod),
		SMA20:  sma(candles, SMA20Period),
		SMA50:  sma(candles, SMA50Period),
		VWMA20: vwma(candles, VWMA20Size),
	}
}

// sma is the close-price simple moving average over the most recent
// `period` candles. Returns zero when fewer candles are available.
func sma(candles []event.Candle, period int) event.Decimal {
	if len(candles) < period || period <= 0 {
		return decimal.Zero
	}
	sum := decimal.Zero
	for i := len(candles) - period; i < len(candles); i++ {
		sum = sum.Add(candles[i].Close)
	}
	return sum.Div(decimal.NewFromInt(int64(period)))
}

// vwma is the volume-weighted moving average of close prices over the
// most recent `period` candles. Returns zero when fewer candles are
// available or when the total volume is zero (would divide by zero).
func vwma(candles []event.Candle, period int) event.Decimal {
	if len(candles) < period || period <= 0 {
		return decimal.Zero
	}
	weighted := decimal.Zero
	totalVolume := decimal.Zero
	for i := len(candles) - period; i < len(candles); i++ {
		c := candles[i]
		weighted = weighted.Add(c.Close.Mul(c.Volume))
		totalVolume = totalVolume.Add(c.Volume)
	}
	if totalVolume.IsZero() {
		return decimal.Zero
	}
	return weighted.Div(totalVolume)
}

// rsi is the classic Wilder RSI over `period` close-to-close moves on
// the most recent slice of candles. Returns zero when fewer than
// period+1 candles are buffered.
func rsi(candles []event.Candle, period int) float64 {
	if len(candles) <= period || period <= 0 {
		return 0
	}
	tail := candles[len(candles)-period-1:]
	gainSum := 0.0
	lossSum := 0.0
	for i := 1; i < len(tail); i++ {
		diff, _ := tail[i].Close.Sub(tail[i-1].Close).Float64()
		switch {
		case diff > 0:
			gainSum += diff
		case diff < 0:
			lossSum += -diff
		}
	}
	avgGain := gainSum / float64(period)
	avgLoss := lossSum / float64(period)
	if avgLoss == 0 {
		if avgGain == 0 {
			return 0
		}
		return 100
	}
	rs := avgGain / avgLoss
	return 100 - 100/(1+rs)
}
