package enrichment_test

import (
	"math"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/AndreyMashukov/go-crypto-bot/server/market-watcher/enrichment"
)

func makeCandle(minute int, closePrice, volume float64) enrichment.KLine {
	openTime := time.Date(2026, 5, 29, 12, minute, 0, 0, time.UTC)
	return enrichment.KLine{
		Symbol:    "BTCUSDT",
		OpenTime:  openTime,
		Open:      closePrice - 0.5,
		High:      closePrice + 0.5,
		Low:       closePrice - 1,
		Close:     closePrice,
		Volume:    volume,
		Timestamp: openTime.Add(time.Minute),
	}
}

func TestStoreSnapshotEmpty(t *testing.T) {
	s := enrichment.NewStore()
	candles, indicators := s.Snapshot("BTCUSDT")
	if len(candles.Series) != 0 {
		t.Errorf("expected empty Series, got %d", len(candles.Series))
	}
	if indicators.RSI14 != 0 || !indicators.SMA20.IsZero() {
		t.Errorf("expected zero indicators on empty store, got %+v", indicators)
	}
}

func TestStoreOnKLineAppends(t *testing.T) {
	s := enrichment.NewStore()
	for i := 0; i < 5; i++ {
		s.OnKLine(makeCandle(i, 100+float64(i), 1))
	}
	if got := s.Len("BTCUSDT"); got != 5 {
		t.Fatalf("Len = %d, want 5", got)
	}
}

func TestStoreEvictsAtMaxWindow(t *testing.T) {
	s := enrichment.NewStore()
	for i := 0; i < enrichment.MaxWindow+10; i++ {
		s.OnKLine(makeCandle(i, 100+float64(i), 1))
	}
	if got := s.Len("BTCUSDT"); got != enrichment.MaxWindow {
		t.Fatalf("Len = %d, want %d", got, enrichment.MaxWindow)
	}
	candles, _ := s.Snapshot("BTCUSDT")
	first := candles.Series[0]
	// After overrun by 10, the first kept candle is the one for minute 10.
	wantClose := decimal.NewFromFloat(110)
	if !first.Close.Equal(wantClose) {
		t.Fatalf("first kept close = %s, want %s", first.Close, wantClose)
	}
}

func TestStoreSameOpenTimeReplacesNotAppends(t *testing.T) {
	s := enrichment.NewStore()
	s.OnKLine(makeCandle(0, 100, 1))
	s.OnKLine(makeCandle(0, 105, 2)) // same minute, updated close+volume
	if got := s.Len("BTCUSDT"); got != 1 {
		t.Fatalf("Len = %d, want 1 (same-OpenTime should replace)", got)
	}
	candles, _ := s.Snapshot("BTCUSDT")
	if !candles.Series[0].Close.Equal(decimal.NewFromFloat(105)) {
		t.Errorf("close = %s, want 105 (replacement)", candles.Series[0].Close)
	}
}

func TestSMA20Math(t *testing.T) {
	s := enrichment.NewStore()
	// 20 candles with close = 100..119; SMA20 = average = 109.5.
	for i := 0; i < 20; i++ {
		s.OnKLine(makeCandle(i, 100+float64(i), 1))
	}
	_, ind := s.Snapshot("BTCUSDT")
	got, _ := ind.SMA20.Float64()
	if math.Abs(got-109.5) > 0.0001 {
		t.Errorf("SMA20 = %f, want 109.5", got)
	}
}

func TestSMA20NotPrimed(t *testing.T) {
	s := enrichment.NewStore()
	for i := 0; i < 10; i++ {
		s.OnKLine(makeCandle(i, 100, 1))
	}
	_, ind := s.Snapshot("BTCUSDT")
	if !ind.SMA20.IsZero() {
		t.Errorf("SMA20 = %s, want zero (only 10 candles, period 20)", ind.SMA20)
	}
}

func TestVWMA20Math(t *testing.T) {
	s := enrichment.NewStore()
	// 20 candles, close = 100, volume = 1..20 → VWMA = sum(close*vol)/sum(vol)
	// = 100 * sum(1..20) / sum(1..20) = 100.
	for i := 0; i < 20; i++ {
		s.OnKLine(makeCandle(i, 100, float64(i+1)))
	}
	_, ind := s.Snapshot("BTCUSDT")
	got, _ := ind.VWMA20.Float64()
	if math.Abs(got-100) > 0.0001 {
		t.Errorf("VWMA20 = %f, want 100", got)
	}
}

func TestRSIAllGains(t *testing.T) {
	s := enrichment.NewStore()
	// Monotone-increasing closes → no losses → RSI = 100.
	for i := 0; i < 20; i++ {
		s.OnKLine(makeCandle(i, 100+float64(i), 1))
	}
	_, ind := s.Snapshot("BTCUSDT")
	if ind.RSI14 != 100 {
		t.Errorf("RSI14 = %f, want 100 for monotone gains", ind.RSI14)
	}
}

func TestRSINotPrimed(t *testing.T) {
	s := enrichment.NewStore()
	// fewer than RSIPeriod+1 candles → RSI = 0 (kill-switch signal).
	for i := 0; i < enrichment.RSIPeriod; i++ {
		s.OnKLine(makeCandle(i, 100+float64(i), 1))
	}
	_, ind := s.Snapshot("BTCUSDT")
	if ind.RSI14 != 0 {
		t.Errorf("RSI14 = %f, want 0 (only %d candles, need %d+1)", ind.RSI14, enrichment.RSIPeriod, enrichment.RSIPeriod)
	}
}

func TestSnapshotIsCopy(t *testing.T) {
	s := enrichment.NewStore()
	s.OnKLine(makeCandle(0, 100, 1))
	candles, _ := s.Snapshot("BTCUSDT")
	// Mutate the returned slice; subsequent Snapshot must be unaffected.
	candles.Series[0].Close = decimal.NewFromFloat(999)
	again, _ := s.Snapshot("BTCUSDT")
	if again.Series[0].Close.Equal(decimal.NewFromFloat(999)) {
		t.Errorf("Snapshot leaks the internal slice; got %s", again.Series[0].Close)
	}
}
