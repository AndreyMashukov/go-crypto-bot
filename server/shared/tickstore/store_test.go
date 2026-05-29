package tickstore_test

import (
	"sync"
	"testing"

	"github.com/AndreyMashukov/go-crypto-bot/server/shared/event"
	"github.com/AndreyMashukov/go-crypto-bot/server/shared/tickstore"
)

func TestLatestMissingReturnsNotOk(t *testing.T) {
	s := tickstore.NewInMemory()
	_, ok := s.Latest("BTCUSDT")
	if ok {
		t.Fatal("Latest(missing) returned ok=true; want false")
	}
}

func TestSetThenLatest(t *testing.T) {
	s := tickstore.NewInMemory()
	s.Set(event.MarketTick{Symbol: "BTCUSDT", SequenceID: 7})
	got, ok := s.Latest("BTCUSDT")
	if !ok {
		t.Fatal("Latest after Set returned ok=false")
	}
	if got.SequenceID != 7 {
		t.Errorf("SequenceID = %d, want 7", got.SequenceID)
	}
}

func TestSetReplaces(t *testing.T) {
	s := tickstore.NewInMemory()
	s.Set(event.MarketTick{Symbol: "BTCUSDT", SequenceID: 1})
	s.Set(event.MarketTick{Symbol: "BTCUSDT", SequenceID: 2})
	got, _ := s.Latest("BTCUSDT")
	if got.SequenceID != 2 {
		t.Errorf("SequenceID = %d, want 2 (Set replaces)", got.SequenceID)
	}
}

func TestConcurrentSetLatestDoesNotRace(_ *testing.T) {
	s := tickstore.NewInMemory()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(seq uint64) {
			defer wg.Done()
			s.Set(event.MarketTick{Symbol: "BTCUSDT", SequenceID: seq})
		}(uint64(i))
	}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = s.Latest("BTCUSDT")
		}()
	}
	wg.Wait()
}
