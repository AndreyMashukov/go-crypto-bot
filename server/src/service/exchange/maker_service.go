package exchange

import (
	"context"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/client"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/model"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/repository"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/service"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/utils"
	"log"
	"runtime"
	"slices"
	"strings"
	"time"
)

type MakerService struct {
	TradeFilterService TradeFilterServiceInterface
	ExchangeApi        client.ExchangeOrderAPIInterface
	OrderRepository    repository.OrderStorageInterface
	ExchangeRepository repository.BaseTradeStorageInterface
	BotService         service.BotServiceInterface
	StrategyFacade     StrategyFacadeInterface
	PriceCalculator    PriceCalculatorInterface
	TradeStack         BuyOrderStackInterface
	OrderExecutor      OrderExecutorInterface
	Binance            client.ExchangePriceAPIInterface
	Formatter          *utils.Formatter
	CurrentBot         *model.Bot
	HoldScore          float64
}

func (m *MakerService) Make(symbol string) {
	openedOrder := m.OrderRepository.GetOpenedOrderCached(symbol, "BUY")

	decision, err := m.StrategyFacade.Decide(symbol)

	if err != nil {
		return
	}

	if decision.Hold >= m.HoldScore {
		return
	}

	tradeLimit, err := m.ExchangeRepository.GetTradeLimit(symbol)

	if err != nil {
		log.Println(err.Error())

		return
	}

	if decision.Sell > decision.Buy {
		if openedOrder != nil {
			m.ProcessSell(tradeLimit, *openedOrder)
		}

		return
	}

	if decision.Buy > decision.Sell {
		if openedOrder == nil {
			m.ProcessBuy(tradeLimit)
		} else {
			m.ProcessExtraBuy(tradeLimit, *openedOrder)
		}
	}
}

func (m *MakerService) ProcessBuy(tradeLimit model.TradeLimit) {
	if !tradeLimit.IsEnabled {
		return
	}

	limitBuy := m.OrderRepository.GetBinanceOrder(tradeLimit.Symbol, "BUY")

	if limitBuy != nil {
		// todo: signal := m.SignalStorage.GetSignal(tradeLimit.Symbol)
		priceModel := m.PriceCalculator.CalculateBuy(tradeLimit)

		err := m.OrderExecutor.Buy(tradeLimit, limitBuy.Price, limitBuy.OrigQty, priceModel.Signal)
		if err != nil {
			log.Printf(
				"[%s] Existing order [%s] BUY Error: %s",
				tradeLimit.Symbol,
				limitBuy.OrderId,
				err,
			)

			if strings.Contains(err.Error(), "not enough balance") {
				log.Printf("[%s] wait 1 minute...", tradeLimit.Symbol)
				time.Sleep(time.Minute * 1)
			}
		}
		return
	}

	if !m.TradeStack.CanBuy(tradeLimit) {
		return
	}

	lastKline := m.ExchangeRepository.GetCurrentKline(tradeLimit.Symbol)

	if lastKline == nil {
		log.Printf("[%s] Last price is unknown... skip!", tradeLimit.Symbol)

		return
	}

	balanceErr := m.OrderExecutor.CheckMinBalance(tradeLimit, *lastKline)

	if balanceErr != nil {
		log.Printf("[%s] Min balance check: %s", tradeLimit.Symbol, balanceErr.Error())
		time.Sleep(time.Minute)
		return
	}

	marketDepth := m.PriceCalculator.GetDepth(tradeLimit.Symbol, 20)
	manualOrder := m.OrderRepository.GetManualOrder(tradeLimit.Symbol)

	if len(marketDepth.Bids) == 0 && manualOrder == nil {
		log.Printf("[%s] Too small BIDs amount: %d\n", tradeLimit.Symbol, len(marketDepth.Bids))
		return
	}

	priceModel := m.PriceCalculator.CalculateBuy(tradeLimit)

	if priceModel.Error != nil {
		log.Printf("[%s] Price error: %s", tradeLimit.Symbol, priceModel.Error.Error())

		return
	}

	price := priceModel.Price

	if manualOrder != nil && manualOrder.IsBuy() {
		price = m.Formatter.FormatPrice(tradeLimit, manualOrder.Price)
	}

	// todo: exclude existing exchange order...
	if lastKline.IsPriceExpired() {
		log.Printf("[%s] Price is expired", tradeLimit.Symbol)
		return
	}

	if price > 0 {
		// todo: get buy quantity, buy to all cutlet! check available balance!
		quantity := m.Formatter.FormatQuantity(tradeLimit, tradeLimit.USDTLimit/price)

		if (quantity * price) < tradeLimit.MinNotional {
			log.Printf("[%s] BUY Notional: %.8f < %.8f", tradeLimit.Symbol, quantity*price, tradeLimit.MinNotional)
			return
		}

		err := m.OrderExecutor.Buy(tradeLimit, price, quantity, priceModel.Signal)
		if err != nil {
			log.Printf("[%s] %s", tradeLimit.Symbol, err)

			if strings.Contains(err.Error(), "not enough balance") {
				log.Printf("[%s] wait 1 minute...", tradeLimit.Symbol)
				time.Sleep(time.Minute * 1)
			}
		}
	} else {
		log.Printf("[%s] No ASKs on the market", tradeLimit.Symbol)
	}
}

