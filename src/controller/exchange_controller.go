package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/service"
	"gitlab.com/open-soft/go-crypto-bot/src/service/exchange"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ExchangeController struct {
	SwapRepository     *repository.SwapRepository
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

	symbol := strings.TrimPrefix(req.URL.Path, "/kline/list/")

	list := e.ExchangeRepository.KLineList(symbol, true, 200)
	encoded, _ := json.Marshal(list)
	fmt.Fprintf(w, string(encoded))
}

func (e *ExchangeController) GetSwapActionListAction(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	botUuid := req.URL.Query().Get("botUuid")

	if botUuid != e.CurrentBot.BotUuid {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	actions := e.SwapRepository.GetSwapActions()
	account := e.BalanceService.GetBalance(false)
	list := make([]model.SwapContainer, 0)
	for _, action := range actions {
		balanceOne := model.Balance{
			Free:   0.00,
			Locked: 0.00,
			Asset:  action.Asset,
		}
		if balance, ok := account[balanceOne.Asset]; ok {
			balanceOne = balance
		}
		balanceTwo := model.Balance{
			Free:   0.00,
			Locked: 0.00,
			Asset:  action.Asset,
		}
		if balance, ok := account[action.GetAssetTwo()]; ok {
			balanceTwo = balance
		}
		balanceThree := model.Balance{
			Free:   0.00,
			Locked: 0.00,
			Asset:  action.Asset,
		}
		if balance, ok := account[action.GetAssetThree()]; ok {
			balanceThree = balance
		}

		list = append(list, model.SwapContainer{
			SwapAction: action,
			Balance: map[string]model.Balance{
				action.Asset:           balanceOne,
				action.GetAssetTwo():   balanceTwo,
				action.GetAssetThree(): balanceThree,
			},
		})
	}

	encoded, _ := json.Marshal(list)
	_, _ = fmt.Fprintf(w, string(encoded))
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
	_, _ = fmt.Fprintf(w, string(encoded))
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
	_, _ = fmt.Fprintf(w, string(encoded))
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

func (e *ExchangeController) GetSwapListAction(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	botUuid := req.URL.Query().Get("botUuid")

	if botUuid != e.CurrentBot.BotUuid {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	list := e.SwapRepository.GetAvailableSwapChains()
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

	fmt.Fprintf(w, encoded)
}
