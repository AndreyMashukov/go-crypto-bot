package transport_test

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/AndreyMashukov/go-crypto-bot/server/shared/transport"
)

// Without a live Redis we cannot exercise PUBLISH/PSUBSCRIBE end-to-end
// at unit-test scope, but we can at least guard against shape regressions:
// the constructors must return non-nil values, the subscriber's Ticks()
// channel must be readable, and Run must return promptly when ctx is
// cancelled (even when the upstream Redis is unreachable, the failure
// must be observable through ctx.Err()).

func TestNewRedisPublisherShape(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	defer func() { _ = rdb.Close() }()

	pub := transport.NewRedisPublisher(rdb, 5*time.Millisecond)
	if pub == nil {
		t.Fatal("NewRedisPublisher returned nil")
	}
}

func TestNewRedisSubscriberShape(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	defer func() { _ = rdb.Close() }()

	sub := transport.NewRedisSubscriber(rdb, "", 0, 0, 0, nil)
	if sub == nil {
		t.Fatal("NewRedisSubscriber returned nil")
	}
	if sub.Ticks() == nil {
		t.Fatal("Ticks() returned nil channel")
	}
}

func TestRedisSubscriberRunCancels(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:1",
		DialTimeout: 50 * time.Millisecond,
	})
	defer func() { _ = rdb.Close() }()

	sub := transport.NewRedisSubscriber(rdb, "", 4, 16, 100*time.Millisecond, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- sub.Run(ctx) }()

	select {
	case <-done:
		// either ctx.Err() or upstream-closed; both are acceptable.
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after ctx cancel + dial fail")
	}
}
