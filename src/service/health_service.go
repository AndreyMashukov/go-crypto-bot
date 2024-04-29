package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/rafacas/sysstats"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/service/ml"
	"runtime"
	"time"
)

type HealthService struct {
	ExchangeRepository *repository.ExchangeRepository
	BotRepository      *repository.BotRepository
	PythonMLBridge     *ml.PythonMLBridge
	DB                 *sql.DB
	SwapDb             *sql.DB
	RDB                *redis.Client
	Ctx                *context.Context
	Binance            *client.Binance
	CurrentBot         *model.Bot
}

func (h *HealthService) HealthCheck() model.BotHealth {
	updateMap := make(map[string]string)
	orderBookMap := make(map[string]string)

	for _, limit := range h.ExchangeRepository.GetTradeLimits() {
		kLine := h.ExchangeRepository.GetCurrentKline(limit.Symbol)
		dateStringPrice := ""
		dateStringOrderBook := ""
		if kLine != nil {
			dateStringPrice = time.Unix(kLine.UpdatedAt, 0).Format("2006-01-02 15:04:05")
		}

		orderBook := h.ExchangeRepository.GetDepth(limit.Symbol, 20)
		if !orderBook.IsEmpty() {
			dateStringOrderBook = time.Unix(orderBook.UpdatedAt, 0).Format("2006-01-02 15:04:05")
		}

		updateMap[limit.Symbol] = dateStringPrice
		orderBookMap[limit.Symbol] = dateStringOrderBook
	}

	memStats, _ := sysstats.GetMemStats()
	loadAvg, _ := sysstats.GetLoadAvg()

	binanceStatus := model.BinanceStatusOk
	if !h.Binance.Connected {
		binanceStatus = model.BinanceStatusDisconnected
	}
	if h.Binance.WaitMode {
		binanceStatus = model.BinanceStatusBan
	}
	if !h.Binance.APIKeyCheckCompleted {
		binanceStatus = model.BinanceStatusApiKeyCheck
	}

	dbStatus := model.DbStatusOk
	if h.DB.Ping() != nil {
		dbStatus = model.DbStatusFail
	}
	swapDbStatus := model.DbStatusOk
	if h.SwapDb.Ping() != nil {
		swapDbStatus = model.DbStatusFail
	}
	redisStatus := model.RedisStatusOk
	if h.RDB.Ping(*h.Ctx).Err() != nil {
		redisStatus = model.RedisStatusFail
	}
	mlStatus := model.MlStatusReady
	if h.PythonMLBridge.IsLearning() {
		mlStatus = model.MlStatusLearning
	}

	bot := h.BotRepository.GetCurrentBot()

	if bot == nil {
		panic("Current Bot is not found")
	}

	if bot.Id != h.CurrentBot.Id {
		panic(fmt.Sprintf("Wrong BOT ID %d != %d", bot.Id, h.CurrentBot.Id))
	}

	return model.BotHealth{
		Bot:           *bot,
		DbStatus:      dbStatus,
		SwapDbStatus:  swapDbStatus,
		BinanceStatus: binanceStatus,
		MlStatus:      mlStatus,
		RedisStatus:   redisStatus,
		Cores:         runtime.NumCPU(),
		Memory:        memStats,
		LoadAvg:       loadAvg,
		Updates:       updateMap,
		OrderBook:     orderBookMap,
	}
}
