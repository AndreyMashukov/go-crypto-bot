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
	"slices"
	"time"
)

type OrderController struct {
	RDB                *redis.Client
	Ctx                *context.Context
	OrderRepository    *ExchangeRepository.OrderRepository
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	Formatter          *service.Formatter
	PriceCalculator    *service.PriceCalculator
	CurrentBot         *model.Bot
}

func (o *OrderController) GetOrderTradeListAction(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if req.Method == "OPTIONS" {
		fmt.Fprintf(w, "OK")
		return
	}

	if req.Method != "GET" {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)

		return
	}

	botUuid := req.URL.Query().Get("botUuid")

	if botUuid != o.CurrentBot.BotUuid {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	list := o.OrderRepository.GetTrades()
	encoded, _ := json.Marshal(list)
	fmt.Fprintf(w, string(encoded))
}

func (o *OrderController) GetPositionListAction(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if req.Method == "OPTIONS" {
		fmt.Fprintf(w, "OK")
		return
	}

	botUuid := req.URL.Query().Get("botUuid")

	if botUuid != o.CurrentBot.BotUuid {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	if req.Method != "GET" {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)

		return
	}

	positions := make([]model.Position, 0)

	for _, limit := range o.ExchangeRepository.GetTradeLimits() {
		openedOrder, err := o.OrderRepository.GetOpenedOrderCached(limit.Symbol, "BUY")
		if err != nil {
			continue
		}

		kLine := o.ExchangeRepository.GetLastKLine(limit.Symbol)
		if kLine == nil {
			continue
		}

		var sellPrice float64

		binanceOrder := o.OrderRepository.GetBinanceOrder(openedOrder.Symbol, "SELL")
		if binanceOrder != nil {
			sellPrice = binanceOrder.Price
		} else {
			sellPriceCacheKey := fmt.Sprintf("sell-price-%d", openedOrder.Id)
			sellPriceCached := o.RDB.Get(*o.Ctx, sellPriceCacheKey).Val()
			if len(sellPriceCached) > 0 {
				_ = json.Unmarshal([]byte(sellPriceCached), &sellPrice)
			} else {
				sellPrice = o.PriceCalculator.CalculateSell(limit, openedOrder)
				encoded, _ := json.Marshal(sellPrice)
				o.RDB.Set(*o.Ctx, sellPriceCacheKey, string(encoded), time.Hour)
			}
		}

		predictedPrice, err := o.ExchangeRepository.GetPredict(limit.Symbol)
		if predictedPrice > 0.00 {
			predictedPrice = o.Formatter.FormatPrice(limit, predictedPrice)
		}

		positions = append(positions, model.Position{
			Symbol:         limit.Symbol,
			Order:          openedOrder,
			KLine:          *kLine,
			Percent:        openedOrder.GetProfitPercent(kLine.Close),
			SellPrice:      sellPrice,
			Profit:         o.Formatter.ToFixed(openedOrder.GetQuoteProfit(kLine.Close), 2),
			TargetProfit:   o.Formatter.ToFixed(openedOrder.GetQuoteProfit(sellPrice), 2),
			PredictedPrice: predictedPrice,
		})
	}

	encoded, _ := json.Marshal(positions)
	fmt.Fprintf(w, string(encoded))
}

func (o *OrderController) GetOrderListAction(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if req.Method == "OPTIONS" {
		fmt.Fprintf(w, "OK")
		return
	}

	botUuid := req.URL.Query().Get("botUuid")

	if botUuid != o.CurrentBot.BotUuid {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	list := o.OrderRepository.GetList()
	encoded, _ := json.Marshal(list)
	fmt.Fprintf(w, string(encoded))
}

func (o *OrderController) PostManualOrderAction(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if req.Method == "OPTIONS" {
		fmt.Fprintf(w, "OK")
		return
	}

	botUuid := req.URL.Query().Get("botUuid")

	if botUuid != o.CurrentBot.BotUuid {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	if req.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)

		return
	}

	var manual model.ManualOrder

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(req.Body).Decode(&manual)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if manual.BotUuid != o.CurrentBot.BotUuid {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	allowedOperations := []string{"BUY", "SELL"}
	if !slices.Contains(allowedOperations, manual.Operation) {
		http.Error(w, "Only BUY/SELL operations are supported", http.StatusBadRequest)

		return
	}

	tradeLimit, err := o.ExchangeRepository.GetTradeLimit(manual.Symbol)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s не поддерживается", manual.Symbol), http.StatusBadRequest)

		return
	}

	opened, err := o.OrderRepository.GetOpenedOrderCached(manual.Symbol, "BUY")
	if err == nil && manual.Operation == "SELL" {
		minPrice := o.Formatter.FormatPrice(tradeLimit, opened.GetManualMinClosePrice())
		if minPrice > manual.Price {
			http.Error(w, fmt.Sprintf("Price can not be less then %.6f", minPrice), http.StatusBadRequest)

			return
		}
	}

	if err != nil && manual.Operation == "SELL" {
		http.Error(w, "There are no opened orders", http.StatusBadRequest)

		return
	}

	if err == nil && manual.Operation == "BUY" {
		http.Error(w, "Manual extra buy is temporary prohibited", http.StatusBadRequest)

		return
	}

	minPrice, buyError := o.PriceCalculator.CalculateBuy(tradeLimit)

	if buyError != nil {
		http.Error(w, fmt.Sprintf("Ошибка: %s", buyError.Error()), http.StatusBadRequest)

		return
	}

	if err != nil && manual.Operation == "BUY" && minPrice < manual.Price {
		http.Error(w, fmt.Sprintf("Price can not be greather then %f", minPrice), http.StatusBadRequest)

		return
	}

	binanceOrder := o.OrderRepository.GetBinanceOrder(manual.Symbol, manual.Operation)

	if binanceOrder != nil && binanceOrder.Status == "PARTIALLY_FILLED" {
		http.Error(w, "Order is filling now, please wait until has been filled", http.StatusBadRequest)

		return
	}

	manual.Price = o.Formatter.FormatPrice(tradeLimit, manual.Price)
	o.OrderRepository.SetManualOrder(manual)

	encoded, _ := json.Marshal(manual)
	fmt.Fprintf(w, string(encoded))
}
