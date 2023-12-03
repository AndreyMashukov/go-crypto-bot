package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/service"
	"net/http"
	"strings"
	"time"
)

type ExchangeController struct {
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	ChartService       *service.ChartService
	RDB                *redis.Client
	Ctx                *context.Context
	CurrentBot         *model.Bot
}

func (e *ExchangeController) GetKlineListAction(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	botUuid := req.URL.Query().Get("botUuid")

	if botUuid != e.CurrentBot.BotUuid {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	symbol := strings.TrimPrefix(req.URL.Path, "/kline/list/")

	list := e.ExchangeRepository.KLineList(symbol, true, 200)
	encoded, _ := json.Marshal(list)
	fmt.Fprintf(w, string(encoded))
}

func (e *ExchangeController) GetDepthAction(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	botUuid := req.URL.Query().Get("botUuid")

	if botUuid != e.CurrentBot.BotUuid {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	symbol := strings.TrimPrefix(req.URL.Path, "/depth/")

	list := e.ExchangeRepository.GetDepth(symbol)
	encoded, _ := json.Marshal(list)
	fmt.Fprintf(w, string(encoded))
}

func (e *ExchangeController) GetTradeListAction(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	botUuid := req.URL.Query().Get("botUuid")

	if botUuid != e.CurrentBot.BotUuid {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	symbol := strings.TrimPrefix(req.URL.Path, "/trade/list/")

	list := e.ExchangeRepository.TradeList(symbol)
	encoded, _ := json.Marshal(list)
	fmt.Fprintf(w, string(encoded))
}

func (e *ExchangeController) GetChartListAction(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	botUuid := req.URL.Query().Get("botUuid")

	if botUuid != e.CurrentBot.BotUuid {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	encoded := e.RDB.Get(*e.Ctx, fmt.Sprintf("chart-cache-bot-%d", e.CurrentBot.Id)).Val()

	if len(encoded) == 0 {
		chart := e.ChartService.GetCharts()
		encodedRes, _ := json.Marshal(chart)
		encoded = string(encodedRes)
		e.RDB.Set(*e.Ctx, fmt.Sprintf("chart-cache-bot-%d", e.CurrentBot.Id), encoded, time.Second*5)
	}

	fmt.Fprintf(w, encoded)
}
