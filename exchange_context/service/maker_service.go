package service

import (
	ExchangeClient "gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"log"
	"slices"
	"strings"
	"time"
)

type MakerService struct {
	OrderExecutor      *OrderExecutor
	OrderRepository    *ExchangeRepository.OrderRepository
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	Binance            *ExchangeClient.Binance
	Formatter          *Formatter
	MinDecisions       float64
	HoldScore          float64
	CurrentBot         *ExchangeModel.Bot
	PriceCalculator    *PriceCalculator
}

func (m *MakerService) Make(symbol string, decisions []ExchangeModel.Decision) {
	buyScore := 0.00
	sellScore := 0.00
	holdScore := 0.00

	sellVolume := 0.00
	buyVolume := 0.00
	smaValue := 0.00
	amount := 0.00
	priceSum := 0.00

	for _, decision := range decisions {
		if decision.StrategyName == "sma_trade_strategy" {
			buyVolume = decision.Params[0]
			sellVolume = decision.Params[1]
			smaValue = decision.Params[2]
		}
		amount = amount + 1.00
		switch decision.Operation {
		case "BUY":
			buyScore += decision.Score
			break
		case "SELL":
			sellScore += decision.Score
			break
		case "HOLD":
			holdScore += decision.Score
			break
		}
		priceSum += decision.Price
	}

	manualOrder := m.OrderRepository.GetManualOrder(symbol)

	if amount != m.MinDecisions && manualOrder == nil {
		return
	}

	tradeLimit, err := m.ExchangeRepository.GetTradeLimit(symbol)

	if err != nil {
		log.Printf("[%s] %s", symbol, err.Error())
		return
	}

	lastKline := m.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)

	if lastKline == nil {
		log.Printf("[%s] Last price is unknown... skip!", symbol)
		return
	}

	order, err := m.OrderRepository.GetOpenedOrderCached(symbol, "BUY")

	if m.OrderExecutor.ProcessSwap(order) {
		return
	}

	if err == nil && tradeLimit.IsExtraChargeEnabled() {
		profitPercent := order.GetProfitPercent(lastKline.Close)

		// If time to extra buy and price is near Low (Low + 0.5%)
		if profitPercent.Lte(tradeLimit.GetBuyOnFallPercent()) && lastKline.Close <= lastKline.GetLowPercent(0.5) {
			log.Printf(
				"[%s] Time to extra charge, profit %.2f of %.2f, price = %.8f",
				symbol,
				profitPercent,
				tradeLimit.GetBuyOnFallPercent().Value(),
				lastKline.Close,
			)
			holdScore = 0
		}
	}

	if manualOrder != nil {
		holdScore = 0
		if strings.ToUpper(manualOrder.Operation) == "BUY" {
			sellScore = 0
			buyScore = 999
		} else {
			sellScore = 999
			buyScore = 0
		}
	}

	// todo: fallback to existing order...

	if holdScore >= m.HoldScore {
		return
	}

	if sellScore >= buyScore {
		//log.Printf("[%s] Maker - H:%f, S:%f, B:%f\n", symbol, holdScore, sellScore, buyScore)

		marketDepth := m.PriceCalculator.GetDepth(tradeLimit.Symbol)

		if len(marketDepth.Asks) < 3 && manualOrder == nil {
			log.Printf("[%s] Too small ASKs amount: %d\n", symbol, len(marketDepth.Asks))
			return
		}

		if lastKline == nil {
			log.Printf("[%s] No information about current price", symbol)
			return
		}

		if err == nil {
			order, err := m.OrderRepository.GetOpenedOrderCached(symbol, "BUY")
			if err == nil {
				price := m.PriceCalculator.CalculateSell(tradeLimit, order)
				smaFormatted := m.Formatter.FormatPrice(tradeLimit, smaValue)

				if manualOrder != nil && strings.ToUpper(manualOrder.Operation) == "SELL" {
					price = m.Formatter.FormatPrice(tradeLimit, manualOrder.Price)
				}

				if price > 0 {
					quantity := m.Formatter.FormatQuantity(tradeLimit, m.OrderExecutor.CalculateSellQuantity(order))

					if quantity >= tradeLimit.MinQuantity {
						log.Printf("[%s] SELL QTY = %f", order.Symbol, quantity)
						err = m.OrderExecutor.Sell(tradeLimit, order, symbol, price, quantity, sellVolume, buyVolume, smaFormatted)
						if err != nil {
							log.Printf("[%s] SELL error: %s", order.Symbol, err.Error())
						}
					} else {
						log.Printf("[%s] SELL QTY = %f is too small!", order.Symbol, quantity)
					}

					if err != nil {
						log.Printf("[%s] %s", symbol, err)
					}
				} else {
					log.Printf("[%s] No BIDs on the market", symbol)
				}
			} else {
				log.Printf("[%s] Nothing to sell\n", symbol)
			}
		}

		return
	}

	if buyScore > sellScore {
		//log.Printf("[%s] Maker - H:%f, S:%f, B:%f\n", symbol, holdScore, sellScore, buyScore)
		tradeLimit, err := m.ExchangeRepository.GetTradeLimit(symbol)

		if !tradeLimit.IsEnabled {
			log.Printf("[%s] BUY operation is disabled", symbol)
			time.Sleep(time.Minute * 15)
			return
		}

		balanceErr := m.OrderExecutor.CheckMinBalance(tradeLimit)

		if balanceErr != nil {
			log.Printf("[%s] Min balance check: %s", tradeLimit.Symbol, balanceErr.Error())
			time.Sleep(time.Minute)
			return
		}

		marketDepth := m.PriceCalculator.GetDepth(tradeLimit.Symbol)

		if len(marketDepth.Bids) < 3 && manualOrder == nil {
			log.Printf("[%s] Too small BIDs amount: %d\n", symbol, len(marketDepth.Bids))
			return
		}

		if err == nil {
			price, err := m.PriceCalculator.CalculateBuy(tradeLimit)

			if err != nil {
				lastKline := m.ExchangeRepository.GetLastKLine(symbol)
				if lastKline == nil {
					log.Printf("[%s] Last price is unknown", symbol)
					return
				}

				log.Printf("[%s] %s, current = %f", symbol, err.Error(), lastKline.Close)
				return
			}

			order, err := m.OrderRepository.GetOpenedOrderCached(symbol, "BUY")

			if manualOrder != nil && strings.ToUpper(manualOrder.Operation) == "BUY" {
				price = m.Formatter.FormatPrice(tradeLimit, manualOrder.Price)
			}

			if err != nil {
				smaFormatted := m.Formatter.FormatPrice(tradeLimit, smaValue)

				if price > smaFormatted {
					log.Printf("[%s] Bad BUY price! SMA: %.6f, Price: %.6f\n", symbol, smaFormatted, price)
					return
				}

				if price > 0 {
					// todo: get buy quantity, buy to all cutlet! check available balance!
					quantity := m.Formatter.FormatQuantity(tradeLimit, tradeLimit.USDTLimit/price)

					if (quantity * price) < tradeLimit.MinNotional {
						log.Printf("[%s] BUY Notional: %.8f < %.8f", symbol, quantity*price, tradeLimit.MinNotional)
						return
					}

					err = m.OrderExecutor.Buy(tradeLimit, symbol, price, quantity, sellVolume, buyVolume, smaFormatted)
					if err != nil {
						log.Printf("[%s] %s", symbol, err)

						if strings.Contains(err.Error(), "not enough balance") {
							log.Printf("[%s] wait 1 minute...", symbol)
							time.Sleep(time.Minute * 1)
						}
					}
				} else {
					log.Printf("[%s] No ASKs on the market", symbol)
				}
			} else {
				lastKline := m.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)
				if lastKline != nil {
					profit := order.GetProfitPercent(lastKline.Close)

					if err == nil && profit.Lte(tradeLimit.GetBuyOnFallPercent()) {
						// extra buy on current price
						if price < lastKline.Close {
							price = m.Formatter.FormatPrice(tradeLimit, lastKline.Close)
						}

						err = m.OrderExecutor.BuyExtra(tradeLimit, order, price, sellVolume, buyVolume, smaValue)
						if err != nil {
							log.Printf("[%s] %s", symbol, err)

							m.OrderExecutor.TrySwap(order)
						}
					} else {
						log.Printf(
							"[%s] Extra charge is not allowed: %.2f of %.2f",
							symbol,
							profit.Value(),
							tradeLimit.GetBuyOnFallPercent().Value(),
						)
					}
				}
			}
		}
	}
}

func (m *MakerService) tradeLimit(symbol string) *ExchangeModel.TradeLimit {
	tradeLimits := m.ExchangeRepository.GetTradeLimits()
	for _, tradeLimit := range tradeLimits {
		if tradeLimit.Symbol == symbol {
			return &tradeLimit
		}
	}

	return nil
}

func (m *MakerService) UpdateSwapPairs() {
	swapMap := make(map[string][]ExchangeModel.ExchangeSymbol)
	exchangeInfo, _ := m.Binance.GetExchangeData(make([]string, 0))
	tradeLimits := m.ExchangeRepository.GetTradeLimits()

	supportedQuoteAssets := []string{"BTC", "ETH", "BNB", "TRX", "XRP", "EUR", "DAI", "TUSD", "USDC", "AUD", "TRY", "BRL"}

	for _, tradeLimit := range tradeLimits {
		if !tradeLimit.IsEnabled {
			continue
		}

		swapMap[tradeLimit.Symbol] = make([]ExchangeModel.ExchangeSymbol, 0)

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
				swapPair := ExchangeModel.SwapPair{
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
	limitMap := make(map[string]ExchangeModel.TradeLimit)
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
