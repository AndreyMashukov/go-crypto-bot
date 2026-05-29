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

	"github.com/AndreyMashukov/go-crypto-bot/server/src/client"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/config"
)

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
		binance.APIKeyCheckCompleted = true
	}
	if bybit, ok := container.Binance.(*client.ByBit); ok {
		bybit.APIKeyCheckCompleted = true
	}

	container.PythonMLBridge.StartAutoLearn()

	if container.IsMasterBot {
		go func() {
			container.MCListener.ListenAll()
		}()
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		container.MarketTradeListener.ListenAll()
	}()

	select {
	case <-rootCtx.Done():
		log.Printf("market-watcher: shutdown signal received: %s", rootCtx.Err().Error())
	case <-done:
		log.Println("market-watcher: MarketTradeListener returned unexpectedly")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	<-shutdownCtx.Done()
	if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
		log.Println("market-watcher: shutdown deadline reached, exiting")
	}
	return nil
}
