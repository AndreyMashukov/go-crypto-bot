// Package transport carries the wire-level Publisher/Subscriber
// interfaces and the binary encoding used between market-watcher and
// market-trader. Both binaries depend on this package; the redis
// implementation lives alongside in this same package.
//
// Wire format: leading byte is the schema version, then a msgpack-
// encoded event.MarketTick.
//
// Compatibility rules:
//   - Additive field changes keep SchemaVersion unchanged. msgpack
//     ignores unknown trailing fields on decode, so a new field on
//     MarketTick rolls out without bumping the version.
//   - Bumping SchemaVersion means a wire-incompatible change. The
//     trader must be upgraded BEFORE the watcher, otherwise every
//     incoming tick will be dropped at Decode.
package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/AndreyMashukov/go-crypto-bot/server/shared/event"
)

// SchemaVersion is the leading byte stamped on every wire-encoded tick.
// Bump it when a backwards-incompatible change ships — never when
// adding a field (msgpack handles that natively).
const SchemaVersion byte = 1

// Publisher is the watcher's outbound seam. Implementations must be
// safe for concurrent calls from multiple goroutines (one publisher
// instance shared across all symbols' enrichment paths).
type Publisher interface {
	Publish(ctx context.Context, t event.MarketTick) error
}

// Subscriber is the trader's inbound seam. Run drives the network read
// loop until ctx cancels; Ticks delivers decoded MarketTick values
// drop-on-full into a small buffered channel (latest-wins).
type Subscriber interface {
	Run(ctx context.Context) error
	Ticks() <-chan event.MarketTick
}

// Channel computes the Redis pub/sub channel name for a tick. Single
// source of truth for the naming convention so the publisher and
// subscriber agree by construction.
func Channel(exchange, symbol string) string {
	return fmt.Sprintf("ticks.%s.%s", exchange, symbol)
}

// PSubscribePattern is the wildcard the trader uses to consume every
// symbol on every exchange. Strategies that only care about a fixed
// whitelist can use Subscribe + Channel(...) per symbol instead.
const PSubscribePattern = "ticks.*"

// Encode stamps SchemaVersion + msgpack-encodes the tick into a single
// payload byte slice ready for Redis.Publish.
func Encode(t event.MarketTick) ([]byte, error) {
	body, err := msgpack.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("encode market tick: %w", err)
	}
	out := make([]byte, 0, len(body)+1)
	out = append(out, SchemaVersion)
	out = append(out, body...)
	return out, nil
}

// Decode strips the version byte and msgpack-decodes the rest into a
// MarketTick. Returns an error when the schema version is unknown so
// the caller can drop the message and metric it.
func Decode(payload []byte) (event.MarketTick, error) {
	var tick event.MarketTick
	if len(payload) == 0 {
		return tick, errors.New("decode market tick: empty payload")
	}
	version := payload[0]
	if version != SchemaVersion {
		return tick, fmt.Errorf("decode market tick: unknown schema version %d", version)
	}
	dec := msgpack.NewDecoder(bytes.NewReader(payload[1:]))
	dec.UseLooseInterfaceDecoding(true)
	if err := dec.Decode(&tick); err != nil {
		return tick, fmt.Errorf("decode market tick: %w", err)
	}
	return tick, nil
}
