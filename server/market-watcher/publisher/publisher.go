// Package publisher is the watcher-internal seam that fans every
// MarketTick out to both the live pub/sub transport (the trader's hot
// path) and the async ClickHouse writer (offline analysis). Both calls
// are non-blocking from the caller's perspective — the pub/sub publish
// has its own per-call deadline inside the transport, and the
// ClickHouse writer is fire-and-forget into a buffered queue.
package publisher

import (
	"context"

	"github.com/AndreyMashukov/go-crypto-bot/server/market-watcher/chwriter"
	"github.com/AndreyMashukov/go-crypto-bot/server/shared/event"
	"github.com/AndreyMashukov/go-crypto-bot/server/shared/transport"
)

// Publisher is the watcher-side entry point. Emit publishes the tick to
// the trader and asynchronously enqueues it for ClickHouse — one call,
// two side-effects, neither one blocks the caller.
type Publisher struct {
	bus    transport.Publisher
	writer *chwriter.TickWriter

	OnPublishFailure func(symbol string, err error)
}

// New wires both side-effects under one entry point. Either argument
// may be nil for tests; nil bus skips Publish, nil writer skips Enqueue.
func New(bus transport.Publisher, writer *chwriter.TickWriter) *Publisher {
	return &Publisher{bus: bus, writer: writer}
}

// Emit publishes the tick to pub/sub and enqueues it for ClickHouse.
// Publish errors are NOT returned — the watcher's WS reader cannot
// stall on a Redis hiccup. Failures route through OnPublishFailure so
// they surface in metrics / logs without affecting the caller's loop.
func (p *Publisher) Emit(ctx context.Context, t event.MarketTick) {
	if p.bus != nil {
		if err := p.bus.Publish(ctx, t); err != nil && p.OnPublishFailure != nil {
			p.OnPublishFailure(t.Symbol, err)
		}
	}
	if p.writer != nil {
		p.writer.Enqueue(t)
	}
}
