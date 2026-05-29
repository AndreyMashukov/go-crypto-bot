package enrichment_test

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/AndreyMashukov/go-crypto-bot/server/market-watcher/enrichment"
	"github.com/AndreyMashukov/go-crypto-bot/server/shared/event"
)

// Direct unit tests for the indicator math, independent of Store. The
// store-level tests in store_test.go cover the typical wiring; these
// pin the edge cases (insufficient series, zero volume on VWMA, mixed
// gain/loss on RSI).

func candleAt(closePrice, volume float64) event.Candle {
	return event.Candle{
		OpenTime: time.Unix(0, 0).UTC(),
		Open:     decimal.NewFromFloat(closePrice),
		High:     decimal.NewFromFloat(closePrice),
		Low:      decimal.NewFromFloat(closePrice),
		Close:    decimal.NewFromFloat(closePrice),
		Volume:   decimal.NewFromFloat(volume),
	}
}

func TestComputeIndicatorsEmpty(t *testing.T) {
	ind := enrichment.ComputeIndicators(nil)
	if ind.RSI14 != 0 || !ind.SMA20.IsZero() || !ind.SMA50.IsZero() || !ind.VWMA20.IsZero() {
		t.Errorf("expected zero indicators, got %+v", ind)
	}
}

func TestComputeIndicatorsVWMAZeroVolume(t *testing.T) {
	series := make([]event.Candle, enrichment.VWMA20Size)
	for i := range series {
		series[i] = candleAt(100, 0)
	}
	ind := enrichment.ComputeIndicators(series)
	if !ind.VWMA20.IsZero() {
		t.Errorf("VWMA20 with zero volume = %s, want zero", ind.VWMA20)
	}
}

func TestComputeIndicatorsRSIAllLosses(t *testing.T) {
	series := make([]event.Candle, enrichment.RSIPeriod+1)
	for i := range series {
		// Strictly-decreasing closes — no gains.
		series[i] = candleAt(100-float64(i), 1)
	}
	ind := enrichment.ComputeIndicators(series)
	if ind.RSI14 != 0 {
		t.Errorf("RSI14 with all losses = %f, want 0", ind.RSI14)
	}
}

func TestComputeIndicatorsRSIMixed(t *testing.T) {
	// Alternating +1 / -1 closes over RSIPeriod moves → avgGain == avgLoss
	// → rs = 1 → rsi = 50.
	series := make([]event.Candle, enrichment.RSIPeriod+1)
	closePrice := 100.0
	for i := range series {
		series[i] = candleAt(closePrice, 1)
		if i%2 == 0 {
			closePrice++
		} else {
			closePrice--
		}
	}
	ind := enrichment.ComputeIndicators(series)
	if ind.RSI14 < 45 || ind.RSI14 > 55 {
		t.Errorf("RSI14 mixed gains/losses = %f, want ~50", ind.RSI14)
	}
}
