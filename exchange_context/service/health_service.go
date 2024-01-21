package service

import (
	"context"
	"database/sql"
	"github.com/rafacas/sysstats"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"runtime"
	"time"
)

type HealthService struct {
	ExchangeRepository *repository.ExchangeRepository
	PythonMLBridge     *PythonMLBridge
	DB                 *sql.DB
	RDB                *redis.Client
	Ctx                *context.Context
	Binance            *client.Binance
	CurrentBot         *model.Bot
}

func (h *HealthService) HealthCheck() model.BotHealth {
	updateMap := make(map[string]string)

	for _, limit := range h.ExchangeRepository.GetTradeLimits() {
		if !limit.IsEnabled {
			continue
		}

		kLine := h.ExchangeRepository.GetLastKLine(limit.Symbol)
		dateString := ""
		if kLine != nil {
			dateString = time.Unix(kLine.UpdatedAt, 0).Format("2006-01-02 15:04:05")
		}

		updateMap[limit.Symbol] = dateString
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
	if binanceStatus == model.BinanceStatusOk && !h.Binance.APIKeyCheckCompleted {
		binanceStatus = model.BinanceStatusApiKeyCheck
	}

	dbStatus := model.DbStatusOk
	if h.DB.Ping() != nil {
		dbStatus = model.DbStatusFail
	}
	redisStatus := model.RedisStatusOk
	if h.RDB.Ping(*h.Ctx).Err() != nil {
		redisStatus = model.RedisStatusFail
	}
	mlStatus := model.MlStatusReady
	if h.PythonMLBridge.Learning {
		mlStatus = model.MlStatusLearning
	}

	return model.BotHealth{
		Bot:           *h.CurrentBot,
		DbStatus:      dbStatus,
		BinanceStatus: binanceStatus,
		MlStatus:      mlStatus,
		RedisStatus:   redisStatus,
		Cores:         runtime.NumCPU(),
		Memory:        memStats,
		LoadAvg:       loadAvg,
		Updates:       updateMap,
	}
}