func (m *MakerService) ProcessExtraBuy(tradeLimit model.TradeLimit, openedOrder model.Order) {
	if !tradeLimit.IsEnabled {
		log.Printf("[%s] BUY operation is disabled", tradeLimit.Symbol)
		return
	}

	limitBuy := m.OrderRepository.GetBinanceOrder(tradeLimit.Symbol, "BUY")

	if limitBuy != nil {
		err := m.OrderExecutor.BuyExtra(tradeLimit, openedOrder, limitBuy.Price)
		if err != nil {
			log.Printf(
				"[%s] Existing order [%s] Extra BUY Error: %s",
				tradeLimit.Symbol,
				limitBuy.OrderId,
				err,
			)
		}
		return
	}

	if !m.TradeStack.CanBuy(tradeLimit) {
		return
	}

	lastKline := m.ExchangeRepository.GetCurrentKline(tradeLimit.Symbol)

	if lastKline == nil {
		log.Printf("[%s] Last price is unknown... skip!", tradeLimit.Symbol)

		return
	}

	// todo: exclude existing exchange order...
	if lastKline.IsPriceExpired() {
		log.Printf("[%s] Price is expired", tradeLimit.Symbol)
		return
	}

	balanceErr := m.OrderExecutor.CheckMinBalance(tradeLimit, *lastKline)

	if balanceErr != nil {
		log.Printf("[%s] Min balance check: %s", tradeLimit.Symbol, balanceErr.Error())
		time.Sleep(time.Minute)
		return
	}

	marketDepth := m.PriceCalculator.GetDepth(tradeLimit.Symbol, 20)
	manualOrder := m.OrderRepository.GetManualOrder(tradeLimit.Symbol)

	if len(marketDepth.Bids) == 0 && manualOrder == nil {
		log.Printf("[%s] Too small BIDs amount: %d\n", tradeLimit.Symbol, len(marketDepth.Bids))
		return
	}

	priceModel := m.PriceCalculator.CalculateBuy(tradeLimit)
	if priceModel.Error != nil {
		log.Printf("[%s] Price error: %s", tradeLimit.Symbol, priceModel.Error.Error())

		return
	}

	price := priceModel.Price

	profit := openedOrder.GetProfitPercent(lastKline.Close.Value(), m.BotService.UseSwapCapital())
	extraChargePercent := tradeLimit.GetBuyOnFallPercent(openedOrder, *lastKline, m.BotService.UseSwapCapital())

	if profit.Lte(extraChargePercent) {
		if price < lastKline.Close.Value() {
			price = m.Formatter.FormatPrice(tradeLimit, lastKline.Close.Value())
		}

		err := m.OrderExecutor.BuyExtra(tradeLimit, openedOrder, price)
		if err != nil {
			log.Printf("[%s] %s", tradeLimit.Symbol, err)
		}
	} else {
		log.Printf(
			"[%s] Extra charge is not allowed: %.2f of %.2f",
			tradeLimit.Symbol,
			profit.Value(),
			extraChargePercent.Value(),
		)
	}
}

func (m *MakerService) ProcessSell(tradeLimit model.TradeLimit, openedOrder model.Order) {
	lastKline := m.ExchangeRepository.GetCurrentKline(tradeLimit.Symbol)

	// todo: exclude existing exchange order...
	if lastKline == nil {
		log.Printf("[%s] Last price is unknown... skip!", tradeLimit.Symbol)

		return
	}

	limitSell := m.OrderRepository.GetBinanceOrder(tradeLimit.Symbol, "SELL")

	if limitSell != nil {
		err := m.OrderExecutor.Sell(
			tradeLimit,
			openedOrder,
			limitSell.Price,
			limitSell.OrigQty,
			false,
		)

		if err != nil {
			log.Printf(
				"[%s] Existing order [%s] SELL error: %s",
				openedOrder.Symbol,
				limitSell.OrderId,
				err.Error(),
			)
		}
		return
	}

	if !m.TradeFilterService.CanSell(tradeLimit) {
		log.Printf("[%s] Can't sell, trade filter conditions is not matched", tradeLimit.Symbol)

		return
	}

	manualOrder := m.OrderRepository.GetManualOrder(tradeLimit.Symbol)
	marketDepth := m.PriceCalculator.GetDepth(tradeLimit.Symbol, 20)

	if len(marketDepth.Asks) == 0 && manualOrder == nil {
		log.Printf("[%s] Too small ASKs amount: %d\n", tradeLimit.Symbol, len(marketDepth.Asks))
		return
	}

	// todo: exclude existing exchange order...
	if lastKline == nil {
		log.Printf("[%s] No information about current price", tradeLimit.Symbol)
		return
	}

	price, priceErr := m.PriceCalculator.CalculateSell(tradeLimit, openedOrder)

	// todo: exclude existing exchange order...
	if priceErr != nil {
		log.Printf("[%s] Price error: %s", tradeLimit.Symbol, priceErr.Error())

		return
	}

	isManual := false

	if manualOrder != nil && manualOrder.IsSell() {
		price = m.Formatter.FormatPrice(tradeLimit, manualOrder.Price)
		isManual = true
	}

	if price > 0 {
		quantity := m.Formatter.FormatQuantity(tradeLimit, m.OrderExecutor.CalculateSellQuantity(openedOrder))

		if quantity >= tradeLimit.MinQuantity {
			log.Printf("[%s] SELL QTY = %f", openedOrder.Symbol, quantity)
			err := m.OrderExecutor.Sell(tradeLimit, openedOrder, price, quantity, isManual)
			if err != nil {
				log.Printf("[%s] SELL error: %s", openedOrder.Symbol, err.Error())
			}
		} else {
			log.Printf("[%s] SELL QTY = %f is too small!", openedOrder.Symbol, quantity)
		}
	}
}

