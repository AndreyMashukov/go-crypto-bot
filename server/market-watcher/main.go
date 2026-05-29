// Phase H entry point for the market-watcher binary.
//
// Owns every exchange WebSocket connection, runs the rolling
// enrichment store, and publishes every observed kline as a MarketTick
// to Redis pub/sub + ClickHouse market_tick. Knows nothing about
// orders or risk envelopes — the trader picks ticks up off the wire.
//
// In master-mode it also runs the MC capitalisation listener; the
// shared container creates the listener regardless, the runtime loop
// here decides whether to spin it.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"

	"github.com/AndreyMashukov/go-crypto-bot/server/shared/metrics"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/client"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/config"
)

const shutdownDeadline = 5 * time.Second

func main() {
	if err := run(); err != nil {
		log.Printf("market-watcher exited with error: %s", err.Error())
		os.Exit(1)
	}
}

func run() error {
	rootCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt, syscall.SIGTERM,
	)
	defer stop()

	pwd, _ := os.Getwd()
	if _, err := os.Stat(fmt.Sprintf("%s/.env", pwd)); err == nil {
		log.Println(".env is found, loading variables...")
		if err := godotenv.Load(); err != nil {
			log.Println(err)
		}
	}

	container := config.InitServiceContainer()
	defer func() { _ = container.Db.Close() }()

	container.PingDB()
	container.PythonMLBridge.Initialize()
	defer container.PythonMLBridge.Finalize()

	log.Printf("market-watcher [%s] initialised", container.CurrentBot.BotUuid)

	if binance, ok := container.Binance.(*client.Binance); ok {
		binance.Connect(container.BinanceWSAddr)
		binance.APIKeyCheckCompleted = true
	}
	if bybit, ok := container.Binance.(*client.ByBit); ok {
		bybit.APIKeyCheckCompleted = true
	}

	container.PythonMLBridge.StartAutoLearn()

	group, groupCtx := errgroup.WithContext(rootCtx)

	group.Go(func() error {
		writerErr := container.TickCHWriter.Run(groupCtx)
		if writerErr != nil && !errors.Is(writerErr, context.Canceled) {
			return fmt.Errorf("clickhouse tick writer: %w", writerErr)
		}
		return nil
	})

	group.Go(func() error {
		serveErr := metrics.Serve(groupCtx, container.MetricsAddr)
		if serveErr != nil && !errors.Is(serveErr, context.Canceled) {
			return fmt.Errorf("metrics server: %w", serveErr)
		}
		return nil
	})

	if container.IsMasterBot {
		group.Go(func() error {
			container.MCListener.ListenAll(groupCtx)
			return nil
		})
	}

	group.Go(func() error {
		container.MarketTradeListener.ListenAll(groupCtx)
		return nil
	})

	err := group.Wait()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownDeadline)
	defer cancel()
	container.TickCHWriter.Close()
	<-shutdownCtx.Done()

	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	log.Println("market-watcher: shutdown complete")
	return nil
}
