package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/service"
	"gitlab.com/open-soft/go-crypto-bot/src/service/exchange"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"gitlab.com/open-soft/go-crypto-bot/src/validator"
	"net/http"
	"slices"
	"strings"
)

type OrderController struct {
	RDB                    *redis.Client
	Ctx                    *context.Context
	OrderRepository        *repository.OrderRepository
	ExchangeRepository     *repository.ExchangeRepository
	Formatter              *utils.Formatter
	PriceCalculator        *exchange.PriceCalculator
	CurrentBot             *model.Bot
	LossSecurity           *exchange.LossSecurity
	OrderExecutor          *exchange.OrderExecutor
	ProfitOptionsValidator *validator.ProfitOptionsValidator
	BotService             service.BotServiceInterface
	ProfitService          exchange.ProfitServiceInterface
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

		// todo: Decomposition is required here, move it to separate service
		binanceOrder := o.OrderRepository.GetBinanceOrder(openedOrder.Symbol, "SELL")
		executedQty := 0.00
		origQty := openedOrder.GetPositionQuantityWithSwap()

		manualOrder := o.OrderRepository.GetManualOrder(limit.Symbol)

		if binanceOrder != nil {
			sellPrice = binanceOrder.Price
			origQty = binanceOrder.OrigQty
			executedQty = binanceOrder.ExecutedQty
		} else {
			sellPrice, _ = o.PriceCalculator.CalculateSell(limit, openedOrder)
		}

		predictedPrice, err := o.ExchangeRepository.GetPredict(limit.Symbol)
		if predictedPrice > 0.00 {
			predictedPrice = o.Formatter.FormatPrice(limit, predictedPrice)
		}

		interpolation := o.PriceCalculator.InterpolatePrice(limit.Symbol)
		if interpolation.BtcInterpolationUsdt > 0.00 {
			interpolation.BtcInterpolationUsdt = o.Formatter.FormatPrice(limit, interpolation.BtcInterpolationUsdt)
		}
		if interpolation.EthInterpolationUsdt > 0.00 {
			interpolation.EthInterpolationUsdt = o.Formatter.FormatPrice(limit, interpolation.EthInterpolationUsdt)
		}

		positions = append(positions, model.Position{
			Symbol:         limit.Symbol,
			Order:          openedOrder,
			KLine:          *kLine,
			Percent:        openedOrder.GetProfitPercent(kLine.Close, o.BotService.UseSwapCapital()),
			SellPrice:      sellPrice,
			Profit:         o.Formatter.ToFixed(openedOrder.GetQuoteProfit(kLine.Close, o.BotService.UseSwapCapital()), 2),
			TargetProfit:   o.Formatter.ToFixed(openedOrder.GetQuoteProfit(sellPrice, o.BotService.UseSwapCapital()), 2),
			PredictedPrice: predictedPrice,
			Interpolation:  interpolation,
			ExecutedQty:    executedQty,
			OrigQty:        origQty,
			ManualOrderConfig: model.ManualOrderConfig{
				PriceStep:     limit.MinPrice,
				MinClosePrice: openedOrder.GetManualMinClosePrice(),
			},
			PositionTime: openedOrder.GetPositionTime(),
			CloseStrategy: model.PositionCloseStrategy{
				MinProfitPercent: o.ProfitService.GetMinProfitPercent(openedOrder),
				MinClosePrice:    o.ProfitService.GetMinClosePrice(openedOrder, openedOrder.Price),
			},
			IsPriceExpired: kLine.IsPriceExpired(),
			BinanceOrder:   binanceOrder,
			ManualOrder:    manualOrder,
		})
	}

	encoded, _ := json.Marshal(positions)
	fmt.Fprintf(w, string(encoded))
}

func (o *OrderController) UpdateExtraChargeAction(w http.ResponseWriter, req *http.Request) {
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

	if req.Method != "PUT" {
		http.Error(w, "Only PUT method is allowed", http.StatusMethodNotAllowed)

		return
	}

	var options model.UpdateOrderExtraChargeOptions

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(req.Body).Decode(&options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	entity, err := o.OrderRepository.Find(options.OrderId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)

		return
	}

	if entity.IsSell() {
		http.Error(w, "Can not update SELL order", http.StatusBadRequest)

		return
	}

	if entity.IsClosed() {
		http.Error(w, "Can not update closed order", http.StatusBadRequest)

		return
	}

	entity.ExtraChargeOptions = options.ExtraChargeOptions
	err = o.OrderRepository.Update(entity)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	entity, err = o.OrderRepository.Find(entity.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)

		return
	}

	// todo: use context with cancel: https://go.dev/doc/database/cancel-operations
	o.OrderExecutor.SetCancelRequest(entity.Symbol)
	o.ExchangeRepository.DeleteDecision(model.OrderBasedStrategyName, entity.Symbol)
	encodedRes, _ := json.Marshal(entity)
	fmt.Fprintf(w, string(encodedRes))
}

