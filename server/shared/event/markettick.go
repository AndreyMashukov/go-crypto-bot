// Package event holds the wire-contract types shared between
// market-watcher (which publishes) and market-trader (which subscribes).
// Both binaries depend on this package; neither imports the other.
//
// Compatibility rule: this is an additive-only schema. Fields are never
// removed or repurposed — old strategy versions still consuming v1 ticks
// must keep working alongside a v2 publisher during rolling restarts.
package event

import (
	"time"

	"github.com/shopspring/decimal"
)

// Decimal is the project-wide decimal type. Locked to shopspring/decimal
// for the watcher/trader hot path. Aliased here so callers depend on the
// shared package rather than on shopspring directly.
type Decimal = decimal.Decimal

// MarketTick is the only thing the trader consumes. Immutable,
// copy-cheap, self-sufficient: every field a strategy needs to decide
// is present on the tick. The trader never calls back to the watcher to
// ask "what's the last candle?" — if it's not on the tick, it's not
// part of the decision.
type MarketTick struct {
	// Identity.
	Exchange   string    `msgpack:"exchange"`
	Symbol     string    `msgpack:"symbol"`
	Source     Source    `msgpack:"source"`
	SequenceID uint64    `msgpack:"sequence_id"`
	EventTime  time.Time `msgpack:"event_time"`
	IngestedAt time.Time `msgpack:"ingested_at"`

	// Price view.
	Price     Decimal `msgpack:"price"`
	BestBid   Decimal `msgpack:"best_bid"`
	BestAsk   Decimal `msgpack:"best_ask"`
	SpreadBps int32   `msgpack:"spread_bps"`
	Volume24h Decimal `msgpack:"volume_24h"`

	// Recent history snapshot — pre-computed by the watcher in Phase D.
	Candles    Candles    `msgpack:"candles"`
	Indicators Indicators `msgpack:"indicators"`

	// Account view — only fields safe for this strategy to read,
	// watcher cap'd to per-bot config. Zero-valued in Phase C; populated
	// in Phase D when enrichment lands.
	OpenPosition OpenPosition `msgpack:"open_position"`
	RiskEnvelope RiskEnvelope `msgpack:"risk_envelope"`
}

// Candles carries the recent kline window for the strategy's primary
// resolution. Multi-resolution (1m + 5m + 15m on the same tick) is
// tracked as an open question in §10 of the architecture doc; for now
// the watcher publishes whichever resolution the strategy configured.
type Candles struct {
	Resolution time.Duration `msgpack:"resolution"`
	Series     []Candle      `msgpack:"series"`
}

// Candle is a single closed bar — open / high / low / close / volume —
// over the Candles.Resolution window starting at OpenTime.
type Candle struct {
	OpenTime time.Time `msgpack:"open_time"`
	Open     Decimal   `msgpack:"open"`
	High     Decimal   `msgpack:"high"`
	Low      Decimal   `msgpack:"low"`
	Close    Decimal   `msgpack:"close"`
	Volume   Decimal   `msgpack:"volume"`
}

// Indicators ride on every tick so the strategy reads them in O(1)
// instead of recomputing N times.
type Indicators struct {
	RSI14  float64 `msgpack:"rsi14"`
	SMA20  Decimal `msgpack:"sma20"`
	SMA50  Decimal `msgpack:"sma50"`
	VWMA20 Decimal `msgpack:"vwma20"`
}

// OpenPosition is the bot's current position on this symbol as observed
// by the watcher's user-data stream. AsOf carries the freshness so the
// trader can reject decisions that would race a recent fill.
type OpenPosition struct {
	Quantity      Decimal   `msgpack:"quantity"`
	AvgEntry      Decimal   `msgpack:"avg_entry"`
	UnrealizedPnL Decimal   `msgpack:"unrealized_pnl"`
	LastOrderID   string    `msgpack:"last_order_id"`
	AsOf          time.Time `msgpack:"as_of"`
}

// RiskEnvelope is the per-symbol risk budget snapshot the watcher
// pulled from `CryptoTradeConfig`. KillSwitchActive trumps everything —
// strategy returns early.
type RiskEnvelope struct {
	MaxPositionUSDT          Decimal `msgpack:"max_position_usdt"`
	DailyLossBudgetRemaining Decimal `msgpack:"daily_loss_budget_remaining"`
	KillSwitchActive         bool    `msgpack:"kill_switch_active"`
}
