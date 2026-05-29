// Package enrichment owns the watcher-side calculation that turns
// raw klines into the Candles + Indicators block that rides on every
// outbound MarketTick. The strategy stops calling repos on the hot
// path; everything it needs to decide is on the tick.
//
// The store keeps the last N closed candles per (exchange, symbol) in
// memory. There is no persistence layer — backtests read from
// ClickHouse market_tick (the second-output path the watcher already
// runs), not from this store.
package enrichment

import (
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/AndreyMashukov/go-crypto-bot/server/shared/event"
)

// MaxWindow caps the per-symbol ring buffer. Sized for SMA50 + RSI14
// headroom. The store keeps the last MaxWindow candles; older entries
// are evicted FIFO.
const MaxWindow = 50

// Resolution names the candle period the watcher publishes. The
// architecture doc tracks multi-resolution per-tick as an open
// question; for Phase D the watcher publishes 1-minute candles only.
const Resolution = time.Minute

// KLine is the minimal kline view enrichment needs. The legacy
// model.KLine fits this shape; using a narrow interface keeps the
// enrichment package free of legacy imports.
type KLine struct {
	Symbol    string
	OpenTime  time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	Timestamp time.Time
}

// Store keeps a rolling per-symbol window of closed candles. Safe for
// concurrent OnKLine + Snapshot calls.
type Store struct {
	mu     sync.RWMutex
	window map[string][]event.Candle
}

// NewStore returns an empty in-memory rolling-window store.
func NewStore() *Store {
	return &Store{window: make(map[string][]event.Candle)}
}

// OnKLine appends a candle to the symbol's window, evicting the oldest
// entry if the window is already at MaxWindow. If the most recent
// stored candle has the same OpenTime as the incoming one (still-open
// bar update from the WS feed), the latest entry is replaced rather
// than duplicated.
func (s *Store) OnKLine(k KLine) {
	candle := event.Candle{
		OpenTime: k.OpenTime,
		Open:     decimal.NewFromFloat(k.Open),
		High:     decimal.NewFromFloat(k.High),
		Low:      decimal.NewFromFloat(k.Low),
		Close:    decimal.NewFromFloat(k.Close),
		Volume:   decimal.NewFromFloat(k.Volume),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	window := s.window[k.Symbol]
	if n := len(window); n > 0 && window[n-1].OpenTime.Equal(candle.OpenTime) {
		window[n-1] = candle
		s.window[k.Symbol] = window
		return
	}
	if len(window) == MaxWindow {
		copy(window, window[1:])
		window[MaxWindow-1] = candle
		s.window[k.Symbol] = window
		return
	}
	s.window[k.Symbol] = append(window, candle)
}

// Snapshot returns a copy of the current window for the symbol along
// with the indicators computed over it. The Candles slice is a fresh
// copy so the caller can stuff it on a MarketTick without worrying
// about subsequent OnKLine mutations.
func (s *Store) Snapshot(symbol string) (event.Candles, event.Indicators) {
	s.mu.RLock()
	src := s.window[symbol]
	if len(src) == 0 {
		s.mu.RUnlock()
		return event.Candles{Resolution: Resolution}, event.Indicators{}
	}
	candles := make([]event.Candle, len(src))
	copy(candles, src)
	s.mu.RUnlock()

	return event.Candles{
			Resolution: Resolution,
			Series:     candles,
		},
		ComputeIndicators(candles)
}

// Len reports the number of candles currently buffered for the symbol.
// Useful for diagnostics; the strategy itself reads Snapshot.
func (s *Store) Len(symbol string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.window[symbol])
}
