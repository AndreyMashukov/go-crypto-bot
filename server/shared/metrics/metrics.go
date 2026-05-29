// Package metrics owns the Prometheus counters and histograms the
// watcher and trader emit. Both binaries depend on this package; the
// concrete metrics live on prometheus.DefaultRegisterer so promhttp.
// Handler() picks them up automatically.
//
// Naming follows the architecture doc §8.3:
//   - tick_publish_failure_total{symbol}   counter — pub/sub Publish error
//   - tick_drop_total{symbol}              counter — drop-on-full at the
//     chwriter buffer or
//     trader handoff
//   - tick_decode_failure_total{channel}   counter — subscriber decode err
//   - tick_consume_lag_seconds{symbol}     histogram — EventTime → recv
//   - handle_duration_seconds              histogram — strategy OnTick
//   - order_place_duration_seconds{outcome}histogram — REST place latency
//   - pubsub_reconnect_total               counter — go-redis ps reconnect
//   - pnl_realized_total{symbol}           counter — closed-position PnL
//   - position_open_seconds{symbol}        histogram — open → close
//
// Metric *names* are stable contracts the dashboards rely on. Do not
// rename without coordinating with the Prometheus scrape config and the
// chart components in ui/app.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// TickPublishFailure counts pub/sub Publish errors per symbol.
var TickPublishFailure = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "tick_publish_failure_total",
	Help: "MarketTick publish-to-Redis failures.",
}, []string{"symbol"})

// TickDrop counts drop-on-full events at the chwriter buffer or the
// trader's per-worker handoff. A non-zero rate is the classic backpressure
// alert: the consumer is too slow for its market.
var TickDrop = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "tick_drop_total",
	Help: "MarketTick drops at the chwriter buffer or trader handoff.",
}, []string{"symbol"})

// TickDecodeFailure counts subscriber-side decode failures per channel.
var TickDecodeFailure = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "tick_decode_failure_total",
	Help: "MarketTick decode failures observed by the subscriber.",
}, []string{"channel"})

// TickConsumeLag observes EventTime → receive-time latency in seconds.
var TickConsumeLag = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "tick_consume_lag_seconds",
	Help:    "EventTime → subscriber-receive latency in seconds.",
	Buckets: prometheus.ExponentialBuckets(0.001, 2, 12),
}, []string{"symbol"})

// HandleDuration observes the strategy's OnTick wall-time.
var HandleDuration = promauto.NewHistogram(prometheus.HistogramOpts{
	Name:    "handle_duration_seconds",
	Help:    "Strategy OnTick wall-time in seconds.",
	Buckets: prometheus.ExponentialBuckets(0.0005, 2, 12),
})

// OrderPlaceDuration observes order-router round-trip per outcome.
var OrderPlaceDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "order_place_duration_seconds",
	Help:    "Order-router REST round-trip in seconds, labelled by outcome (success / error / timeout).",
	Buckets: prometheus.ExponentialBuckets(0.005, 2, 12),
}, []string{"outcome"})

// PubsubReconnect counts go-redis pub/sub reconnects.
var PubsubReconnect = promauto.NewCounter(prometheus.CounterOpts{
	Name: "pubsub_reconnect_total",
	Help: "go-redis pub/sub reconnection events.",
})

// PnLRealized accumulates realised P&L per symbol when a position closes.
var PnLRealized = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "pnl_realized_total",
	Help: "Realised P&L in quote currency, summed over closed positions.",
}, []string{"symbol"})

// PositionOpenSeconds observes how long each position stayed open.
var PositionOpenSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "position_open_seconds",
	Help:    "Time a position stayed open, from entry to exit, in seconds.",
	Buckets: prometheus.ExponentialBuckets(60, 2, 12),
}, []string{"symbol"})
