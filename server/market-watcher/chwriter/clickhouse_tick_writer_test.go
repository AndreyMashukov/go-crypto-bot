package chwriter_test

import (
	"testing"

	"github.com/AndreyMashukov/go-crypto-bot/server/market-watcher/chwriter"
	"github.com/AndreyMashukov/go-crypto-bot/server/shared/event"
)

// Without a live ClickHouse we cannot exercise the INSERT path at
// unit-test scope (the upstream tests cover that). We can guard the
// drop-on-full backpressure invariant: when the buffer is full, Enqueue
// must call OnDrop and not block the caller.

func TestEnqueueDropOnFull(t *testing.T) {
	w := chwriter.New(nil, chwriter.Config{
		BufferSize: 2,
		BatchSize:  10,
	}, nil)

	drops := 0
	w.OnDrop = func(_ string) { drops++ }

	for i := 0; i < 10; i++ {
		w.Enqueue(event.MarketTick{Symbol: "BTCUSDT"})
	}

	if drops == 0 {
		t.Fatal("OnDrop never fired; expected drops when overrun")
	}
	if drops > 8 {
		t.Fatalf("too many drops: %d; buffer size was 2 so first 2 should be queued", drops)
	}
}

func TestEnqueueNoDropWhenUnderCapacity(t *testing.T) {
	w := chwriter.New(nil, chwriter.Config{
		BufferSize: 100,
		BatchSize:  10,
	}, nil)

	drops := 0
	w.OnDrop = func(_ string) { drops++ }

	for i := 0; i < 50; i++ {
		w.Enqueue(event.MarketTick{Symbol: "BTCUSDT"})
	}

	if drops != 0 {
		t.Fatalf("OnDrop fired %d time(s) while buffer was undersized", drops)
	}
}
