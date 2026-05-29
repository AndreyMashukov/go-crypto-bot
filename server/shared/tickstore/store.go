// Package tickstore owns the per-process latest-tick map. The watcher
// writes the freshly enriched MarketTick into the store right after
// publishing it; the strategy facade reads the same store to make
// decisions. Both binaries will keep their own concrete store after
// the Phase H split — the watcher's stays write-only for itself,
// the trader's gets filled by its subscriber off the pub/sub stream.
package tickstore

import (
	"sync"

	"github.com/AndreyMashukov/go-crypto-bot/server/shared/event"
)

// Store is the consumer-facing interface. Tests inject a fake that
// returns whatever shape the assertion needs.
type Store interface {
	Set(t event.MarketTick)
	Latest(symbol string) (event.MarketTick, bool)
}

// InMemory is the production implementation. A single Store instance
// is safe for concurrent Set + Latest from many goroutines.
type InMemory struct {
	mu   sync.RWMutex
	last map[string]event.MarketTick
}

// NewInMemory returns an empty store.
func NewInMemory() *InMemory {
	return &InMemory{last: make(map[string]event.MarketTick)}
}

// Set replaces the latest tick for tick.Symbol. The previous value (if
// any) is discarded — the strategy reads "current state", not history.
func (s *InMemory) Set(t event.MarketTick) {
	s.mu.Lock()
	s.last[t.Symbol] = t
	s.mu.Unlock()
}

// Latest returns the most recent tick for the symbol. ok is false
// when the watcher has not yet produced a tick for this symbol (window
// still empty after restart, symbol not enabled, etc.).
func (s *InMemory) Latest(symbol string) (event.MarketTick, bool) {
	s.mu.RLock()
	t, ok := s.last[symbol]
	s.mu.RUnlock()
	return t, ok
}
