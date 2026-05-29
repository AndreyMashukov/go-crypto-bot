// Package chwriter holds the watcher's async ClickHouse writer. Every
// tick the watcher emits is dual-pathed: the live pub/sub channel feeds
// the trader's hot path, and this writer batches the same tick into
// ClickHouse market_tick for offline analysis (backtests, ML training).
// Slowness on the ClickHouse side must NEVER block the publish hot
// path, so Enqueue is non-blocking and silently drops on full.
package chwriter

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/AndreyMashukov/go-crypto-bot/server/shared/event"
)

// Defaults tuned for ~10 ticks/s/symbol × 50 symbols workload.
const (
	defaultBufferSize    = 4096
	defaultBatchSize     = 1024
	defaultFlushInterval = 250 * time.Millisecond
	defaultInsertTimeout = 2 * time.Second
)

// Config is the writer's tunable knobs. Zero-valued fields fall back to
// the package defaults. ClickHouse-side behaviour (compression, async
// inserts, etc.) lives in the *sql.DB passed in by the caller.
type Config struct {
	BufferSize    int
	BatchSize     int
	FlushInterval time.Duration
	InsertTimeout time.Duration
}

// TickWriter buffers MarketTick rows in memory and flushes them into
// ClickHouse on either a size or time trigger. Enqueue is the hot-path
// entry point; it is non-blocking and drops on full so the caller is
// never coupled to ClickHouse latency.
type TickWriter struct {
	db  *sql.DB
	log *slog.Logger

	queue chan event.MarketTick

	batchSize     int
	flushInterval time.Duration
	insertTimeout time.Duration

	wg sync.WaitGroup

	OnDrop          func(symbol string)
	OnInsertFailure func(rows int, err error)
}

// New wires up a TickWriter. Call Run from a supervised goroutine and
// remember to Close on shutdown so the in-memory queue gets flushed.
func New(db *sql.DB, cfg Config, log *slog.Logger) *TickWriter {
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = defaultBufferSize
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = defaultBatchSize
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = defaultFlushInterval
	}
	if cfg.InsertTimeout <= 0 {
		cfg.InsertTimeout = defaultInsertTimeout
	}
	if log == nil {
		log = slog.Default()
	}
	return &TickWriter{
		db:            db,
		log:           log,
		queue:         make(chan event.MarketTick, cfg.BufferSize),
		batchSize:     cfg.BatchSize,
		flushInterval: cfg.FlushInterval,
		insertTimeout: cfg.InsertTimeout,
	}
}

// Enqueue pushes a tick into the buffer. Non-blocking: if the buffer
// is full, the tick is dropped and OnDrop is invoked. This is the
// correct behaviour for tick data — the next tick arrives in
// milliseconds and a backlogged tick is stale by the time it would be
// written.
func (w *TickWriter) Enqueue(t event.MarketTick) {
	select {
	case w.queue <- t:
	default:
		if w.OnDrop != nil {
			w.OnDrop(t.Symbol)
		} else {
			w.log.Warn("clickhouse buffer full, dropping tick",
				"symbol", t.Symbol,
			)
		}
	}
}

// Run blocks while batching + flushing until ctx cancels. It always
// drains any remaining buffer through one final flush before returning
// so a graceful shutdown loses no ticks that already made it past
// Enqueue.
func (w *TickWriter) Run(ctx context.Context) error {
	w.wg.Add(1)
	defer w.wg.Done()

	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()

	batch := make([]event.MarketTick, 0, w.batchSize)

	flush := func(parent context.Context) {
		if len(batch) == 0 {
			return
		}
		insertCtx, cancel := context.WithTimeout(parent, w.insertTimeout)
		defer cancel()
		if err := w.flush(insertCtx, batch); err != nil {
			if w.OnInsertFailure != nil {
				w.OnInsertFailure(len(batch), err)
			} else {
				w.log.Error("clickhouse insert failed",
					"rows", len(batch),
					"err", err.Error(),
				)
			}
		}
		batch = batch[:0]
	}

	for {
		select {
		case <-ctx.Done():
			// Drain whatever is already queued before we hand back. The
			// parent ctx is already cancelled, so flush gets its own
			// background ctx with the per-insert deadline.
			for {
				select {
				case t := <-w.queue:
					batch = append(batch, t)
					if len(batch) >= w.batchSize {
						flush(context.Background())
					}
				default:
					flush(context.Background())
					return ctx.Err()
				}
			}
		case <-ticker.C:
			flush(ctx)
		case t, ok := <-w.queue:
			if !ok {
				flush(ctx)
				return errors.New("clickhouse tick writer queue closed")
			}
			batch = append(batch, t)
			if len(batch) >= w.batchSize {
				flush(ctx)
			}
		}
	}
}

// Close waits for an in-flight Run to return so callers can sequence a
// clean shutdown after cancelling the parent context.
func (w *TickWriter) Close() { w.wg.Wait() }

func (w *TickWriter) flush(ctx context.Context, batch []event.MarketTick) error {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// clickhouse-go/v2's batch driver parses the table name out of the
	// prepared SQL with a strict regex; multi-line or tab-indented INSERTs
	// trip it up with "cannot get table name from query". Build the SQL by
	// concatenation so each Go source line stays inside the lll budget but
	// the runtime string is still a single uninterrupted INSERT.
	insertSQL := "INSERT INTO default.market_tick (" +
		"exchange, symbol, event_time, ingested_at, sequence_id, source, " +
		"price, best_bid, best_ask, spread_bps, volume_24h" +
		") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	stmt, err := tx.PrepareContext(ctx, insertSQL)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	for i := range batch {
		t := &batch[i]
		if _, err := stmt.ExecContext(ctx,
			t.Exchange,
			t.Symbol,
			t.EventTime.UTC(),
			t.IngestedAt.UTC(),
			t.SequenceID,
			uint8(t.Source),
			t.Price,
			t.BestBid,
			t.BestAsk,
			t.SpreadBps,
			t.Volume24h,
		); err != nil {
			_ = stmt.Close()
			_ = tx.Rollback()
			return err
		}
	}
	if err := stmt.Close(); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
