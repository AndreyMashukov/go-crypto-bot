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
	"golang.org/x/sync/errgroup"

	"github.com/AndreyMashukov/go-crypto-bot/server/shared/metrics"
	"github.com/AndreyMashukov/go-crypto-bot/server/shared/transport"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/client"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/config"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/model"
)

const (
	shutdownDeadline = 5 * time.Second
	startupGrace     = 10 * time.Second
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
	subscriber := transport.NewRedisSubscriber(container.Rdb, transport.PSubscribePattern, 16, 256, 10*time.Second, logger)

	group, groupCtx := errgroup.WithContext(rootCtx)

	group.Go(func() error {
		serveErr := metrics.Serve(groupCtx, container.MetricsAddr)
		if serveErr != nil && !errors.Is(serveErr, context.Canceled) {
			return fmt.Errorf("metrics server: %w", serveErr)
		}
		return nil
	})

	group.Go(func() error {
		subErr := subscriber.Run(groupCtx)
		if subErr != nil && !errors.Is(subErr, context.Canceled) {
			return fmt.Errorf("tick subscriber: %w", subErr)
		}
		return nil
	})

	group.Go(func() error {
		for tick := range subscriber.Ticks() {
			container.LatestTicks.Set(tick)
		}
		return nil
	})

	group.Go(func() error {
		container.MakerService.RecoverOrders()
		waitErr := container.TimeService.WaitSecondsCtx(groupCtx, int64(startupGrace/time.Second))
		if waitErr != nil {
			return waitErr
		}
		container.MakerService.StartTrade(groupCtx)
		<-groupCtx.Done()
		return nil
	})

	err = group.Wait()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownDeadline)
	defer cancel()
	<-shutdownCtx.Done()

	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	log.Println("market-trader: shutdown complete")
	return nil
}