func (o *OrderController) UpdateProfitOptionsAction(w http.ResponseWriter, req *http.Request) {
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

	if req.Method != "PUT" {
		http.Error(w, "Only PUT method is allowed", http.StatusMethodNotAllowed)

		return
	}

	var options model.UpdateOrderProfitOptions

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(req.Body).Decode(&options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if len(options.ProfitOptions) == 0 {
		http.Error(w, "ProfitOptions length has to be greater than 0", http.StatusBadRequest)

		return
	}

	violation := o.ProfitOptionsValidator.Validate(options.ProfitOptions)

	if violation != nil {
		http.Error(w, violation.Error(), http.StatusBadRequest)

		return
	}

	entity, err := o.OrderRepository.Find(options.OrderId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)

		return
	}

	if entity.IsSell() {
		http.Error(w, "Can not update SELL order", http.StatusBadRequest)

		return
	}

	if entity.IsClosed() {
		http.Error(w, "Can not update closed order", http.StatusBadRequest)

		return
	}

	entity.ProfitOptions = options.ProfitOptions
	err = o.OrderRepository.Update(entity)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	entity, err = o.OrderRepository.Find(entity.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)

		return
	}

	// todo: use context with cancel: https://go.dev/doc/database/cancel-operations
	o.OrderExecutor.SetCancelRequest(entity.Symbol)
	o.ExchangeRepository.DeleteDecision(model.OrderBasedStrategyName, entity.Symbol)
	encodedRes, _ := json.Marshal(entity)
	fmt.Fprintf(w, string(encodedRes))
}

func (o *OrderController) GetPendingOrderListAction(w http.ResponseWriter, req *http.Request) {
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

	pending := make([]model.PendingOrder, 0)

	for _, limit := range o.ExchangeRepository.GetTradeLimits() {
		binanceOrder := o.OrderRepository.GetBinanceOrder(limit.Symbol, "BUY")
		if binanceOrder == nil {
			continue
		}

		kLine := o.ExchangeRepository.GetLastKLine(limit.Symbol)
		if kLine == nil {
			continue
		}

		predictedPrice, _ := o.ExchangeRepository.GetPredict(limit.Symbol)
		if predictedPrice > 0.00 {
			predictedPrice = o.Formatter.FormatPrice(limit, predictedPrice)
		}

		interpolation := o.PriceCalculator.InterpolatePrice(limit.Symbol)
		if interpolation.BtcInterpolationUsdt > 0.00 {
			interpolation.BtcInterpolationUsdt = o.Formatter.FormatPrice(limit, interpolation.BtcInterpolationUsdt)
		}
		if interpolation.EthInterpolationUsdt > 0.00 {
			interpolation.EthInterpolationUsdt = o.Formatter.FormatPrice(limit, interpolation.EthInterpolationUsdt)
		}

		pending = append(pending, model.PendingOrder{
			Symbol:         limit.Symbol,
			BinanceOrder:   *binanceOrder,
			KLine:          *kLine,
			PredictedPrice: predictedPrice,
			Interpolation:  interpolation,
			IsRisky:        o.LossSecurity.IsRiskyBuy(*binanceOrder, limit),
		})
	}

	encoded, _ := json.Marshal(pending)
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

func (o *OrderController) DeleteManualOrderAction(w http.ResponseWriter, req *http.Request) {
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

	if req.Method != "DELETE" {
		http.Error(w, "Only DELETE method is allowed", http.StatusMethodNotAllowed)

		return
	}

	symbol := strings.TrimPrefix(req.URL.Path, "/order/")
	o.OrderRepository.DeleteManualOrder(symbol)

	fmt.Fprintf(w, "OK")
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
		if opened.Swap {
			http.Error(w, "Can not sell position when SWAP is processing", http.StatusBadRequest)

			return
		}

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
	o.OrderExecutor.SetCancelRequest(tradeLimit.Symbol)

	encoded, _ := json.Marshal(manual)
	fmt.Fprintf(w, string(encoded))
}
