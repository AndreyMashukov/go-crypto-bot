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
	OrderExecutor      *OrderExecutor
	OrderRepository    *repository.OrderRepository
	ExchangeRepository *repository.ExchangeRepository
	Binance            *client.Binance
	Formatter          *utils.Formatter
	HoldScore          float64
	CurrentBot         *model.Bot
	PriceCalculator    *PriceCalculator
	TradeStack         *TradeStack
	BotService         service.BotServiceInterface
	StrategyFacade     *StrategyFacade
}

func (m *MakerService) Make(symbol string) {
	decision, err := m.StrategyFacade.Decide(symbol)

	if err != nil {
		log.Println(err.Error())

		return
	}

	openedOrder, buyOrderErr := m.OrderRepository.GetOpenedOrderCached(symbol, "BUY")

	if buyOrderErr == nil && m.OrderExecutor.ProcessSwap(openedOrder) {
		return
	}

	// todo: fallback to existing order...
	if decision.Hold >= m.HoldScore {
		return
	}

	tradeLimit, err := m.ExchangeRepository.GetTradeLimit(symbol)

	if err != nil {
		log.Println(err.Error())

		return
	}

	if decision.Sell > decision.Buy {
		if buyOrderErr == nil {
			m.ProcessSell(tradeLimit, openedOrder)
		}

		return
	}

	if decision.Buy > decision.Sell {
		if buyOrderErr != nil {
			m.ProcessBuy(tradeLimit)
		} else {
			m.ProcessExtraBuy(tradeLimit, openedOrder)
		}
	}
}

func (m *MakerService) ProcessBuy(tradeLimit model.TradeLimit) {
	if !tradeLimit.IsEnabled {
		log.Printf("[%s] BUY operation is disabled", tradeLimit.Symbol)
		return
	}

	if !m.TradeStack.CanBuy(tradeLimit) {
		log.Printf("[%s] Trade Stack check is not passed, wait order.", tradeLimit.Symbol)
		return
	}

	lastKline := m.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)

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

	marketDepth := m.PriceCalculator.GetDepth(tradeLimit.Symbol)
	manualOrder := m.OrderRepository.GetManualOrder(tradeLimit.Symbol)

	if len(marketDepth.Bids) < 3 && manualOrder == nil {
		log.Printf("[%s] Too small BIDs amount: %d\n", tradeLimit.Symbol, len(marketDepth.Bids))
		return
	}

	price, err := m.PriceCalculator.CalculateBuy(tradeLimit)

	if err != nil {
		log.Printf("[%s] Price error: %s", tradeLimit.Symbol, err.Error())

		return
	}

	if manualOrder != nil && strings.ToUpper(manualOrder.Operation) == "BUY" {
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

		err = m.OrderExecutor.Buy(tradeLimit, tradeLimit.Symbol, price, quantity)
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

	if !m.TradeStack.CanBuy(tradeLimit) {
		log.Printf("[%s] Trade Stack check is not passed, wait order.", tradeLimit.Symbol)

		return
	}

	lastKline := m.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)

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

	marketDepth := m.PriceCalculator.GetDepth(tradeLimit.Symbol)
	manualOrder := m.OrderRepository.GetManualOrder(tradeLimit.Symbol)

	if len(marketDepth.Bids) < 3 && manualOrder == nil {
		log.Printf("[%s] Too small BIDs amount: %d\n", tradeLimit.Symbol, len(marketDepth.Bids))
		return
	}

	price, err := m.PriceCalculator.CalculateBuy(tradeLimit)
	if err != nil {
		log.Printf("[%s] Price error: %s", tradeLimit.Symbol, err.Error())

		return
	}

	profit := openedOrder.GetProfitPercent(lastKline.Close, m.BotService.UseSwapCapital())

	if profit.Lte(tradeLimit.GetBuyOnFallPercent(openedOrder, *lastKline, m.BotService.UseSwapCapital())) {
		// extra buy on current price
		if price < lastKline.Close {
			price = m.Formatter.FormatPrice(tradeLimit, lastKline.Close)
		}

		err = m.OrderExecutor.BuyExtra(tradeLimit, openedOrder, price)
		if err != nil {
			log.Printf("[%s] %s", tradeLimit.Symbol, err)

			m.OrderExecutor.TrySwap(openedOrder)
		}
	} else {
		log.Printf(
			"[%s] Extra charge is not allowed: %.2f of %.2f",
			tradeLimit.Symbol,
			profit.Value(),
			tradeLimit.GetBuyOnFallPercent(openedOrder, *lastKline, m.BotService.UseSwapCapital()).Value(),
		)
	}
}

func (m *MakerService) ProcessSell(tradeLimit model.TradeLimit, openedOrder model.Order) {
	lastKline := m.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)

	if lastKline == nil {
		log.Printf("[%s] Last price is unknown... skip!", tradeLimit.Symbol)

		return
	}

	manualOrder := m.OrderRepository.GetManualOrder(tradeLimit.Symbol)
	marketDepth := m.PriceCalculator.GetDepth(tradeLimit.Symbol)

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

	if manualOrder != nil && strings.ToUpper(manualOrder.Operation) == "SELL" {
		price = m.Formatter.FormatPrice(tradeLimit, manualOrder.Price)
		isManual = true
	}

	if price > 0 {
		quantity := m.Formatter.FormatQuantity(tradeLimit, m.OrderExecutor.CalculateSellQuantity(openedOrder))

		if quantity >= tradeLimit.MinQuantity {
			log.Printf("[%s] SELL QTY = %f", openedOrder.Symbol, quantity)
			err := m.OrderExecutor.Sell(tradeLimit, openedOrder, tradeLimit.Symbol, price, quantity, isManual)
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
		if !tradeLimit.IsEnabled {
			continue
		}

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
