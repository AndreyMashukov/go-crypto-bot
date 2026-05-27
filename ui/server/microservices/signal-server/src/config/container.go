package config

import (
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/AndreyMashukov/go-crypto-bot/ui/server/microservices/signal-server/src/client"
	"github.com/AndreyMashukov/go-crypto-bot/ui/server/microservices/signal-server/src/http"
	"github.com/AndreyMashukov/go-crypto-bot/ui/server/microservices/signal-server/src/model"
	"github.com/AndreyMashukov/go-crypto-bot/ui/server/microservices/signal-server/src/repository"
	"github.com/AndreyMashukov/go-crypto-bot/ui/server/microservices/signal-server/src/service"
	"github.com/AndreyMashukov/go-crypto-bot/ui/server/microservices/signal-server/src/utils"
	"log"
	"os"
	"time"
)

func InitServiceContainer() Container {
	clickhouseDb := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{os.Getenv("CLICKHOUSE_DSN")},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: os.Getenv("CLICKHOUSE_PASSWORD"),
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 30 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Protocol: clickhouse.HTTP,
	})
	chErr := clickhouseDb.Ping()

	if chErr != nil {
		log.Panic(fmt.Sprintf("[Stat DB] Clickhouse can't connect: %s", chErr.Error()))
	}

	clickhouseDb.SetMaxIdleConns(64)
	clickhouseDb.SetMaxOpenConns(64)
	clickhouseDb.SetConnMaxLifetime(time.Minute)

	tradeRepository := repository.TradeRepository{
		DB: clickhouseDb,
	}

	signalService := service.SignalService{
		TradeRepository: &tradeRepository,
		Formatter:       &utils.Formatter{},
		AutoTradeClient: &client.AutoTradeClient{
			InternalAPIToken: os.Getenv("INTERNAL_SERVICE_TOKEN"),
		},
		TradingIntervalMinutes:  30,
		MinAllowedProfitPercent: model.Percent(0.50),
	}

	return Container{
		SignalService: &signalService,
		StatsController: &http.StatsController{
			TradeRepository: &tradeRepository,
		},
	}
}

type Container struct {
	SignalService   *service.SignalService
	StatsController *http.StatsController
}

func (c *Container) Start() {
	for {
		c.SignalService.GenerateSignals()
		// Generate new signals every minute
		time.Sleep(time.Minute)
	}
}
