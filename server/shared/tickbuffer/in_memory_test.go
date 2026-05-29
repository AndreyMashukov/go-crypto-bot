package tickbuffer

import (
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/AndreyMashukov/go-crypto-bot/server/shared/event"
)

func TestPut_InsertsWhenAbsent(t *testing.T) {
	b := NewInMemory()
	tick := event.MarketTick{
		Exchange:  "binance",
		Symbol:    "BTCUSDT",
		Source:    event.SourceTrade,
		EventTime: time.Unix(1700000000, 0),
		Price:     decimal.NewFromFloat(50000),
	}
	b.Put(tick)
	got, ok := b.Latest("BTCUSDT")
	assert.True(t, ok)
	assert.Equal(t, tick, got)
}

func TestLatest_AbsentReturnsFalse(t *testing.T) {
	b := NewInMemory()
	_, ok := b.Latest("NOPE")
	assert.False(t, ok)
}

func TestPut_DedupsBySymbol(t *testing.T) {
	b := NewInMemory()
	b.Put(event.MarketTick{Exchange: "binance", Symbol: "BTCUSDT", Price: decimal.NewFromFloat(50000)})
	b.Put(event.MarketTick{Exchange: "binance", Symbol: "BTCUSDT", Price: decimal.NewFromFloat(50100)})
	b.Put(event.MarketTick{Exchange: "binance", Symbol: "ETHUSDT", Price: decimal.NewFromFloat(3000)})

	btc, _ := b.Latest("BTCUSDT")
	assert.True(t, btc.Price.Equal(decimal.NewFromFloat(50100)), "merged BTC should hold latest price")

	eth, _ := b.Latest("ETHUSDT")
	assert.True(t, eth.Price.Equal(decimal.NewFromFloat(3000)), "ETH untouched by BTC writes")
}

func TestMerge_KeepsExistingPriceWhenIncomingIsZero(t *testing.T) {
	b := NewInMemory()
	b.Put(event.MarketTick{Exchange: "binance", Symbol: "BTCUSDT", Price: decimal.NewFromFloat(50000)})
	// A depth-only tick with Price=0 must NOT wipe the fresh price.
	b.Put(event.MarketTick{
		Exchange: "binance", Symbol: "BTCUSDT",
		BestBid: decimal.NewFromFloat(49999), BestAsk: decimal.NewFromFloat(50001),
	})
	got, _ := b.Latest("BTCUSDT")
	assert.True(t, got.Price.Equal(decimal.NewFromFloat(50000)), "Price preserved")
	assert.True(t, got.BestBid.Equal(decimal.NewFromFloat(49999)), "BestBid taken from incoming")
}

func TestMerge_TakesIncomingPriceWhenNonZero(t *testing.T) {
	b := NewInMemory()
	b.Put(event.MarketTick{Exchange: "binance", Symbol: "BTCUSDT", Price: decimal.NewFromFloat(50000)})
	b.Put(event.MarketTick{Exchange: "binance", Symbol: "BTCUSDT", Price: decimal.NewFromFloat(50500)})
	got, _ := b.Latest("BTCUSDT")
	assert.True(t, got.Price.Equal(decimal.NewFromFloat(50500)))
}

func TestMerge_TakesMaxEventTimeAndSequenceID(t *testing.T) {
	b := NewInMemory()
	later := time.Unix(1700000010, 0)
	earlier := time.Unix(1700000000, 0)

	b.Put(event.MarketTick{Exchange: "binance", Symbol: "BTCUSDT", EventTime: later, SequenceID: 100})
	// Out-of-order incoming with smaller times — must not roll back.
	b.Put(event.MarketTick{Exchange: "binance", Symbol: "BTCUSDT", EventTime: earlier, SequenceID: 50})

	got, _ := b.Latest("BTCUSDT")
	assert.True(t, got.EventTime.Equal(later))
	assert.EqualValues(t, 100, got.SequenceID)
}

func TestMerge_TakesNewerOpenPosition(t *testing.T) {
	b := NewInMemory()
	older := event.OpenPosition{Quantity: decimal.NewFromFloat(1.0), AsOf: time.Unix(1700000000, 0)}
	newer := event.OpenPosition{Quantity: decimal.NewFromFloat(2.0), AsOf: time.Unix(1700000010, 0)}

	b.Put(event.MarketTick{Exchange: "binance", Symbol: "BTCUSDT", OpenPosition: newer})
	b.Put(event.MarketTick{Exchange: "binance", Symbol: "BTCUSDT", OpenPosition: older})

	got, _ := b.Latest("BTCUSDT")
	assert.True(t, got.OpenPosition.Quantity.Equal(newer.Quantity), "newer AsOf wins")
}

func TestMerge_TakesCandlesWhenNonEmpty(t *testing.T) {
	b := NewInMemory()
	series := []event.Candle{{OpenTime: time.Unix(1700000000, 0), Close: decimal.NewFromFloat(50000)}}
	b.Put(event.MarketTick{
		Exchange: "binance", Symbol: "BTCUSDT",
		Candles: event.Candles{Resolution: time.Minute, Series: series},
	})
	// Empty Candles incoming must not wipe the existing window.
	b.Put(event.MarketTick{Exchange: "binance", Symbol: "BTCUSDT", Price: decimal.NewFromFloat(50100)})
	got, _ := b.Latest("BTCUSDT")
	assert.Len(t, got.Candles.Series, 1)
	assert.Equal(t, time.Minute, got.Candles.Resolution)
}

func TestConcurrentPut(t *testing.T) {
	b := NewInMemory()
	const goroutines = 64
	const perGoroutine = 200
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				b.Put(event.MarketTick{
					Exchange:   "binance",
					Symbol:     "BTCUSDT",
					SequenceID: uint64(id*perGoroutine + i),
					Price:      decimal.NewFromInt(int64(id*perGoroutine + i)),
				})
			}
		}(g)
	}
	wg.Wait()

	got, ok := b.Latest("BTCUSDT")
	assert.True(t, ok)
	assert.NotZero(t, got.SequenceID, "some tick made it through")
}
