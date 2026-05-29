// Package tickbuffer is the trader-side coalescing buffer for incoming
// MarketTicks. The Redis pub/sub subscriber Puts every received tick
// into the buffer; the StrategyFacade reads Latest(symbol) before each
// decision. Per-symbol dedup + field-level enrichment means a partial
// tick (e.g. depth update without a Price) does NOT clobber a fresh
// price, and burst traffic for a single symbol collapses into one
// merged view that is what the strategy sees on its next tick.
package tickbuffer

import (
	"log"
	"sync"

	"github.com/AndreyMashukov/go-crypto-bot/server/shared/event"
)

// MarketTickBufferInterface is the contract every tick consumer reads
// through. Implementations must be safe for concurrent Put + Latest
// from many goroutines.
type MarketTickBufferInterface interface {
	// Put inserts the tick when the symbol is absent, otherwise merges
	// the incoming tick into the existing entry under the package's
	// enrichment rules (see merge below).
	Put(t event.MarketTick)
	// Latest returns the merged view for the symbol; ok is false when
	// no producer has yet Put a tick for it.
	Latest(symbol string) (event.MarketTick, bool)
}

// InMemory is the production implementation: a per-symbol map of the
// most-recent merged MarketTick guarded by sync.RWMutex. Lookups take
// the read lock; Puts take the write lock for one map read + one merge
// + one map write.
type InMemory struct {
	mu     sync.RWMutex
	latest map[string]event.MarketTick
}

// NewInMemory returns an empty buffer ready for Put + Latest.
func NewInMemory() *InMemory {
	return &InMemory{latest: make(map[string]event.MarketTick)}
}

// Put either inserts the tick or merges it into the existing entry.
// Identity mismatch (incoming.Exchange != existing.Exchange) is logged
// and the incoming wins — this should not happen in practice because
// the buffer is keyed on Symbol, but a defensive log makes a routing
// bug visible instead of silently corrupting state.
func (b *InMemory) Put(t event.MarketTick) {
	b.mu.Lock()
	defer b.mu.Unlock()
	existing, ok := b.latest[t.Symbol]
	if !ok {
		b.latest[t.Symbol] = t
		return
	}
	if existing.Exchange != t.Exchange {
		log.Printf(
			"tickbuffer: exchange mismatch for symbol %s: existing=%s incoming=%s; taking incoming",
			t.Symbol, existing.Exchange, t.Exchange,
		)
	}
	b.latest[t.Symbol] = merge(existing, t)
}

// Latest returns the merged view for symbol or (zero, false) when the
// symbol has not been seen yet. The returned MarketTick is a value
// copy — callers may safely retain it across further Puts.
func (b *InMemory) Latest(symbol string) (event.MarketTick, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	t, ok := b.latest[symbol]
	return t, ok
}

// merge applies the enrichment rules from the Phase I plan. See the
// table in the plan file for the per-field semantics.
func merge(existing, incoming event.MarketTick) event.MarketTick {
	out := existing

	if incoming.EventTime.After(out.EventTime) {
		out.EventTime = incoming.EventTime
	}
	if incoming.SequenceID > out.SequenceID {
		out.SequenceID = incoming.SequenceID
	}
	out.IngestedAt = incoming.IngestedAt
	out.Source = incoming.Source

	if !incoming.Price.IsZero() {
		out.Price = incoming.Price
	}
	if !incoming.BestBid.IsZero() {
		out.BestBid = incoming.BestBid
	}
	if !incoming.BestAsk.IsZero() {
		out.BestAsk = incoming.BestAsk
	}
	if incoming.SpreadBps != 0 {
		out.SpreadBps = incoming.SpreadBps
	}
	if !incoming.Volume24h.IsZero() {
		out.Volume24h = incoming.Volume24h
	}

	if len(incoming.Candles.Series) > 0 {
		out.Candles.Series = incoming.Candles.Series
	}
	if incoming.Candles.Resolution != 0 {
		out.Candles.Resolution = incoming.Candles.Resolution
	}

	if hasIndicatorData(incoming.Indicators) {
		out.Indicators = incoming.Indicators
	}

	if incoming.OpenPosition.AsOf.After(out.OpenPosition.AsOf) {
		out.OpenPosition = incoming.OpenPosition
	}

	if hasRiskEnvelopeData(incoming.RiskEnvelope) {
		out.RiskEnvelope = incoming.RiskEnvelope
	}

	return out
}

func hasIndicatorData(i event.Indicators) bool {
	return i.RSI14 != 0 || !i.SMA20.IsZero() || !i.SMA50.IsZero() || !i.VWMA20.IsZero()
}

func hasRiskEnvelopeData(r event.RiskEnvelope) bool {
	return r.KillSwitchActive ||
		!r.MaxPositionUSDT.IsZero() ||
		!r.DailyLossBudgetRemaining.IsZero()
}
