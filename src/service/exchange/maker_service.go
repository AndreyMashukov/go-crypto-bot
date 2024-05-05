package exchange

import (
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/service"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"log"
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
	decision, err := m.StrategyFacade.Decide(symbol)

	if err != nil {
		return
	}

	openedOrder := m.OrderRepository.GetOpenedOrderCached(symbol, "BUY")

	if openedOrder != nil && m.OrderExecutor.ProcessSwap(*openedOrder) {
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

	// allow process already opened order
	limitBuy := m.OrderRepository.GetBinanceOrder(tradeLimit.Symbol, "BUY")

	if !m.TradeStack.CanBuy(tradeLimit) && limitBuy == nil {
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

	if len(marketDepth.Bids) < 3 && manualOrder == nil {
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

	// allow process already opened order
	limitBuy := m.OrderRepository.GetBinanceOrder(tradeLimit.Symbol, "BUY")

	if !m.TradeStack.CanBuy(tradeLimit) && limitBuy == nil {
		return
	}

	lastKline := m.ExchangeRepository.GetCurrentKline(tradeLimit.Symbol)

	if lastKline == nil {
		log.Printf("[%s] Last price is unknown... skip!", tradeLimit.Symbol)

		return
	}

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

	if len(marketDepth.Bids) < 3 && manualOrder == nil {
		log.Printf("[%s] Too small BIDs amount: %d\n", tradeLimit.Symbol, len(marketDepth.Bids))
		return
	}

	priceModel := m.PriceCalculator.CalculateBuy(tradeLimit)
	if priceModel.Error != nil {
		log.Printf("[%s] Price error: %s", tradeLimit.Symbol, priceModel.Error.Error())

		return
	}

	price := priceModel.Price

	profit := openedOrder.GetProfitPercent(lastKline.Close, m.BotService.UseSwapCapital())
	extraChargePercent := tradeLimit.GetBuyOnFallPercent(openedOrder, *lastKline, m.BotService.UseSwapCapital())

	if profit.Lte(extraChargePercent) {
		// extra buy on current price
		if price < lastKline.Close {
			price = m.Formatter.FormatPrice(tradeLimit, lastKline.Close)
		}

		err := m.OrderExecutor.BuyExtra(tradeLimit, openedOrder, price)
		if err != nil {
			log.Printf("[%s] %s", tradeLimit.Symbol, err)

			if m.BotService.IsSwapEnabled() {
				m.OrderExecutor.TrySwap(openedOrder)
			}
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

	if lastKline == nil {
		log.Printf("[%s] Last price is unknown... skip!", tradeLimit.Symbol)

		return
	}

	// allow process already opened order
	limitSell := m.OrderRepository.GetBinanceOrder(tradeLimit.Symbol, "SELL")

	if !m.TradeFilterService.CanSell(tradeLimit) && limitSell == nil {
		log.Printf("[%s] Can't sell, trade filter conditions is not matched", tradeLimit.Symbol)

		return
	}

	manualOrder := m.OrderRepository.GetManualOrder(tradeLimit.Symbol)
	marketDepth := m.PriceCalculator.GetDepth(tradeLimit.Symbol, 20)

	if len(marketDepth.Asks) < 3 && manualOrder == nil {
		log.Printf("[%s] Too small ASKs amount: %d\n", tradeLimit.Symbol, len(marketDepth.Asks))
		return
	}

	if lastKline == nil {
		log.Printf("[%s] No information about current price", tradeLimit.Symbol)
		return
	}

	price, priceErr := m.PriceCalculator.CalculateSell(tradeLimit, openedOrder)

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

func (m *MakerService) UpdateSwapPairs() {
	swapMap := make(map[string][]model.ExchangeSymbol)
	exchangeInfo, _ := m.Binance.GetExchangeData(make([]string, 0))
	tradeLimits := m.ExchangeRepository.GetTradeLimits()

	supportedQuoteAssets := []string{"BTC", "ETH", "BNB", "TRX", "XRP", "EUR", "DAI", "TUSD", "USDC", "AUD", "TRY", "BRL"}

	for _, tradeLimit := range tradeLimits {
		if !tradeLimit.IsEnabled {
			continue
		}

		swapMap[tradeLimit.Symbol] = make([]model.ExchangeSymbol, 0)

		for _, exchangeSymbol := range exchangeInfo.Symbols {
			if !exchangeSymbol.IsTrading() {
				continue
			}

			if exchangeSymbol.Symbol == tradeLimit.Symbol {
				baseAsset := exchangeSymbol.BaseAsset
				quoteAsset := exchangeSymbol.QuoteAsset

				for _, exchangeItem := range exchangeInfo.Symbols {
					if !exchangeItem.IsTrading() {
						continue
					}

					if !slices.Contains(supportedQuoteAssets, exchangeItem.QuoteAsset) {
						continue
					}

					if exchangeItem.BaseAsset == baseAsset && exchangeItem.QuoteAsset != quoteAsset {
						swapMap[tradeLimit.Symbol] = append(swapMap[tradeLimit.Symbol], exchangeItem)
					}
				}
			}
		}

		for _, exchangeItem := range swapMap[tradeLimit.Symbol] {
			swapPair, err := m.ExchangeRepository.GetSwapPair(exchangeItem.Symbol)
			if err != nil {
				swapPair := model.SwapPair{
					SourceSymbol:   tradeLimit.Symbol,
					Symbol:         exchangeItem.Symbol,
					BaseAsset:      exchangeItem.BaseAsset,
					QuoteAsset:     exchangeItem.QuoteAsset,
					BuyPrice:       0.00,
					SellPrice:      0.00,
					PriceTimestamp: 0,
				}

				for _, filter := range exchangeItem.Filters {
					if filter.FilterType == "PRICE_FILTER" {
						swapPair.MinPrice = *filter.MinPrice
					}
					if filter.FilterType == "LOT_SIZE" {
						swapPair.MinQuantity = *filter.MinQuantity
					}
					if filter.FilterType == "NOTIONAL" {
						swapPair.MinNotional = *filter.MinNotional
					}
				}

				_, _ = m.ExchangeRepository.CreateSwapPair(swapPair)
			} else {
				for _, filter := range exchangeItem.Filters {
					if filter.FilterType == "PRICE_FILTER" {
						swapPair.MinPrice = *filter.MinPrice
					}
					if filter.FilterType == "LOT_SIZE" {
						swapPair.MinQuantity = *filter.MinQuantity
					}
					if filter.FilterType == "NOTIONAL" {
						swapPair.MinNotional = *filter.MinNotional
					}
				}

				_ = m.ExchangeRepository.UpdateSwapPair(swapPair)
			}
		}
	}
}

func (m *MakerService) UpdateLimits() {
	symbols := make([]string, 0)

	tradeLimits := m.ExchangeRepository.GetTradeLimits()
	limitMap := make(map[string]model.TradeLimit)
	for _, tradeLimit := range tradeLimits {
		symbols = append(symbols, tradeLimit.Symbol)
		limitMap[tradeLimit.Symbol] = tradeLimit
	}

	exchangeInfo, err := m.Binance.GetExchangeData(symbols)

	if err != nil {
		log.Printf("Exchange Limits: %s", err.Error())
		return
	}

	for _, exchangeSymbol := range exchangeInfo.Symbols {
		tradeLimit := limitMap[exchangeSymbol.Symbol]
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
			"[%s] Trade Limit Updated, MIN_LOT = %f, MIN_PRICE = %f",
			tradeLimit.Symbol,
			tradeLimit.MinQuantity,
			tradeLimit.MinPrice,
		)
	}
}

func (m *MakerService) StartTrade() {
	go func() {
		for {
			m.UpdateLimits()
			time.Sleep(time.Minute * 5)
		}
	}()

	for _, tradeLimit := range m.ExchangeRepository.GetTradeLimits() {
		go func(symbol string) {
			for {
				m.Make(symbol)
				time.Sleep(time.Millisecond * 500)
			}
		}(tradeLimit.Symbol)
	}
}

func (m *MakerService) RecoverOrders() {
	tradeLimits := m.ExchangeRepository.GetTradeLimits()
	symbols := make([]string, 0)
	for _, limit := range tradeLimits {
		symbols = append(symbols, limit.Symbol)

		sellBinance := m.OrderRepository.GetBinanceOrder(limit.Symbol, "SELL")
		if sellBinance != nil {
			_, err := m.ExchangeApi.QueryOrder(sellBinance.Symbol, sellBinance.OrderId)
			// todo: Delete already registered???
			if err != nil {
				m.OrderRepository.DeleteBinanceOrder(*sellBinance)
			}
		}

		buyBinance := m.OrderRepository.GetBinanceOrder(limit.Symbol, "BUY")
		if buyBinance != nil {
			_, err := m.ExchangeApi.QueryOrder(buyBinance.Symbol, buyBinance.OrderId)
			// todo: Delete already registered???
			if err != nil {
				m.OrderRepository.DeleteBinanceOrder(*buyBinance)
			}
		}
	}

	binanceOrders, err := m.Binance.GetOpenedOrders()
	if err == nil {
		for _, binanceOrder := range binanceOrders {
			if !slices.Contains(symbols, binanceOrder.Symbol) {
				log.Printf("[%s] binance order %d skipped", binanceOrder.Symbol, binanceOrder.OrderId)

				continue
			}

			log.Printf("[%s] loaded binance order %d", binanceOrder.Symbol, binanceOrder.OrderId)
			m.OrderRepository.SetBinanceOrder(binanceOrder)
		}
	}

	// Wait 5 seconds, here API can update some settings...
	time.Sleep(time.Second * 5)
}
