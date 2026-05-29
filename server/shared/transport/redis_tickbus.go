package transport

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/AndreyMashukov/go-crypto-bot/server/shared/event"
)

// RedisPublisher publishes encoded MarketTicks to ticks.<exchange>.<symbol>.
// Publish has a tight per-call deadline (default 20 ms) so the WS reader
// never blocks on a slow Redis hop — a stalled publish drops the tick
// at the metric layer, the next one arrives in milliseconds anyway.
type RedisPublisher struct {
	rdb            *redis.Client
	publishTimeout time.Duration
}

// NewRedisPublisher wires a Redis client into the Publisher seam. Pass
// publishTimeout=0 for the default 20 ms ceiling.
func NewRedisPublisher(rdb *redis.Client, publishTimeout time.Duration) *RedisPublisher {
	if publishTimeout <= 0 {
		publishTimeout = 20 * time.Millisecond
	}
	return &RedisPublisher{rdb: rdb, publishTimeout: publishTimeout}
}

// Publish encodes the tick and PUBLISHes it. The deadline is derived
// from the caller's ctx so a parent SIGTERM also short-circuits the
// publish.
func (p *RedisPublisher) Publish(ctx context.Context, t event.MarketTick) error {
	payload, err := Encode(t)
	if err != nil {
		return err
	}
	publishCtx, cancel := context.WithTimeout(ctx, p.publishTimeout)
	defer cancel()
	return p.rdb.Publish(publishCtx, Channel(t.Exchange, t.Symbol), payload).Err()
}

// RedisSubscriber consumes PSubscribePattern through go-redis, decodes
// each payload, and hands the result off into a small in-memory channel
// using drop-on-full semantics. A latest-wins backpressure policy is
// the correct one for tick data: a stale tick is worthless for trading,
// so when the worker is behind, dropping the queued tick and waiting
// for the next one is better than buffering.
//
// Calls Run from its own goroutine; consumers read Ticks() until the
// channel is closed (on Run return).
type RedisSubscriber struct {
	rdb            *redis.Client
	pattern        string
	ticks          chan event.MarketTick
	log            *slog.Logger
	healthInterval time.Duration
	channelSize    int

	OnDecodeFailure func(channel string, err error)
	OnDrop          func(symbol string)
	OnReconnect     func()
}

// NewRedisSubscriber wires the subscriber. workerBuffer sets the size
// of the latest-wins handoff channel (default 16); channelSize controls
// the go-redis intermediate buffer (default 256). healthInterval pings
// the server so a half-open TCP gets noticed (default 10 s).
func NewRedisSubscriber(
	rdb *redis.Client,
	pattern string,
	workerBuffer int,
	channelSize int,
	healthInterval time.Duration,
	log *slog.Logger,
) *RedisSubscriber {
	if pattern == "" {
		pattern = PSubscribePattern
	}
	if workerBuffer <= 0 {
		workerBuffer = 16
	}
	if channelSize <= 0 {
		channelSize = 256
	}
	if healthInterval <= 0 {
		healthInterval = 10 * time.Second
	}
	if log == nil {
		log = slog.Default()
	}
	return &RedisSubscriber{
		rdb:            rdb,
		pattern:        pattern,
		ticks:          make(chan event.MarketTick, workerBuffer),
		log:            log,
		healthInterval: healthInterval,
		channelSize:    channelSize,
	}
}

// Ticks is the consumer-facing channel. Closed when Run returns.
func (s *RedisSubscriber) Ticks() <-chan event.MarketTick { return s.ticks }

// Run owns the subscription lifetime. It blocks until ctx cancels or
// the upstream channel closes; on return it closes s.ticks so consumers
// unblock cleanly.
func (s *RedisSubscriber) Run(ctx context.Context) error {
	defer close(s.ticks)
	ps := s.rdb.PSubscribe(ctx, s.pattern)
	defer func() { _ = ps.Close() }()

	msgs := ps.Channel(
		redis.WithChannelHealthCheckInterval(s.healthInterval),
		redis.WithChannelSize(s.channelSize),
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case m, ok := <-msgs:
			if !ok {
				return errors.New("redis pubsub channel closed")
			}
			tick, err := Decode([]byte(m.Payload))
			if err != nil {
				if s.OnDecodeFailure != nil {
					s.OnDecodeFailure(m.Channel, err)
				} else {
					s.log.Warn("decode tick failed",
						"channel", m.Channel,
						"err", err.Error(),
					)
				}
				continue
			}
			select {
			case s.ticks <- tick:
			default:
				if s.OnDrop != nil {
					s.OnDrop(tick.Symbol)
				} else {
					s.log.Warn("dropped tick (worker behind)",
						"symbol", tick.Symbol,
					)
				}
			}
		}
	}
}