func (m *MakerService) tradeLimit(symbol string) *model.TradeLimit {
	tradeLimits := m.ExchangeRepository.GetTradeLimits()
	for _, tradeLimit := range tradeLimits {
		if tradeLimit.Symbol == symbol {
			return &tradeLimit
		}
	}

	return nil
}

func (m *MakerService) UpdateLimits() {
	tradeLimits := m.ExchangeRepository.GetTradeLimits()
	symbolMap := make(map[string]model.TradeLimit)
	for _, tradeLimit := range tradeLimits {
		symbolMap[tradeLimit.Symbol] = tradeLimit
	}

	exchangeInfo, err := m.Binance.GetExchangeData([]string{})

	if err != nil {
		log.Printf("Exchange Limits: %s", err.Error())
		return
	}

	for _, exchangeSymbol := range exchangeInfo.Symbols {
		tradeLimit, ok := symbolMap[exchangeSymbol.Symbol]
		if !ok {
			continue
		}

		for _, filter := range exchangeSymbol.Filters {
			if filter.FilterType == "PRICE_FILTER" {
				tradeLimit.MinPrice = *filter.MinPrice
			}
			if filter.FilterType == "LOT_SIZE" {
				tradeLimit.MinQuantity = *filter.MinQuantity
			}
			if filter.FilterType == "NOTIONAL" {
				tradeLimit.MinNotional = *filter.MinNotional
			}
		}
		err := m.ExchangeRepository.UpdateTradeLimit(tradeLimit)
		if err != nil {
			log.Printf("[%s] Trade Limit Update: %s", tradeLimit.Symbol, err.Error())
			continue
		}

		log.Printf(
			"[%s] Trade Limit Updated, MIN_LOT = %.10f, MIN_PRICE = %.10f",
			tradeLimit.Symbol,
			tradeLimit.MinQuantity,
			tradeLimit.MinPrice,
		)
	}
}

// StartTrade spawns the per-symbol trading loops and the periodic
// trade-limit refresher. Each goroutine observes ctx; on cancel they
// all exit at their next tick boundary.
func (m *MakerService) StartTrade(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Minute * 5)
		defer ticker.Stop()
		m.UpdateLimits()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.UpdateLimits()
			}
		}
	}()

	for _, tradeLimit := range m.ExchangeRepository.GetTradeLimits() {
		go func(symbol string) {
			ticker := time.NewTicker(time.Millisecond * 250)
			defer ticker.Stop()
			for {
				m.Make(symbol)

				runtime.GC()
				runtime.Gosched()
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
				}
			}
		}(tradeLimit.Symbol)
	}
}

func (m *MakerService) RecoverOrders() {
	tradeLimits := m.ExchangeRepository.GetTradeLimits()
	symbols := make([]string, 0)
	for _, limit := range tradeLimits {
		symbols = append(symbols, limit.Symbol)
	}

	binanceOrders, err := m.Binance.GetOpenedOrders()
	if err == nil {
		for _, binanceOrder := range binanceOrders {
			if binanceOrder.IsCanceled() || binanceOrder.IsExpired() {
				continue
			}

			if !slices.Contains(symbols, binanceOrder.Symbol) {
				log.Printf("[%s] %s order %s skipped", binanceOrder.Symbol, m.CurrentBot.Exchange, binanceOrder.OrderId)

				continue
			}

			log.Printf("[%s] loaded %s order %s, status = %s", binanceOrder.Symbol, m.CurrentBot.Exchange, binanceOrder.OrderId, binanceOrder.Status)
			m.OrderRepository.SetBinanceOrder(binanceOrder)
		}
	}

	time.Sleep(time.Second * 5)
}
