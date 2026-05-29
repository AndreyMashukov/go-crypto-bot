package transport_test

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/AndreyMashukov/go-crypto-bot/server/shared/event"
	"github.com/AndreyMashukov/go-crypto-bot/server/shared/transport"
)

func TestChannel(t *testing.T) {
	got := transport.Channel("binance", "BTCUSDT")
	if got != "ticks.binance.BTCUSDT" {
		t.Fatalf("Channel(binance, BTCUSDT) = %q, want ticks.binance.BTCUSDT", got)
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	original := event.MarketTick{
		Exchange:   "binance",
		Symbol:     "BTCUSDT",
		Source:     event.SourceTrade,
		SequenceID: 42,
		EventTime:  time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC),
		IngestedAt: time.Date(2026, 5, 29, 12, 0, 1, 0, time.UTC),
		Price:      decimal.NewFromFloat(60123.456789),
		BestBid:    decimal.NewFromFloat(60123.40),
		BestAsk:    decimal.NewFromFloat(60123.50),
		SpreadBps:  17,
		Volume24h:  decimal.NewFromInt(987654),
	}

	payload, err := transport.Encode(original)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if payload[0] != transport.SchemaVersion {
		t.Fatalf("schema version byte = %d, want %d", payload[0], transport.SchemaVersion)
	}

	decoded, err := transport.Decode(payload)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if decoded.Exchange != original.Exchange ||
		decoded.Symbol != original.Symbol ||
		decoded.Source != original.Source ||
		decoded.SequenceID != original.SequenceID ||
		decoded.SpreadBps != original.SpreadBps {
		t.Errorf("identity fields drifted: got %+v, want %+v", decoded, original)
	}
	if !decoded.EventTime.Equal(original.EventTime) {
		t.Errorf("EventTime drift: got %v, want %v", decoded.EventTime, original.EventTime)
	}
	if !decoded.Price.Equal(original.Price) {
		t.Errorf("Price drift: got %s, want %s", decoded.Price, original.Price)
	}
	if !decoded.BestBid.Equal(original.BestBid) {
		t.Errorf("BestBid drift: got %s, want %s", decoded.BestBid, original.BestBid)
	}
	if !decoded.Volume24h.Equal(original.Volume24h) {
		t.Errorf("Volume24h drift: got %s, want %s", decoded.Volume24h, original.Volume24h)
	}
}

func TestDecodeEmptyPayload(t *testing.T) {
	_, err := transport.Decode(nil)
	if err == nil {
		t.Fatal("Decode(nil) returned nil error; want error")
	}
}

func TestDecodeUnknownSchemaVersion(t *testing.T) {
	_, err := transport.Decode([]byte{255})
	if err == nil {
		t.Fatal("Decode of bogus version returned nil error; want error")
	}
}
