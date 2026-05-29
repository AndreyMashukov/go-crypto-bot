package publisher_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AndreyMashukov/go-crypto-bot/server/market-watcher/publisher"
	"github.com/AndreyMashukov/go-crypto-bot/server/shared/event"
)

type stubBus struct {
	calls int
	err   error
}

func (s *stubBus) Publish(_ context.Context, _ event.MarketTick) error {
	s.calls++
	return s.err
}

func TestEmitNilTolerant(_ *testing.T) {
	// Both seams optional — Emit must not panic with nothing wired.
	p := publisher.New(nil, nil)
	p.Emit(context.Background(), event.MarketTick{Symbol: "BTCUSDT"})
}

func TestEmitPublishSuccess(t *testing.T) {
	bus := &stubBus{}
	failures := 0

	p := publisher.New(bus, nil)
	p.OnPublishFailure = func(_ string, _ error) { failures++ }

	p.Emit(context.Background(), event.MarketTick{Symbol: "BTCUSDT"})

	if bus.calls != 1 {
		t.Errorf("Publish call count = %d, want 1", bus.calls)
	}
	if failures != 0 {
		t.Errorf("OnPublishFailure fired %d time(s); want 0", failures)
	}
}

func TestEmitPublishFailureRoutesToCallback(t *testing.T) {
	bus := &stubBus{err: errors.New("redis down")}
	gotSymbol := ""
	gotErr := error(nil)

	p := publisher.New(bus, nil)
	p.OnPublishFailure = func(sym string, err error) {
		gotSymbol = sym
		gotErr = err
	}

	p.Emit(context.Background(), event.MarketTick{Symbol: "ETHUSDT"})

	if gotSymbol != "ETHUSDT" {
		t.Errorf("OnPublishFailure symbol = %q, want ETHUSDT", gotSymbol)
	}
	if gotErr == nil {
		t.Error("OnPublishFailure err is nil; want non-nil")
	}
}
