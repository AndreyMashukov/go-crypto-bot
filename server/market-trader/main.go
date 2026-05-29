// Phase H entry point for the market-trader binary.
//
// Subscribes to MarketTicks on Redis pub/sub (ticks.*), feeds them
// into its local tickstore, runs the trading decision loop on top of
// the StrategyFacade, and serves the admin HTTP API. Owns zero
// exchange WebSocket connections; the watcher does all the WS work
// and publishes ticks for the trader to consume.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"github.com/AndreyMashukov/go-crypto-bot/server/shared/transport"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/client"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/config"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/model"
)

func main() {
	if err := run(); err != nil {
		log.Printf("market-trader exited with error: %s", err.Error())
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
	container.StartHttpServer()

	log.Printf("market-trader [%s] initialised", container.CurrentBot.BotUuid)

	usdtBalance, err := container.BalanceService.GetAssetBalance("USDT", false)
	if err != nil {
		log.Printf("Balance check error: %s", err.Error())
		if err.Error() == model.BinanceErrorInvalidAPIKeyOrPermissions {
			container.CallbackManager.Error(
				*container.CurrentBot,
				model.BinanceErrorInvalidAPIKeyOrPermissions,
				"Please check API Key permissions or IP address binding",
				true,
			)
		}
		return err
	}
	log.Printf("API Key permission check passed, balance is: %.2f USDT", usdtBalance)

	container.PythonMLBridge.StartAutoLearn()

	if binance, ok := container.Binance.(*client.Binance); ok {
		binance.APIKeyCheckCompleted = true
	}
	if bybit, ok := container.Binance.(*client.ByBit); ok {
		bybit.APIKeyCheckCompleted = true
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	subscriber := transport.NewRedisSubscriber(container.Rdb, "", 16, 256, 10*time.Second, logger)
	go func() {
		if err := subscriber.Run(rootCtx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("market-trader: tick subscriber stopped: %s", err.Error())
		}
	}()
	go func() {
		for tick := range subscriber.Ticks() {
			container.LatestTicks.Set(tick)
		}
	}()

	container.MakerService.RecoverOrders()
	container.TimeService.WaitSeconds(10)
	container.MakerService.StartTrade()

	<-rootCtx.Done()
	log.Printf("market-trader: shutdown signal received: %s", rootCtx.Err().Error())

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	<-shutdownCtx.Done()
	if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
		log.Println("market-trader: shutdown deadline reached, exiting")
	}
	return nil
}
