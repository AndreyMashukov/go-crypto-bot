package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/client"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/model"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/repository"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/service"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/service/exchange"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ExchangeController struct {
	ExchangeRepository *repository.ExchangeRepository
	ChartService       *service.ChartService
	RDB                *redis.Client
	Ctx                *context.Context
	CurrentBot         *model.Bot
	BotService         *service.BotService
	BalanceService     *exchange.BalanceService
	Exchange           client.ExchangeAPIInterface
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

	_ = strings.TrimPrefix(req.URL.Path, "/kline/list/")

	// Phase E removed the Redis kline cache. The Prometheus-backed
	// chart endpoint that replaces this one is added in Phase F; until
	// then the action returns an empty list so the FE keeps parsing.
	encoded, _ := json.Marshal([]struct{}{})
	_, _ = fmt.Fprint(w, string(encoded))
}

func (e *ExchangeController) GetExchangeOrderAction(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	botUuid := req.URL.Query().Get("botUuid")

	if botUuid != e.CurrentBot.BotUuid {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	symbol := strings.TrimSpace(strings.TrimPrefix(req.URL.Path, "/exchange/order/"))
	if "" == symbol {
		http.Error(w, "Symbol should not be empty", http.StatusBadRequest)

		return
	}

	orderId := strings.TrimSpace(req.URL.Query().Get("orderId"))
	order, err := e.Exchange.QueryOrder(symbol, orderId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	encoded, _ := json.Marshal(order)
	_, _ = fmt.Fprint(w, string(encoded))
}

func (e *ExchangeController) GetAccountAction(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	botUuid := req.URL.Query().Get("botUuid")

	if botUuid != e.CurrentBot.BotUuid {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	hideZero, err := strconv.ParseBool(req.URL.Query().Get("hideZero"))
	if err != nil {
		hideZero = false
	}

	account := e.BalanceService.GetBalance(hideZero)

	encoded, _ := json.Marshal(account)
	_, _ = fmt.Fprint(w, string(encoded))
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

	list := e.ExchangeRepository.GetDepth(symbol, 20)
	encoded, _ := json.Marshal(list)
	_, _ = fmt.Fprint(w, string(encoded))
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
	_, _ = fmt.Fprint(w, string(encoded))
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

	symbol := req.URL.Query().Get("symbol")

	symbolFilter := make([]string, 0)

	if len(symbol) > 0 {
		symbolFilter = append(symbolFilter, symbol)
	}

	encoded := e.RDB.Get(*e.Ctx, fmt.Sprintf("chart-cache-bot-%d", e.CurrentBot.Id)).Val()

	if len(encoded) == 0 {
		chart := e.ChartService.GetCharts(symbolFilter)
		encodedRes, _ := json.Marshal(chart)
		encoded = string(encodedRes)
		e.RDB.Set(*e.Ctx, fmt.Sprintf("chart-cache-bot-%d", e.CurrentBot.Id), encoded, time.Second*5)
	}

	_, _ = fmt.Fprint(w, encoded)
}
