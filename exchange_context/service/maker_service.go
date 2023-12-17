package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	ExchangeClient "gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"log"
	"slices"
	"strings"
	"sync"
	"time"
)

type MakerService struct {
	SwapValidator      *SwapValidator
	OrderRepository    *ExchangeRepository.OrderRepository
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	SwapRepository     *ExchangeRepository.SwapRepository
	Binance            *ExchangeClient.Binance
	LockChannel        *chan ExchangeModel.Lock
	Formatter          *Formatter
	FrameService       *FrameService
	Lock               map[string]bool
	TradeLockMutex     sync.RWMutex
	MinDecisions       float64
	HoldScore          float64
	RDB                *redis.Client
	Ctx                *context.Context
	CurrentBot         *ExchangeModel.Bot
	BalanceService     *BalanceService
	SwapEnabled        bool
	SwapSellOrderDays  int64
	SwapProfitPercent  float64
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

	if m.SwapEnabled && order.IsSwap() {
		log.Printf("[%s] Swap Order [%d] Mode: processing...", symbol, order.Id)
		m.ProcessSwap(order)
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
		log.Printf("[%s] Maker - H:%f, S:%f, B:%f\n", symbol, holdScore, sellScore, buyScore)

		marketDepth := m.GetDepth(tradeLimit.Symbol)

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
				price := m.calculateSellPrice(tradeLimit, order)
				smaFormatted := m.Formatter.FormatPrice(tradeLimit, smaValue)

				if manualOrder != nil && strings.ToUpper(manualOrder.Operation) == "SELL" {
					price = m.Formatter.FormatPrice(tradeLimit, manualOrder.Price)
				}

				if price > 0 {
					quantity := m.Formatter.FormatQuantity(tradeLimit, m.calculateSellQuantity(order))

					if quantity >= tradeLimit.MinQuantity {
						log.Printf("[%s] SELL QTY = %f", order.Symbol, quantity)
						err = m.Sell(tradeLimit, order, symbol, price, quantity, sellVolume, buyVolume, smaFormatted)
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
		log.Printf("[%s] Maker - H:%f, S:%f, B:%f\n", symbol, holdScore, sellScore, buyScore)
		tradeLimit, err := m.ExchangeRepository.GetTradeLimit(symbol)

		marketDepth := m.GetDepth(tradeLimit.Symbol)

		if len(marketDepth.Bids) < 3 && manualOrder == nil {
			log.Printf("[%s] Too small BIDs amount: %d\n", symbol, len(marketDepth.Bids))
			return
		}

		if err == nil {
			price, err := m.CalculateBuyPrice(tradeLimit)

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

					err = m.Buy(tradeLimit, symbol, price, quantity, sellVolume, buyVolume, smaFormatted)
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
				profit, err := m.getCurrentProfitPercent(order)

				if err == nil && profit.Lte(tradeLimit.GetBuyOnFallPercent()) {
					err = m.BuyExtra(tradeLimit, order, price, sellVolume, buyVolume, smaValue)
					if err != nil {
						log.Printf("[%s] %s", symbol, err)

						swapChain := m.SwapRepository.GetSwapChainCache(order.GetBaseAsset())
						if swapChain != nil && m.SwapEnabled {
							possibleSwaps := m.SwapRepository.GetSwapChains(order.GetBaseAsset())

							if len(possibleSwaps) > 0 {
								log.Printf("[%s] Found %d possible swaps...", order.Symbol, len(possibleSwaps))
							} else {
								m.SwapRepository.InvalidateSwapChainCache(order.GetBaseAsset())
							}

							for _, possibleSwap := range possibleSwaps {
								violation := m.SwapValidator.Validate(possibleSwap)

								if violation == nil {
									chainCurrentPercent := m.SwapValidator.CalculatePercent(possibleSwap)
									log.Printf(
										"[%s] EXTRA BUY FAILED -> Swap chain [%s] is found for order #%d, initial percent: %.2f, current = %.2f",
										order.Symbol,
										swapChain.Title,
										order.Id,
										swapChain.Percent,
										chainCurrentPercent,
									)
									m.makeSwap(order, possibleSwap)
								} else {
									log.Println(violation)
								}
							}
						}

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

func (m *MakerService) calculateSellQuantity(order ExchangeModel.Order) float64 {
	binanceOrder := m.OrderRepository.GetBinanceOrder(order.Symbol, "SELL")

	if binanceOrder != nil {
		return binanceOrder.OrigQty
	}

	m.recoverCommission(order)
	sellQuantity := order.GetRemainingToSellQuantity()
	balance, err := m.BalanceService.GetAssetBalance(order.GetBaseAsset())

	if err != nil {
		return sellQuantity
	}

	if balance > sellQuantity {
		// User can have own asset which bot is not allowed to sell!
		return sellQuantity
	}

	return balance
}

func (m *MakerService) calculateSellPrice(tradeLimit ExchangeModel.TradeLimit, order ExchangeModel.Order) float64 {
	marketDepth := m.GetDepth(tradeLimit.Symbol)
	avgPrice := marketDepth.GetBestAvgAsk()

	if 0.00 == avgPrice {
		return m.Formatter.FormatPrice(tradeLimit, avgPrice)
	}

	minPrice := m.Formatter.FormatPrice(tradeLimit, order.GetMinClosePrice(tradeLimit))
	openedOrder, err := m.OrderRepository.GetOpenedOrderCached(tradeLimit.Symbol, "BUY")

	if err != nil {
		return 0.00
	}

	lastKline := m.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)
	if lastKline == nil {
		return 0.00
	}

	currentPrice := lastKline.Close

	if avgPrice > minPrice {
		log.Printf("[%s] Choosen AVG sell price %f", tradeLimit.Symbol, avgPrice)
		minPrice = avgPrice
	}

	var frame ExchangeModel.Frame
	orderHours := openedOrder.GetHoursOpened()

	if orderHours >= 48.00 {
		log.Printf("[%s] Order is opened for %d hours, will be used 8-hours frame", tradeLimit.Symbol, orderHours)
		frame = m.FrameService.GetFrame(tradeLimit.Symbol, "2h", 4)
	} else {
		frame = m.FrameService.GetFrame(tradeLimit.Symbol, "2h", 8)
	}

	bestFrameSell, err := frame.GetBestFrameSell(marketDepth)

	if err == nil {
		if bestFrameSell[0] > minPrice {
			minPrice = bestFrameSell[0]
			log.Printf(
				"[%s] Choosen Frame [low:%f - high:%f] Sell price = %f",
				tradeLimit.Symbol,
				frame.AvgLow,
				frame.AvgHigh,
				minPrice,
			)
		}
	} else {
		log.Printf("[%s] Sell Frame: %s, current = %f", tradeLimit.Symbol, err.Error(), currentPrice)
	}

	if currentPrice > minPrice {
		minPrice = currentPrice
		log.Printf("[%s] Choosen Current sell price %f", tradeLimit.Symbol, minPrice)
	}

	profit := (minPrice - order.Price) * order.ExecutedQuantity

	log.Printf("[%s] Sell price = %f, expected profit = %f$", order.Symbol, minPrice, profit)

	return m.Formatter.FormatPrice(tradeLimit, minPrice)
}

func (m *MakerService) CalculateBuyPrice(tradeLimit ExchangeModel.TradeLimit) (float64, error) {
	// todo: check manual!!!!

	marketDepth := m.GetDepth(tradeLimit.Symbol)
	lastKline := m.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)

	if lastKline == nil {
		return 0.00, errors.New(fmt.Sprintf("[%s] Current price is unknown, wait...", tradeLimit.Symbol))
	}

	minPrice := m.ExchangeRepository.GetPeriodMinPrice(tradeLimit.Symbol, 200)
	order, err := m.OrderRepository.GetOpenedOrderCached(tradeLimit.Symbol, "BUY")

	// Extra charge by current price
	if err == nil && order.GetProfitPercent(lastKline.Close).Lte(tradeLimit.GetBuyOnFallPercent()) {
		extraBuyPrice := minPrice
		// todo: For next extras has to be used last extra order
		if order.GetHoursOpened() >= 24 {
			extraBuyPrice = lastKline.Close
			log.Printf(
				"[%s] Extra buy price is %f (more than 24 hours), profit: %.2f",
				tradeLimit.Symbol,
				extraBuyPrice,
				order.GetProfitPercent(lastKline.Close).Value(),
			)
		} else {
			extraBuyPrice = minPrice
			log.Printf(
				"[%s] Extra buy price is %f (less than 24 hours), profit: %.2f",
				tradeLimit.Symbol,
				extraBuyPrice,
				order.GetProfitPercent(lastKline.Close),
			)
		}

		if extraBuyPrice > lastKline.Close {
			extraBuyPrice = lastKline.Close
		}

		return m.Formatter.FormatPrice(tradeLimit, extraBuyPrice), nil
	}

	frame := m.FrameService.GetFrame(tradeLimit.Symbol, "2h", 6)
	bestFramePrice, err := frame.GetBestFrameBuy(tradeLimit, marketDepth)
	buyPrice := minPrice

	if err == nil {
		if buyPrice > bestFramePrice[1] {
			buyPrice = bestFramePrice[1]
		}
	} else {
		log.Printf("[%s] Buy Frame Error: %s, current = %f", tradeLimit.Symbol, err.Error(), lastKline.Close)
		potentialOpenPrice := lastKline.Close
		for {
			closePrice := potentialOpenPrice * (100 + tradeLimit.GetMinProfitPercent().Value()) / 100

			if closePrice <= frame.AvgHigh {
				break
			}

			potentialOpenPrice -= tradeLimit.MinPrice
		}

		if buyPrice > potentialOpenPrice {
			buyPrice = potentialOpenPrice
			log.Printf("[%s] Choosen potential open price = %f", tradeLimit.Symbol, buyPrice)
		}
	}

	if buyPrice > lastKline.Close {
		buyPrice = lastKline.Close
	}

	log.Printf(
		"[%s] Trade Frame [low:%f - high:%f](%.2f%s/%.2f%s): BUY Price = %f [min(200) = %f, current = %f]",
		tradeLimit.Symbol,
		frame.AvgLow,
		frame.AvgHigh,
		frame.GetMediumVolatilityPercent(),
		"%",
		frame.GetVolatilityPercent(),
		"%",
		buyPrice,
		minPrice,
		lastKline.Close,
	)

	return m.Formatter.FormatPrice(tradeLimit, buyPrice), nil
}

func (m *MakerService) BuyExtra(tradeLimit ExchangeModel.TradeLimit, order ExchangeModel.Order, price float64, sellVolume float64, buyVolume float64, smaValue float64) error {
	if tradeLimit.GetBuyOnFallPercent().Gte(0.00) {
		return errors.New(fmt.Sprintf("[%s] Extra buy is disabled", tradeLimit.Symbol))
	}

	if !order.CanExtraBuy(tradeLimit) {
		return errors.New(fmt.Sprintf("[%s] Not enough budget to buy more", tradeLimit.Symbol))
	}

	profit, _ := m.getCurrentProfitPercent(order)

	if profit.Gt(tradeLimit.GetBuyOnFallPercent()) {
		return errors.New(fmt.Sprintf(
			"[%s] Extra buy percent is not reached %.2f of %.2f",
			tradeLimit.Symbol,
			profit,
			tradeLimit.GetBuyOnFallPercent().Value()),
		)
	}

	m.acquireLock(order.Symbol)
	defer m.releaseLock(order.Symbol)
	// todo: get buy quantity, buy to all cutlet! check available balance!
	quantity := m.Formatter.FormatQuantity(tradeLimit, order.GetAvailableExtraBudget(tradeLimit)/price)

	if (quantity * price) < tradeLimit.MinNotional {
		return errors.New(fmt.Sprintf("[%s] Extra BUY Notional: %.8f < %.8f", order.Symbol, quantity*price, tradeLimit.MinNotional))
	}

	cached, _ := m.findBinanceOrder(order.Symbol, "BUY")

	// Check balance for new order
	if cached == nil {
		usdtAvailableBalance, err := m.BalanceService.GetAssetBalance("USDT")

		if err != nil {
			return errors.New(fmt.Sprintf("[%s] BUY balance error: %s", order.Symbol, err.Error()))
		}

		requiredUsdtAmount := price * quantity

		if requiredUsdtAmount > usdtAvailableBalance {
			return errors.New(fmt.Sprintf("[%s] BUY not enough balance: %f/%f", order.Symbol, usdtAvailableBalance, requiredUsdtAmount))
		}
	}

	var extraOrder = ExchangeModel.Order{
		Symbol:      order.Symbol,
		Quantity:    quantity,
		Price:       price,
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
		SellVolume:  sellVolume,
		BuyVolume:   buyVolume,
		SmaValue:    smaValue,
		Status:      "closed",
		Operation:   "buy",
		ExternalId:  nil,
		ClosesOrder: &order.Id,
		// todo: add commission???
	}

	balanceBefore, balanceErr := m.BalanceService.GetAssetBalance(order.GetBaseAsset())

	binanceOrder, err := m.tryLimitOrder(extraOrder, "BUY", 120)

	if err != nil {
		m.BalanceService.InvalidateBalanceCache("USDT")
		return err
	}

	executedQty := binanceOrder.GetExecutedQuantity()

	// fill from API
	extraOrder.ExternalId = &binanceOrder.OrderId
	extraOrder.ExecutedQuantity = executedQty
	extraOrder.Price = binanceOrder.Price
	extraOrder.CreatedAt = time.Now().Format("2006-01-02 15:04:05")

	avgPrice := m.getAvgPrice(order, extraOrder)

	_, err = m.OrderRepository.Create(extraOrder)
	if err != nil {
		return err
	}

	if balanceErr == nil {
		m.UpdateCommission(balanceBefore, extraOrder)
	}

	order.ExecutedQuantity = executedQty + order.ExecutedQuantity
	order.Price = avgPrice
	order.UsedExtraBudget = order.UsedExtraBudget + (extraOrder.Price * executedQty)
	commission := 0.00
	if order.Commission != nil {
		commission = *order.Commission
	}
	extraCommission := 0.00
	if extraOrder.Commission != nil {
		extraCommission = *extraOrder.Commission
	}
	commissionSum := commission + extraCommission
	if commissionSum < 0 {
		commissionSum = 0
	}

	order.Commission = &commissionSum

	err = m.OrderRepository.Update(order)
	_, err = m.OrderRepository.Create(order)
	m.BalanceService.InvalidateBalanceCache("USDT")
	m.BalanceService.InvalidateBalanceCache(order.GetBaseAsset())

	if err != nil {
		return err
	}

	m.OrderRepository.DeleteManualOrder(order.Symbol)

	return nil
}

func (m *MakerService) getCurrentProfitPercent(order ExchangeModel.Order) (ExchangeModel.Percent, error) {
	lastKline := m.ExchangeRepository.GetLastKLine(order.Symbol)

	if lastKline == nil {
		return ExchangeModel.Percent(0.00), errors.New(fmt.Sprintf("[%s] Do not have info about the price", order.Symbol))
	}

	return order.GetProfitPercent(lastKline.Close), nil
}

func (m *MakerService) getAvgPrice(opened ExchangeModel.Order, extra ExchangeModel.Order) float64 {
	return ((opened.ExecutedQuantity * opened.Price) + (extra.ExecutedQuantity * extra.Price)) / (opened.ExecutedQuantity + extra.ExecutedQuantity)
}

func (m *MakerService) Buy(tradeLimit ExchangeModel.TradeLimit, symbol string, price float64, quantity float64, sellVolume float64, buyVolume float64, smaValue float64) error {
	if !tradeLimit.IsEnabled {
		return errors.New(fmt.Sprintf("[%s] BUY operation is disabled", symbol))
	}

	if m.isTradeLocked(symbol) {
		return errors.New(fmt.Sprintf("Operation Buy is Locked %s", symbol))
	}

	if quantity <= 0.00 {
		return errors.New(fmt.Sprintf("Available quantity is %f", quantity))
	}

	cached, _ := m.findBinanceOrder(symbol, "BUY")

	// Check balance for new order
	if cached == nil {
		usdtAvailableBalance, err := m.BalanceService.GetAssetBalance("USDT")

		if err != nil {
			return errors.New(fmt.Sprintf("[%s] BUY balance error: %s", symbol, err.Error()))
		}

		requiredUsdtAmount := price * quantity

		if requiredUsdtAmount > usdtAvailableBalance {
			return errors.New(fmt.Sprintf("[%s] BUY not enough balance: %f/%f", symbol, usdtAvailableBalance, requiredUsdtAmount))
		}
	}

	// to avoid concurrent map writes
	m.acquireLock(symbol)
	defer m.releaseLock(symbol)

	// todo: commission
	// You place an order to buy 10 ETH for 3,452.55 USDT each:
	// Trading fee = 10 ETH * 0.1% = 0.01 ETH

	// todo: check min quantity

	var order = ExchangeModel.Order{
		Symbol:      symbol,
		Quantity:    quantity,
		Price:       price,
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
		SellVolume:  sellVolume,
		BuyVolume:   buyVolume,
		SmaValue:    smaValue,
		Status:      "opened",
		Operation:   "buy",
		ExternalId:  nil,
		ClosesOrder: nil,
		// todo: add commission???
	}

	balanceBefore, balanceErr := m.BalanceService.GetAssetBalance(order.GetBaseAsset())

	binanceOrder, err := m.tryLimitOrder(order, "BUY", 480)

	if err != nil {
		m.BalanceService.InvalidateBalanceCache("USDT")
		return err
	}

	// fill from API
	order.ExternalId = &binanceOrder.OrderId
	order.ExecutedQuantity = binanceOrder.GetExecutedQuantity()
	order.Price = binanceOrder.Price
	order.CreatedAt = time.Now().Format("2006-01-02 15:04:05")

	_, err = m.OrderRepository.Create(order)
	m.BalanceService.InvalidateBalanceCache("USDT")
	m.BalanceService.InvalidateBalanceCache(order.GetBaseAsset())

	if err != nil {
		log.Printf("Can't create order: %s", order)

		return err
	}

	m.OrderRepository.DeleteManualOrder(order.Symbol)

	if balanceErr == nil {
		m.UpdateCommission(balanceBefore, order)
	}

	return nil
}

func (m *MakerService) Sell(tradeLimit ExchangeModel.TradeLimit, opened ExchangeModel.Order, symbol string, price float64, quantity float64, sellVolume float64, buyVolume float64, smaValue float64) error {
	if m.isTradeLocked(symbol) {
		return errors.New(fmt.Sprintf("Operation Sell is Locked %s", symbol))
	}

	m.acquireLock(symbol)
	defer m.releaseLock(symbol)

	// todo: commission
	// Or you place an order to sell 10 ETH for 3,452.55 USDT each:
	// Trading fee = (10 ETH * 3,452.55 USDT) * 0.1% = 34.5255 USDT

	profit := (price - opened.Price) * quantity

	// loose money control
	if opened.Price >= price {
		return errors.New(fmt.Sprintf(
			"[%s] Bad deal, wait for positive profit: %.6f [o:%.6f, c:%.6f]",
			symbol,
			profit,
			opened.Price,
			price,
		))
	}

	minPrice := m.Formatter.FormatPrice(tradeLimit, opened.GetMinClosePrice(tradeLimit))

	if price < minPrice {
		return errors.New(fmt.Sprintf(
			"[%s] Minimum profit is not reached, Price %.6f < %.6f",
			symbol,
			price,
			minPrice,
		))
	}

	var order = ExchangeModel.Order{
		Symbol:      symbol,
		Quantity:    quantity,
		Price:       price,
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
		SellVolume:  sellVolume,
		BuyVolume:   buyVolume,
		SmaValue:    smaValue,
		Status:      "closed",
		Operation:   "sell",
		ExternalId:  nil,
		ClosesOrder: &opened.Id,
		// todo: add commission???
	}

	//balanceBefore, balanceErr := m.BalanceService.GetAssetBalanceorder.GetBaseAsset())

	binanceOrder, err := m.tryLimitOrder(order, "SELL", 480)

	if err != nil {
		m.BalanceService.InvalidateBalanceCache(order.GetBaseAsset())
		return err
	}

	// fill from API
	order.ExternalId = &binanceOrder.OrderId
	order.ExecutedQuantity = binanceOrder.GetExecutedQuantity()
	order.Price = binanceOrder.Price
	order.CreatedAt = time.Now().Format("2006-01-02 15:04:05")

	lastId, err := m.OrderRepository.Create(order)

	if err != nil {
		log.Printf("Can't create order: %s", order)

		return err
	}

	m.OrderRepository.DeleteManualOrder(order.Symbol)
	_, err = m.OrderRepository.Find(*lastId)

	if err != nil {
		log.Printf("Can't get created order [%d]: %s", lastId, order)

		return err
	}

	closings := m.OrderRepository.GetClosesOrderList(opened)
	totalExecuted := 0.00
	for _, closeOrder := range closings {
		if closeOrder.IsClosed() {
			totalExecuted += closeOrder.ExecutedQuantity
		}
	}

	// commission can be around 0.4% (0.2% to one side)
	// @see https://www.binance.com/en/fee/trading
	if (totalExecuted + (totalExecuted * 0.004 * float64(len(closings)))) >= opened.ExecutedQuantity {
		opened.Status = "closed"
	}

	err = m.OrderRepository.Update(opened)
	m.BalanceService.InvalidateBalanceCache("USDT")
	m.BalanceService.InvalidateBalanceCache(opened.GetBaseAsset())

	if err != nil {
		log.Printf("Can't udpdate order [%d]: %s", order.Id, order)

		return err
	}

	//if balanceErr == nil {
	//	m.UpdateCommission(balanceBefore, order)
	//}

	return nil
}

// todo: order has to be Interface
func (m *MakerService) tryLimitOrder(order ExchangeModel.Order, operation string, ttl int64) (ExchangeModel.BinanceOrder, error) {
	// todo: extra order flag...
	binanceOrder, err := m.findOrCreateOrder(order, operation)

	if err != nil {
		return binanceOrder, err
	}

	binanceOrder, err = m.waitExecution(binanceOrder, ttl)

	if err != nil {
		return binanceOrder, err
	}

	return binanceOrder, nil
}

func (m *MakerService) waitExecution(binanceOrder ExchangeModel.BinanceOrder, seconds int64) (ExchangeModel.BinanceOrder, error) {
	depth := m.GetDepth(binanceOrder.Symbol)

	var currentPosition int
	var book [2]ExchangeModel.Number
	if "BUY" == binanceOrder.Side {
		currentPosition, book = depth.GetBidPosition(binanceOrder.Price)
	} else {
		currentPosition, book = depth.GetAskPosition(binanceOrder.Price)
	}
	log.Printf(
		"[%s] Order Book start position is [%d] %.6f\n",
		binanceOrder.Symbol,
		currentPosition,
		book[0],
	)

	executedQty := 0.00

	orderManageChannel := make(chan string)
	control := make(chan string)
	defer close(orderManageChannel)
	defer close(control)
	defer m.OrderRepository.DeleteBinanceOrder(binanceOrder)

	tradeLimit := m.tradeLimit(binanceOrder.Symbol)

	go func(
		tradeLimit ExchangeModel.TradeLimit,
		binanceOrder *ExchangeModel.BinanceOrder,
		ttl *int64,
		control chan string,
		orderManageChannel chan string,
	) {
		timer := 0
		allowedExtraRequest := true
		start := time.Now().Unix()

		for {
			if binanceOrder.IsCanceled() || binanceOrder.IsExpired() || binanceOrder.IsFilled() {
				orderManageChannel <- "status"
				action := <-control
				if action == "stop" {
					return
				}

				time.Sleep(time.Second * 40)
				continue
			}

			end := time.Now().Unix()
			kline := m.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)

			if binanceOrder.IsSell() && binanceOrder.IsNew() && m.SwapEnabled {
				openedBuyPosition, err := m.OrderRepository.GetOpenedOrderCached(binanceOrder.Symbol, "BUY")
				// Try arbitrage for long orders >= 4 hours and with profit < -1.00%
				if err == nil && openedBuyPosition.GetHoursOpened() >= m.SwapSellOrderDays && openedBuyPosition.GetProfitPercent(kline.Close).Lte(ExchangeModel.Percent(m.SwapProfitPercent)) && !openedBuyPosition.IsSwap() {
					swapChain := m.SwapRepository.GetSwapChainCache(openedBuyPosition.GetBaseAsset())
					if swapChain != nil {
						possibleSwaps := m.SwapRepository.GetSwapChains(openedBuyPosition.GetBaseAsset())

						if len(possibleSwaps) > 0 {
							log.Printf("[%s] Found %d possible swaps...", openedBuyPosition.Symbol, len(possibleSwaps))
						} else {
							m.SwapRepository.InvalidateSwapChainCache(openedBuyPosition.GetBaseAsset())
						}

						for _, possibleSwap := range possibleSwaps {
							violation := m.SwapValidator.Validate(possibleSwap)

							if violation == nil {
								chainCurrentPercent := m.SwapValidator.CalculatePercent(possibleSwap)
								log.Printf(
									"[%s] Swap chain [%s] is found for order #%d, initial percent: %.2f, current = %.2f",
									binanceOrder.Symbol,
									swapChain.Title,
									openedBuyPosition.Id,
									swapChain.Percent,
									chainCurrentPercent,
								)
								orderManageChannel <- "status"
								action := <-control
								if action == "stop" {
									return
								}

								if binanceOrder.IsNew() {
									log.Printf("[%s] Cancel signal sent!", binanceOrder.Symbol)
									orderManageChannel <- "cancel"
									action := <-control
									if action == "stop" {
										m.makeSwap(openedBuyPosition, possibleSwap)
										return
									}
								}
							} else {
								log.Println(violation)
							}
						}
					}
				}
			}

			if binanceOrder.IsBuy() && binanceOrder.IsNew() && binanceOrder.Price > kline.Close {
				fallPercent := ExchangeModel.Percent(100 - m.Formatter.ComparePercentage(binanceOrder.Price, kline.Close).Value())
				minPrice := m.ExchangeRepository.GetPeriodMinPrice(tradeLimit.Symbol, 200)

				cancelFallPercent := ExchangeModel.Percent(0.50)

				// If falls more then (min - 0.5%) cancel current
				if fallPercent.Gte(cancelFallPercent) && minPrice-(minPrice*0.005) > kline.Close && allowedExtraRequest {
					log.Printf("[%s] Check status signal sent!", binanceOrder.Symbol)
					allowedExtraRequest = false
					orderManageChannel <- "status"
					action := <-control
					if action == "stop" {
						return
					}
					log.Printf("[%s] Order status is [%s]", binanceOrder.Symbol, binanceOrder.Status)

					if binanceOrder.IsNew() {
						log.Printf("[%s] Cancel signal sent!", binanceOrder.Symbol)
						orderManageChannel <- "cancel"
						action := <-control
						if action == "stop" {
							return
						}
					}
				}
			}

			// [BUY] Check is it time to sell (maybe we have already partially filled)
			if binanceOrder.IsBuy() && binanceOrder.IsPartiallyFilled() && binanceOrder.GetProfitPercent(kline.Close).Gte(tradeLimit.GetMinProfitPercent()) {
				log.Printf(
					"[%s] Max profit percent reached, current profit is: %.2f, %s [%d] order is cancelled",
					binanceOrder.Symbol,
					binanceOrder.GetProfitPercent(kline.Close).Value(),
					binanceOrder.Side,
					binanceOrder.OrderId,
				)
				orderManageChannel <- "cancel"
				action := <-control
				if action == "stop" {
					return
				}
			}

			// Check is time to extra buy, but we have sell partial...
			if binanceOrder.IsSell() && binanceOrder.IsPartiallyFilled() {
				openedBuyPosition, err := m.OrderRepository.GetOpenedOrderCached(binanceOrder.Symbol, "BUY")
				if err == nil && openedBuyPosition.GetProfitPercent(kline.Close).Lte(tradeLimit.GetBuyOnFallPercent()) {
					log.Printf(
						"[%s] Extra Charge percent reached, current profit is: %.2f, SELL order is cancelled",
						binanceOrder.Symbol,
						openedBuyPosition.GetProfitPercent(kline.Close).Value(),
					)
					orderManageChannel <- "cancel"
					action := <-control
					if action == "stop" {
						return
					}
				}
			}

			if timer > 30000 {
				orderManageChannel <- "status"
				action := <-control
				log.Printf(
					"[%s] %s Order [%d] status [%s] wait handler (%s), current price is [%.8f], order price [%.8f], ExecutedQty: %.6f of %.6f\"",
					binanceOrder.Symbol,
					binanceOrder.Side,
					binanceOrder.OrderId,
					binanceOrder.Status,
					action,
					kline.Close,
					binanceOrder.Price,
					binanceOrder.ExecutedQty,
					binanceOrder.OrigQty,
				)
				if action == "stop" {
					return
				}

				timer = 0
				time.Sleep(time.Second)

				// check only new timeout
				if end >= (start+*ttl) && binanceOrder.IsNew() {
					if binanceOrder.IsSell() {
						openedBuyPosition, err := m.OrderRepository.GetOpenedOrderCached(binanceOrder.Symbol, "BUY")
						if err == nil {
							profitPercent := openedBuyPosition.GetProfitPercent(kline.Close)
							if profitPercent.Lte(0.00) {
								log.Printf(
									"[%s] %s Order [%d] status [%s] ttl reached, current price is [%.8f], order price [%.8f], open [%.8f], profit: %.2f",
									binanceOrder.Symbol,
									binanceOrder.Side,
									binanceOrder.OrderId,
									binanceOrder.Status,
									kline.Close,
									binanceOrder.Price,
									openedBuyPosition.Price,
									profitPercent.Value(),
								)
								orderManageChannel <- "cancel"
								action := <-control
								if action == "stop" {
									return
								}
							} else {
								log.Printf(
									"[%s] %s Order [%d] status [%s] ttl ignored, current price is [%.8f], order price [%.8f], open [%.8f], profit: %.2f",
									binanceOrder.Symbol,
									binanceOrder.Side,
									binanceOrder.OrderId,
									binanceOrder.Status,
									kline.Close,
									binanceOrder.Price,
									openedBuyPosition.Price,
									profitPercent.Value(),
								)
							}
						} else {
							log.Printf(
								"[%s] %s Order [%d] %s",
								binanceOrder.Symbol,
								binanceOrder.Side,
								binanceOrder.OrderId,
								err.Error(),
							)
							orderManageChannel <- "cancel"
							action := <-control
							if action == "stop" {
								return
							}
						}
					}

					if binanceOrder.IsBuy() {
						positionPercentage := m.Formatter.ComparePercentage(binanceOrder.Price, kline.Close)
						if positionPercentage.Gte(101) {
							log.Printf(
								"[%s] %s Order [%d] status [%s] ttl reached, current price is [%.8f], order price [%.8f], diff percent: %.2f",
								binanceOrder.Symbol,
								binanceOrder.Side,
								binanceOrder.OrderId,
								binanceOrder.Status,
								kline.Close,
								binanceOrder.Price,
								positionPercentage.Value(),
							)
							orderManageChannel <- "cancel"
							action := <-control
							if action == "stop" {
								return
							}
						} else {
							log.Printf(
								"[%s] %s Order [%d] status [%s] ttl ignored, current price is [%.8f], order price [%.8f], diff percent: %.2f",
								binanceOrder.Symbol,
								binanceOrder.Side,
								binanceOrder.OrderId,
								binanceOrder.Status,
								kline.Close,
								binanceOrder.Price,
								positionPercentage.Value(),
							)
						}
					}
				}
			} else {
				manualOrder := m.OrderRepository.GetManualOrder(binanceOrder.Symbol)
				// cancel current immediately on new manual order
				if manualOrder != nil && manualOrder.Price != binanceOrder.Price {
					orderManageChannel <- "cancel"
					action := <-control
					if action == "stop" {
						return
					}
				} else {
					orderManageChannel <- "continue"
					action := <-control
					if action == "stop" {
						return
					}
				}
			}

			allowedExtraRequest = true
			timer = timer + 30
			time.Sleep(time.Millisecond * 20)
		}
	}(*tradeLimit, &binanceOrder, &seconds, control, orderManageChannel)

	for {
		action := <-orderManageChannel
		if action == "continue" {
			control <- "continue"
			continue
		}

		if action == "cancel" {
			log.Printf(
				"[%s] %s Order %d, cancel signal has received",
				binanceOrder.Symbol,
				binanceOrder.Side,
				binanceOrder.OrderId,
			)
			break
		}

		queryOrder, err := m.Binance.QueryOrder(binanceOrder.Symbol, binanceOrder.OrderId)

		if err != nil {
			log.Printf("[%s] QueryOrder: %s", binanceOrder.Symbol, err.Error())

			if strings.Contains(err.Error(), "Order was canceled or expired") {
				break
			}

			if strings.Contains(err.Error(), "Order does not exist") {
				break
			}

			log.Printf("[%s] Retry query order...", binanceOrder.Symbol)
			time.Sleep(time.Minute * 2)

			control <- "continue"
			continue
		}

		binanceOrder = queryOrder

		if binanceOrder.IsPartiallyFilled() {
			// Add 5 minutes more if ExecutedQty moves up!
			if binanceOrder.GetExecutedQuantity() > executedQty {
				seconds = seconds + (60 * 5)
			}

			executedQty = binanceOrder.GetExecutedQuantity()
			m.OrderRepository.SetBinanceOrder(binanceOrder)
			control <- "continue"
			continue
		}

		if binanceOrder.IsExpired() {
			if binanceOrder.HasExecutedQuantity() {
				control <- "stop"
				return binanceOrder, nil
			}

			break
		}

		if binanceOrder.IsCanceled() {
			if binanceOrder.HasExecutedQuantity() {
				control <- "stop"
				return binanceOrder, nil
			}

			break
		}

		if binanceOrder.IsFilled() {
			log.Printf("[%s] Order [%d] is executed [%s]", binanceOrder.Symbol, binanceOrder.OrderId, binanceOrder.Status)

			control <- "stop"
			return binanceOrder, nil
		}

		control <- "continue"
	}

	// If you cancel an order that has already been partially filled,
	// the cryptocurrency or fiat currency that was used to fill the order
	// will be returned to your account. The remaining cryptocurrency or fiat
	// currency will be used to fill other orders that are waiting to be executed.
	// {
	//    "symbol": "ETHUSDT",
	//    "origClientOrderId": "aSUn6e7pktn5fVFuNEb0TK",
	//    "orderId": 31314,
	//    "orderListId": -1,
	//    "clientOrderId": "wnpmGUt6RgoyuZB48NXbFG",
	//    "transactTime": 1701886419972,
	//    "price": "2100.93000000",
	//    "origQty": "0.04750000",
	//    "executedQty": "0.04670000",
	//    "cummulativeQuoteQty": "98.11343100",
	//    "status": "CANCELED",
	//    "timeInForce": "GTC",
	//    "type": "LIMIT",
	//    "side": "BUY",
	//    "selfTradePreventionMode": "EXPIRE_MAKER"
	//}
	cancelOrder, err := m.Binance.CancelOrder(binanceOrder.Symbol, binanceOrder.OrderId)

	if err != nil {
		// Possible case: {"code": -2011,"msg": "Order was not canceled due to cancel restrictions."}

		log.Printf("[%s] Cancel failed: %s", binanceOrder.Symbol, err.Error())
		queryOrder, retryErr := m.Binance.QueryOrder(binanceOrder.Symbol, binanceOrder.OrderId)

		if retryErr == nil {
			binanceOrder = queryOrder
			control <- "stop"
			log.Printf("[%s] Order [%d] is recovered [%s]", binanceOrder.Symbol, binanceOrder.OrderId, binanceOrder.Status)

			if binanceOrder.IsFilled() {
				return binanceOrder, nil
			}

			// Just in case of bug...
			if binanceOrder.IsPartiallyFilled() {
				log.Printf(
					"[%s] Order [%d] status is [%s], try again waitExecution...",
					binanceOrder.Symbol,
					binanceOrder.OrderId,
					binanceOrder.Status,
				)

				return m.waitExecution(binanceOrder, 120)
			}

			// Just in case of bug...
			if binanceOrder.IsNew() {
				log.Printf(
					"[%s] Order [%d] status is [%s], try again waitExecution...",
					binanceOrder.Symbol,
					binanceOrder.OrderId,
					binanceOrder.Status,
				)

				return m.waitExecution(binanceOrder, 120)
			}

			if binanceOrder.HasExecutedQuantity() {
				log.Printf(
					"Order [%d] is [%s], ExecutedQty = %.8f",
					binanceOrder.OrderId,
					binanceOrder.Status,
					binanceOrder.GetExecutedQuantity(),
				)

				return binanceOrder, nil
			} else {
				return binanceOrder, errors.New(fmt.Sprintf("Order %d was CANCELED", binanceOrder.OrderId))
			}
		} else {
			// todo: loop??? timeout + loop???
			control <- "stop"
			return binanceOrder, err
		}
	}

	binanceOrder = cancelOrder
	control <- "stop"

	// handle cancel error and get again

	if binanceOrder.HasExecutedQuantity() {
		log.Printf(
			"Order [%d] is [%s], ExecutedQty = %.8f",
			binanceOrder.OrderId,
			binanceOrder.Status,
			binanceOrder.GetExecutedQuantity(),
		)

		return binanceOrder, nil
	}

	log.Printf("Order [%d] is [%s]", binanceOrder.OrderId, binanceOrder.Status)

	return binanceOrder, errors.New(fmt.Sprintf("Order %d was CANCELED", binanceOrder.OrderId))
}

func (m *MakerService) isTradeLocked(symbol string) bool {
	m.TradeLockMutex.Lock()
	isLocked, _ := m.Lock[symbol]
	m.TradeLockMutex.Unlock()

	return isLocked
}

func (m *MakerService) acquireLock(symbol string) {
	*m.LockChannel <- ExchangeModel.Lock{IsLocked: true, Symbol: symbol}
}

func (m *MakerService) releaseLock(symbol string) {
	*m.LockChannel <- ExchangeModel.Lock{IsLocked: false, Symbol: symbol}
}

func (m *MakerService) findBinanceOrder(symbol string, operation string) (*ExchangeModel.BinanceOrder, error) {
	cached := m.OrderRepository.GetBinanceOrder(symbol, operation)

	if cached != nil {
		log.Printf("[%s] Found cached %s order %d in binance", symbol, operation, cached.OrderId)

		return cached, nil
	}

	openedOrders, err := m.Binance.GetOpenedOrders()

	if err != nil {
		log.Printf("[%s] Opened: %s", symbol, err.Error())
		return nil, err
	}

	for _, opened := range *openedOrders {
		if opened.Side == operation && opened.Symbol == symbol {
			log.Printf("[%s] Found opened %s order %d in binance", symbol, operation, opened.OrderId)
			m.OrderRepository.SetBinanceOrder(opened)

			return &opened, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("[%s] Binance order is not found", symbol))
}

func (m *MakerService) findOrCreateOrder(order ExchangeModel.Order, operation string) (ExchangeModel.BinanceOrder, error) {
	// todo: extra order flag...
	cached, err := m.findBinanceOrder(order.Symbol, operation)

	if cached != nil {
		log.Printf("[%s] Found cached %s order %d in binance", order.Symbol, operation, cached.OrderId)

		return *cached, nil
	}

	binanceOrder, err := m.Binance.LimitOrder(order.Symbol, order.Quantity, order.Price, operation, "GTC")

	if err != nil {
		log.Printf("[%s] Limit: %s", order.Symbol, err.Error())
		return binanceOrder, err
	}

	log.Printf("[%s] %s Order created %d, Price: %.6f", order.Symbol, operation, binanceOrder.OrderId, binanceOrder.Price)
	m.OrderRepository.SetBinanceOrder(binanceOrder)
	if order.IsBuy() {
		m.BalanceService.InvalidateBalanceCache("USDT")
	} else {
		m.BalanceService.InvalidateBalanceCache(order.GetBaseAsset())
	}

	return binanceOrder, nil
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

func (m *MakerService) SetDepth(depth ExchangeModel.Depth) {
	m.ExchangeRepository.SetDepth(depth)
}

func (m *MakerService) GetDepth(symbol string) ExchangeModel.Depth {
	depth := m.ExchangeRepository.GetDepth(symbol)

	if len(depth.Asks) == 0 && len(depth.Bids) == 0 {
		book, err := m.Binance.GetDepth(symbol)
		if err == nil {
			depth = book.ToDepth(symbol)
			m.SetDepth(depth)
		}
	}

	return depth
}

func (m *MakerService) UpdateSwapPairs() {
	swapMap := make(map[string][]ExchangeModel.ExchangeSymbol)
	exchangeInfo, _ := m.Binance.GetExchangeData(make([]string, 0))
	tradeLimits := m.ExchangeRepository.GetTradeLimits()

	supportedQuoteAssets := []string{"BTC", "ETH", "BNB", "TRX", "XRP", "EUR", "DAI", "TUSD", "GBP"}

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

func (m *MakerService) recoverCommission(order ExchangeModel.Order) {
	if order.Commission != nil {
		return
	}
	assetSymbol := order.GetBaseAsset()

	balanceAfter, err := m.BalanceService.GetAssetBalance(assetSymbol)

	if err != nil {
		log.Printf("[%s] Can't recover commission: %s", order.Status, err.Error())
		return
	}

	commission := order.ExecutedQuantity - balanceAfter
	if commission < 0 {
		commission = 0
	}
	order.Commission = &commission
	order.CommissionAsset = &assetSymbol

	err = m.OrderRepository.Update(order)
	if err != nil {
		log.Printf("[%s] Order Commission Recover: %s", order.Symbol, err.Error())
	}
}

func (m *MakerService) UpdateCommission(balanceBefore float64, order ExchangeModel.Order) {
	assetSymbol := order.GetBaseAsset()
	balanceAfter, err := m.BalanceService.GetAssetBalance(assetSymbol)

	if err != nil {
		log.Printf("[%s] Can't update commission: %s", order.Status, err.Error())
		return
	}

	arrived := balanceAfter - balanceBefore

	commission := order.ExecutedQuantity - arrived

	if commission < 0 {
		commission = 0.00
	}

	order.Commission = &commission
	order.CommissionAsset = &assetSymbol

	err = m.OrderRepository.Update(order)
	if err != nil {
		log.Printf("[%s] Order Commission Update: %s", order.Symbol, err.Error())
	}
}

func (m *MakerService) makeSwap(order ExchangeModel.Order, swapChain ExchangeModel.SwapChainEntity) {
	assetBalance, err := m.BalanceService.GetAssetBalance(swapChain.SwapOne.BaseAsset)

	if err != nil {
		return
	}

	swapAction, err := m.SwapRepository.GetActiveSwapAction(order)

	if err == nil {
		log.Printf("[%s] Swap has already exists: %s", swapChain.SwapOne.BaseAsset, swapAction.Status)

		return
	}

	// todo: transaction
	// create swap
	_, err = m.SwapRepository.CreateSwapAction(ExchangeModel.SwapAction{
		Id:              0,
		OrderId:         order.Id,
		BotId:           m.CurrentBot.Id,
		SwapChainId:     swapChain.Id,
		Asset:           swapChain.SwapOne.BaseAsset,
		Status:          ExchangeModel.SwapActionStatusPending,
		StartTimestamp:  time.Now().Unix(),
		StartQuantity:   assetBalance,
		SwapOneSymbol:   swapChain.SwapOne.GetSymbol(),
		SwapOnePrice:    swapChain.SwapOne.Price,
		SwapTwoSymbol:   swapChain.SwapTwo.GetSymbol(),
		SwapTwoPrice:    swapChain.SwapTwo.Price,
		SwapThreeSymbol: swapChain.SwapThree.GetSymbol(),
		SwapThreePrice:  swapChain.SwapThree.Price,
	})

	if err != nil {
		log.Printf(
			"[%s] Swap couldn't be created: %s",
			swapChain.SwapOne.BaseAsset,
			err.Error(),
		)

		return
	}

	// enable order swap mode
	order.Swap = true
	err = m.OrderRepository.Update(order)
	if err == nil {
		log.Printf("[%s] Swap order mode enabled [%s]", order.Symbol, swapChain.Title)
	}
}

func (m *MakerService) ProcessSwap(order ExchangeModel.Order) {
	swapAction, err := m.SwapRepository.GetActiveSwapAction(order)

	if err != nil {
		log.Printf("[%s] Swap processing error: %s", order.Symbol, err.Error())

		if strings.Contains(err.Error(), "no rows in result set") {
			order.Swap = false
			_ = m.OrderRepository.Update(order)
		}

		return
	}

	balanceBefore, _ := m.BalanceService.GetAssetBalance(swapAction.Asset)

	if swapAction.IsPending() {
		swapAction.Status = ExchangeModel.SwapActionStatusProcess
		_ = m.SwapRepository.UpdateSwapAction(swapAction)
	}

	var swapOneOrder *ExchangeModel.BinanceOrder = nil

	swapChain, err := m.SwapRepository.GetSwapChainById(swapAction.SwapChainId)

	if err != nil {
		log.Printf("Swap chain %d is not found", swapAction.SwapChainId)
		return
	}

	if swapAction.SwapOneExternalId == nil {
		swapPair, err := m.SwapRepository.GetSwapPairBySymbol(swapAction.SwapOneSymbol)
		var binanceOrder ExchangeModel.BinanceOrder

		if swapChain.IsSSB() || swapChain.IsSBB() {
			binanceOrder, err = m.Binance.LimitOrder(
				swapAction.SwapOneSymbol,
				m.Formatter.FormatQuantity(swapPair, swapAction.StartQuantity),
				m.Formatter.FormatPrice(swapPair, swapAction.SwapOnePrice),
				"SELL",
				"GTC",
			)
		} else {
			err = errors.New("Swap type is not supported")
		}

		if err != nil {
			log.Printf("[%s] Swap error: %s", order.Symbol, err.Error())
			return
		}

		swapOneOrder = &binanceOrder
		swapAction.SwapOneExternalId = &binanceOrder.OrderId
		nowTimestamp := time.Now().Unix()
		swapAction.SwapOneTimestamp = &nowTimestamp
		swapAction.SwapOneExternalStatus = &binanceOrder.Status
		_ = m.SwapRepository.UpdateSwapAction(swapAction)
	} else {
		binanceOrder, err := m.Binance.QueryOrder(swapAction.SwapOneSymbol, *swapAction.SwapOneExternalId)
		if err != nil {
			log.Printf("[%s] Swap error: %s", order.Symbol, err.Error())
			return
		}

		if binanceOrder.IsCanceled() || binanceOrder.IsExpired() {
			swapAction.SwapOneExternalId = nil
			swapAction.SwapOneTimestamp = nil
			swapAction.SwapOneExternalStatus = nil
			_ = m.SwapRepository.UpdateSwapAction(swapAction)
			m.BalanceService.InvalidateBalanceCache(swapAction.Asset)

			return
		}

		swapOneOrder = &binanceOrder
		swapAction.SwapOneExternalStatus = &binanceOrder.Status
		_ = m.SwapRepository.UpdateSwapAction(swapAction)
	}

	// step 1

	// todo: if expired, clear and call recursively
	if !swapOneOrder.IsFilled() {
		time.Sleep(time.Second * 5)
		for {
			binanceOrder, err := m.Binance.QueryOrder(swapOneOrder.Symbol, swapOneOrder.OrderId)
			if err != nil {
				log.Printf("[%s] Swap Binance error: %s", order.Symbol, err.Error())

				continue
			}

			swapPair, err := m.SwapRepository.GetSwapPairBySymbol(binanceOrder.Symbol)

			log.Printf(
				"[%s] Swap one processing, status %s [%d], price %f, current = %f, Executed %f of %f",
				binanceOrder.Symbol,
				binanceOrder.Status,
				binanceOrder.OrderId,
				binanceOrder.Price,
				swapPair.SellPrice,
				binanceOrder.ExecutedQty,
				binanceOrder.OrigQty,
			)

			// update value, set new memory address
			swapOneOrder = &binanceOrder

			nowTimestamp := time.Now().Unix()
			swapAction.SwapOneTimestamp = &nowTimestamp
			swapAction.SwapOneExternalStatus = &binanceOrder.Status
			_ = m.SwapRepository.UpdateSwapAction(swapAction)

			if binanceOrder.IsFilled() {
				break
			}

			// todo: timeout... cancel and remove swap action...

			if binanceOrder.IsCanceled() || binanceOrder.IsExpired() {
				swapAction.SwapOneExternalStatus = &binanceOrder.Status
				swapAction.Status = ExchangeModel.SwapActionStatusCanceled
				nowTimestamp := time.Now().Unix()
				swapAction.EndTimestamp = &nowTimestamp
				swapAction.EndQuantity = &swapAction.StartQuantity
				_ = m.SwapRepository.UpdateSwapAction(swapAction)
				order.Swap = false
				_ = m.OrderRepository.Update(order)
				// invalidate balance cache
				m.BalanceService.InvalidateBalanceCache(swapAction.Asset)
				log.Printf("[%s] Swap one process cancelled, cancel all the operation!", order.Symbol)

				return
			}

			if binanceOrder.IsNew() && (time.Now().Unix()-swapAction.StartTimestamp) >= 60 {
				cancelOrder, err := m.Binance.CancelOrder(binanceOrder.Symbol, binanceOrder.OrderId)
				if err == nil {
					swapAction.SwapOneExternalStatus = &cancelOrder.Status
					swapAction.Status = ExchangeModel.SwapActionStatusCanceled
					nowTimestamp := time.Now().Unix()
					swapAction.EndTimestamp = &nowTimestamp
					swapAction.EndQuantity = &swapAction.StartQuantity
					_ = m.SwapRepository.UpdateSwapAction(swapAction)
					order.Swap = false
					_ = m.OrderRepository.Update(order)
					// invalidate balance cache
					m.BalanceService.InvalidateBalanceCache(swapAction.Asset)
					log.Printf("[%s] Swap process cancelled, couldn't be processed more than 60 seconds", order.Symbol)

					return
				}
			}

			if binanceOrder.IsPartiallyFilled() {
				time.Sleep(time.Second * 7)
			} else {
				time.Sleep(time.Second * 15)
			}
		}
	}

	assetTwo := strings.ReplaceAll(swapOneOrder.Symbol, swapAction.Asset, "")

	// step 2
	var swapTwoOrder *ExchangeModel.BinanceOrder = nil

	if swapAction.SwapTwoExternalId == nil {
		m.BalanceService.InvalidateBalanceCache(assetTwo)
		balance, _ := m.BalanceService.GetAssetBalance(assetTwo)
		// Calculate how much we earn, and sell it!
		quantity := swapOneOrder.ExecutedQty * swapOneOrder.Price

		if quantity > balance {
			quantity = balance
		}

		log.Printf("[%s] Swap two balance %s is %f, operation SELL %s", order.Symbol, assetTwo, balance, swapAction.SwapTwoSymbol)

		swapPair, err := m.SwapRepository.GetSwapPairBySymbol(swapAction.SwapTwoSymbol)
		var binanceOrder ExchangeModel.BinanceOrder

		if swapChain.IsSSB() {
			binanceOrder, err = m.Binance.LimitOrder(
				swapAction.SwapTwoSymbol,
				m.Formatter.FormatQuantity(swapPair, quantity),
				m.Formatter.FormatPrice(swapPair, swapAction.SwapTwoPrice),
				"SELL",
				"GTC",
			)
		}

		if swapChain.IsSBB() {
			binanceOrder, err = m.Binance.LimitOrder(
				swapAction.SwapTwoSymbol,
				m.Formatter.FormatQuantity(swapPair, quantity/swapAction.SwapTwoPrice),
				m.Formatter.FormatPrice(swapPair, swapAction.SwapTwoPrice),
				"BUY",
				"GTC",
			)
		}

		if err != nil {
			log.Printf("[%s] Swap error: %s", order.Symbol, err.Error())
			return
		}

		swapTwoOrder = &binanceOrder
		swapAction.SwapTwoExternalId = &binanceOrder.OrderId
		nowTimestamp := time.Now().Unix()
		swapAction.SwapTwoTimestamp = &nowTimestamp
		swapAction.SwapTwoExternalStatus = &binanceOrder.Status
		_ = m.SwapRepository.UpdateSwapAction(swapAction)
	} else {
		binanceOrder, err := m.Binance.QueryOrder(swapAction.SwapTwoSymbol, *swapAction.SwapTwoExternalId)
		if err != nil {
			log.Printf("[%s] Swap error: %s", order.Symbol, err.Error())
			return
		}

		if binanceOrder.IsCanceled() || binanceOrder.IsExpired() {
			swapAction.SwapTwoExternalId = nil
			swapAction.SwapTwoTimestamp = nil
			swapAction.SwapTwoExternalStatus = nil
			_ = m.SwapRepository.UpdateSwapAction(swapAction)
			m.BalanceService.InvalidateBalanceCache(assetTwo)

			return
		}

		swapTwoOrder = &binanceOrder
		swapAction.SwapTwoExternalStatus = &binanceOrder.Status
		_ = m.SwapRepository.UpdateSwapAction(swapAction)
	}

	// todo: if expired, clear and call recursively
	if !swapTwoOrder.IsFilled() {
		time.Sleep(time.Second * 5)
		for {
			binanceOrder, err := m.Binance.QueryOrder(swapTwoOrder.Symbol, swapTwoOrder.OrderId)
			if err != nil {
				log.Printf("[%s] Swap Binance error: %s", order.Symbol, err.Error())

				continue
			}

			swapPair, err := m.SwapRepository.GetSwapPairBySymbol(binanceOrder.Symbol)

			log.Printf(
				"[%s] Swap two processing, status %s [%d], price %f, current = %f, Executed %f of %f",
				binanceOrder.Symbol,
				binanceOrder.Status,
				binanceOrder.OrderId,
				binanceOrder.Price,
				swapPair.SellPrice,
				binanceOrder.ExecutedQty,
				binanceOrder.OrigQty,
			)

			// update value, set new memory address
			swapTwoOrder = &binanceOrder

			nowTimestamp := time.Now().Unix()
			swapAction.SwapTwoTimestamp = &nowTimestamp
			swapAction.SwapTwoExternalStatus = &binanceOrder.Status
			_ = m.SwapRepository.UpdateSwapAction(swapAction)

			if binanceOrder.IsFilled() {
				break
			}

			if binanceOrder.IsCanceled() || binanceOrder.IsExpired() {
				swapAction.SwapTwoExternalId = nil
				swapAction.SwapTwoTimestamp = nil
				swapAction.SwapTwoExternalStatus = nil
				_ = m.SwapRepository.UpdateSwapAction(swapAction)
				m.BalanceService.InvalidateBalanceCache(assetTwo)

				return
			}

			if binanceOrder.IsPartiallyFilled() {
				time.Sleep(time.Second * 7)
			} else {
				time.Sleep(time.Second * 15)
			}
		}
	}

	assetThree := strings.ReplaceAll(swapTwoOrder.Symbol, assetTwo, "")

	// step 3
	var swapThreeOrder *ExchangeModel.BinanceOrder = nil

	if swapAction.SwapThreeExternalId == nil {
		m.BalanceService.InvalidateBalanceCache(assetThree)
		balance, _ := m.BalanceService.GetAssetBalance(assetThree)
		quantity := swapTwoOrder.ExecutedQty * swapTwoOrder.Price

		if quantity > balance {
			quantity = balance
		}

		log.Printf("[%s] Swap three balance %s is %f, operation BUY %s", order.Symbol, assetThree, balance, swapAction.SwapThreeSymbol)

		swapPair, err := m.SwapRepository.GetSwapPairBySymbol(swapAction.SwapThreeSymbol)
		var binanceOrder ExchangeModel.BinanceOrder

		if swapChain.IsSSB() || swapChain.IsSBB() {
			binanceOrder, err = m.Binance.LimitOrder(
				swapAction.SwapThreeSymbol,
				m.Formatter.FormatQuantity(swapPair, quantity/swapAction.SwapThreePrice),
				m.Formatter.FormatPrice(swapPair, swapAction.SwapThreePrice),
				"BUY",
				"GTC",
			)
		} else {
			err = errors.New("Swap type is not supported")
		}

		if err != nil {
			log.Printf("[%s] Swap error: %s", order.Symbol, err.Error())
			return
		}

		swapThreeOrder = &binanceOrder
		swapAction.SwapThreeExternalId = &binanceOrder.OrderId
		nowTimestamp := time.Now().Unix()
		swapAction.SwapThreeTimestamp = &nowTimestamp
		swapAction.SwapThreeExternalStatus = &binanceOrder.Status
		_ = m.SwapRepository.UpdateSwapAction(swapAction)
	} else {
		binanceOrder, err := m.Binance.QueryOrder(swapAction.SwapThreeSymbol, *swapAction.SwapThreeExternalId)
		if err != nil {
			log.Printf("[%s] Swap error: %s", order.Symbol, err.Error())
			return
		}

		if binanceOrder.IsCanceled() || binanceOrder.IsExpired() {
			swapAction.SwapThreeExternalId = nil
			swapAction.SwapThreeTimestamp = nil
			swapAction.SwapThreeExternalStatus = nil
			_ = m.SwapRepository.UpdateSwapAction(swapAction)
			m.BalanceService.InvalidateBalanceCache(assetThree)

			return
		}

		swapThreeOrder = &binanceOrder
		swapAction.SwapThreeExternalStatus = &binanceOrder.Status
		_ = m.SwapRepository.UpdateSwapAction(swapAction)
	}

	// todo: if expired, clear and call recursively
	if !swapThreeOrder.IsFilled() {
		time.Sleep(time.Second * 5)
		for {
			binanceOrder, err := m.Binance.QueryOrder(swapThreeOrder.Symbol, swapThreeOrder.OrderId)
			if err != nil {
				log.Printf("[%s] Swap Binance error: %s", order.Symbol, err.Error())

				continue
			}

			swapPair, err := m.SwapRepository.GetSwapPairBySymbol(binanceOrder.Symbol)

			log.Printf(
				"[%s] Swap three processing, status %s [%d], price %f, current = %f, Executed %f of %f",
				binanceOrder.Symbol,
				binanceOrder.Status,
				binanceOrder.OrderId,
				binanceOrder.Price,
				swapPair.BuyPrice,
				binanceOrder.ExecutedQty,
				binanceOrder.OrigQty,
			)

			// update value, set new memory address
			swapThreeOrder = &binanceOrder

			nowTimestamp := time.Now().Unix()
			swapAction.SwapThreeTimestamp = &nowTimestamp
			swapAction.SwapThreeExternalStatus = &binanceOrder.Status
			_ = m.SwapRepository.UpdateSwapAction(swapAction)

			if binanceOrder.IsFilled() {
				break
			}

			if binanceOrder.IsCanceled() || binanceOrder.IsExpired() {
				swapAction.SwapThreeExternalId = nil
				swapAction.SwapThreeTimestamp = nil
				swapAction.SwapThreeExternalStatus = nil
				_ = m.SwapRepository.UpdateSwapAction(swapAction)
				m.BalanceService.InvalidateBalanceCache(assetThree)

				return
			}

			if binanceOrder.IsPartiallyFilled() {
				swapAction.EndQuantity = &binanceOrder.ExecutedQty
				_ = m.SwapRepository.UpdateSwapAction(swapAction)

				if (nowTimestamp-swapAction.StartTimestamp) > (3600*4) && binanceOrder.IsNearlyFilled() {
					break // Do not cancel order, but check it later...
				}
				time.Sleep(time.Second * 7)
			} else {
				time.Sleep(time.Second * 15)
			}
		}
	}

	order.Swap = false
	swapAction.Status = ExchangeModel.SwapActionStatusSuccess
	nowTimestamp := time.Now().Unix()
	swapAction.EndTimestamp = &nowTimestamp
	swapAction.SwapThreeTimestamp = &nowTimestamp
	swapAction.EndQuantity = &swapThreeOrder.ExecutedQty
	_ = m.SwapRepository.UpdateSwapAction(swapAction)
	_ = m.OrderRepository.Update(order)

	m.BalanceService.InvalidateBalanceCache(swapAction.Asset)
	balanceAfter, _ := m.BalanceService.GetAssetBalance(swapAction.Asset)

	log.Printf(
		"[%s] Swap funished, balance %s before = %f after = %f",
		order.Symbol,
		swapAction.Asset,
		balanceBefore,
		balanceAfter,
	)
}
