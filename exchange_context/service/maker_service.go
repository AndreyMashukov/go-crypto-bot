package service

import (
	"errors"
	"fmt"
	ExchangeClient "gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"log"
	"strings"
	"sync"
	"time"
)

type MakerService struct {
	OrderRepository    *ExchangeRepository.OrderRepository
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	Binance            *ExchangeClient.Binance
	LockChannel        *chan ExchangeModel.Lock
	Formatter          *Formatter
	FrameService       *FrameService
	Lock               map[string]bool
	TradeLockMutex     sync.RWMutex
	MinDecisions       float64
	HoldScore          float64
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
		tradeLimit, err := m.ExchangeRepository.GetTradeLimit(symbol)

		marketDepth := m.GetDepth(tradeLimit.Symbol)

		if len(marketDepth.Asks) < 3 && manualOrder == nil {
			log.Printf("[%s] Too small ASKs amount: %d\n", symbol, len(marketDepth.Asks))
			return
		}

		lastKline := m.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)
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
					log.Printf("[%s] SELL QTY = %f", order.Symbol, quantity)
					err = m.Sell(tradeLimit, order, symbol, price, quantity, sellVolume, buyVolume, smaFormatted)
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
					quantity := m.Formatter.FormatQuantity(tradeLimit, tradeLimit.USDTLimit/price)

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

				if err == nil && profit <= 0.00 {
					err = m.BuyExtra(tradeLimit, order, price, sellVolume, buyVolume, smaValue)
					if err != nil {
						log.Printf("[%s] %s", symbol, err)
					}
				}
			}
		}
	}
}

func (m *MakerService) calculateSellQuantity(order ExchangeModel.Order) float64 {
	m.recoverCommission(order)
	sellQuantity := order.Quantity

	// Decrease QTY by commission amount
	if order.Commission != nil && order.GetAsset() == *order.CommissionAsset {
		sellQuantity = sellQuantity - *order.Commission
	}

	balance, err := m.getAssetBalance(order.GetAsset())

	if err != nil {
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

	minPrice := m.Formatter.FormatPrice(tradeLimit, order.Price*(100+tradeLimit.MinProfitPercent)/100)
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

	profit := (minPrice - order.Price) * order.Quantity

	log.Printf("[%s] Sell price = %f, expected profit = %f$", order.Symbol, minPrice, profit)

	return m.Formatter.FormatPrice(tradeLimit, minPrice)
}

func (m *MakerService) CalculateBuyPrice(tradeLimit ExchangeModel.TradeLimit) (float64, error) {
	marketDepth := m.GetDepth(tradeLimit.Symbol)
	lastKline := m.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)

	if lastKline == nil {
		return 0.00, errors.New(fmt.Sprintf("[%s] Current price is unknown, wait...", tradeLimit.Symbol))
	}

	minPrice := m.ExchangeRepository.GetPeriodMinPrice(tradeLimit.Symbol, 200)
	order, err := m.OrderRepository.GetOpenedOrderCached(tradeLimit.Symbol, "BUY")

	// Extra charge by current price
	if err == nil && order.GetProfitPercent(lastKline.Close) <= tradeLimit.BuyOnFallPercent {
		extraBuyPrice := minPrice
		// todo: For next extras has to be used last extra order
		if order.GetHoursOpened() >= 24 {
			extraBuyPrice = lastKline.Close
			log.Printf(
				"[%s] Extra buy price is %f (more than 24 hours), profit: %.2f",
				tradeLimit.Symbol,
				extraBuyPrice,
				order.GetProfitPercent(lastKline.Close),
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
			closePrice := potentialOpenPrice * (100 + tradeLimit.MinProfitPercent) / 100

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
	if tradeLimit.BuyOnFallPercent >= 0.00 {
		return errors.New(fmt.Sprintf("[%s] Extra buy is disabled", tradeLimit.Symbol))
	}

	availableExtraBudget := tradeLimit.USDTExtraBudget - order.UsedExtraBudget

	if availableExtraBudget <= 0.00 {
		return errors.New(fmt.Sprintf("[%s] Not enough budget to buy more", tradeLimit.Symbol))
	}

	profit, _ := m.getCurrentProfitPercent(order)

	if profit > tradeLimit.BuyOnFallPercent {
		return errors.New(fmt.Sprintf("[%s] Extra buy percent is not reached %.2f of %.2f", tradeLimit.Symbol, profit, tradeLimit.BuyOnFallPercent))
	}

	m.acquireLock(order.Symbol)
	defer m.releaseLock(order.Symbol)
	quantity := m.Formatter.FormatQuantity(tradeLimit, availableExtraBudget/price)

	// todo: check min quantity

	var extraOrder = ExchangeModel.Order{
		Symbol:     order.Symbol,
		Quantity:   quantity,
		Price:      price,
		CreatedAt:  time.Now().Format("2006-01-02 15:04:05"),
		SellVolume: sellVolume,
		BuyVolume:  buyVolume,
		SmaValue:   smaValue,
		Status:     "closed",
		Operation:  "buy",
		ExternalId: nil,
		ClosedBy:   nil,
		// todo: add commission???
	}

	balanceBefore, balanceErr := m.getAssetBalance(order.GetAsset())

	binanceOrder, err := m.tryLimitOrder(extraOrder, "BUY", 120)

	if err != nil {
		return err
	}

	executedQty := binanceOrder.ExecutedQty

	// fill from API
	extraOrder.ExternalId = &binanceOrder.OrderId
	extraOrder.Quantity = executedQty
	extraOrder.Price = binanceOrder.Price
	extraOrder.ClosedBy = &order.Id
	extraOrder.CreatedAt = time.Now().Format("2006-01-02 15:04:05")

	avgPrice := m.getAvgPrice(order, extraOrder)

	// change QTY to zero for extra order
	extraOrder.Quantity = 0.00

	_, err = m.OrderRepository.Create(extraOrder)
	if err != nil {
		return err
	}

	if balanceErr == nil {
		m.UpdateCommission(balanceBefore, extraOrder)
	}

	order.Quantity = executedQty + order.Quantity
	order.Price = avgPrice
	order.UsedExtraBudget = order.UsedExtraBudget + (extraOrder.Price * executedQty)
	commission := 0.00
	if order.Commission != nil {
		commission = *order.Commission
	}
	commissionSum := commission + *extraOrder.Commission
	order.Commission = &commissionSum

	err = m.OrderRepository.Update(order)
	if err != nil {
		return err
	}

	m.OrderRepository.DeleteManualOrder(order.Symbol)

	return nil
}

func (m *MakerService) getCurrentProfitPercent(order ExchangeModel.Order) (float64, error) {
	lastKline := m.ExchangeRepository.GetLastKLine(order.Symbol)

	if lastKline == nil {
		return 0.00, errors.New(fmt.Sprintf("[%s] Do not have info about the price", order.Symbol))
	}

	return order.GetProfitPercent(lastKline.Close), nil
}

func (m *MakerService) getAvgPrice(opened ExchangeModel.Order, extra ExchangeModel.Order) float64 {
	return ((opened.Quantity * opened.Price) + (extra.Quantity * extra.Price)) / (opened.Quantity + extra.Quantity)
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

	cached := m.OrderRepository.GetBinanceOrder(symbol, "BUY")

	// Check balance for new order
	if cached == nil {
		usdtAvailableBalance, err := m.getAssetBalance("USDT")

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
		Symbol:     symbol,
		Quantity:   quantity,
		Price:      price,
		CreatedAt:  time.Now().Format("2006-01-02 15:04:05"),
		SellVolume: sellVolume,
		BuyVolume:  buyVolume,
		SmaValue:   smaValue,
		Status:     "opened",
		Operation:  "buy",
		ExternalId: nil,
		ClosedBy:   nil,
		// todo: add commission???
	}

	balanceBefore, balanceErr := m.getAssetBalance(order.GetAsset())

	binanceOrder, err := m.tryLimitOrder(order, "BUY", 480)

	if err != nil {
		return err
	}

	// fill from API
	order.ExternalId = &binanceOrder.OrderId
	order.Quantity = binanceOrder.ExecutedQty
	order.Price = binanceOrder.Price
	order.CreatedAt = time.Now().Format("2006-01-02 15:04:05")

	_, err = m.OrderRepository.Create(order)

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

	minPrice := m.Formatter.FormatPrice(tradeLimit, opened.Price*(100+tradeLimit.MinProfitPercent)/100)

	if price < minPrice {
		return errors.New(fmt.Sprintf(
			"[%s] Minimum profit is not reached, Price %.6f < %.6f",
			symbol,
			price,
			minPrice,
		))
	}

	var order = ExchangeModel.Order{
		Symbol:     symbol,
		Quantity:   quantity,
		Price:      price,
		CreatedAt:  time.Now().Format("2006-01-02 15:04:05"),
		SellVolume: sellVolume,
		BuyVolume:  buyVolume,
		SmaValue:   smaValue,
		Status:     "closed",
		Operation:  "sell",
		ExternalId: nil,
		ClosedBy:   nil,
		// todo: add commission???
	}

	//balanceBefore, balanceErr := m.getAssetBalance(order.GetAsset())

	binanceOrder, err := m.tryLimitOrder(order, "SELL", 480)

	if err != nil {
		return err
	}

	// fill from API
	order.ExternalId = &binanceOrder.OrderId
	order.Quantity = binanceOrder.ExecutedQty
	order.Price = binanceOrder.Price
	order.CreatedAt = time.Now().Format("2006-01-02 15:04:05")

	lastId, err := m.OrderRepository.Create(order)

	if err != nil {
		log.Printf("Can't create order: %s", order)

		return err
	}

	m.OrderRepository.DeleteManualOrder(order.Symbol)
	created, err := m.OrderRepository.Find(*lastId)

	if err != nil {
		log.Printf("Can't get created order [%d]: %s", lastId, order)

		return err
	}

	opened.Status = "closed"
	opened.ClosedBy = &created.Id
	err = m.OrderRepository.Update(opened)

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

	start := time.Now().Unix()
	sleepDuration := time.Second * 20

	executedQty := 0.00
	for {
		queryOrder, err := m.Binance.QueryOrder(binanceOrder.Symbol, binanceOrder.OrderId)
		if err != nil {
			log.Printf("[%s] QueryOrder: %s", binanceOrder.Symbol, err.Error())

			if strings.Contains(err.Error(), "Order was canceled or expired") {
				m.OrderRepository.DeleteBinanceOrder(queryOrder)
				break
			}

			if strings.Contains(err.Error(), "Order does not exist") {
				m.OrderRepository.DeleteBinanceOrder(queryOrder)
				break
			}

			log.Printf("[%s] Retry query order...", binanceOrder.Symbol)
			time.Sleep(sleepDuration)

			continue
		}

		log.Printf(
			"[%s] Wait %s [%.6f] order execution %d, current status is: [%s], ExecutedQty: %.6f of %.6f",
			binanceOrder.Symbol,
			binanceOrder.Side,
			binanceOrder.Price,
			binanceOrder.OrderId,
			queryOrder.Status,
			executedQty,
			queryOrder.OrigQty,
		)

		end := time.Now().Unix()

		if err == nil && queryOrder.Status == "PARTIALLY_FILLED" {
			// Add 5 minutes more if ExecutedQty moves up!
			if queryOrder.ExecutedQty > executedQty {
				seconds = seconds + (60 * 5)
			}

			executedQty = queryOrder.ExecutedQty
			m.OrderRepository.SetBinanceOrder(queryOrder)

			if (end - start) > seconds {
				break
			}

			time.Sleep(sleepDuration)
			continue
		}

		if err == nil && queryOrder.Status == "EXPIRED" {
			m.OrderRepository.DeleteBinanceOrder(queryOrder)

			break
		}

		if err == nil && queryOrder.Status == "CANCELED" {
			m.OrderRepository.DeleteBinanceOrder(queryOrder)

			break
		}

		// todo: handle EXPIRED status...

		if err == nil && queryOrder.Status == "FILLED" {
			log.Printf("[%s] Order [%d] is executed [%s]", binanceOrder.Symbol, queryOrder.OrderId, queryOrder.Status)

			m.OrderRepository.DeleteBinanceOrder(queryOrder)
			return queryOrder, nil
		}

		manualOrder := m.OrderRepository.GetManualOrder(queryOrder.Symbol)
		// cancel current immediately on new manual order
		if manualOrder != nil && manualOrder.Price != queryOrder.Price {
			break
		}

		depth := m.GetDepth(binanceOrder.Symbol)

		var bookPosition int
		var book [2]ExchangeModel.Number
		if "BUY" == binanceOrder.Side {
			bookPosition, book = depth.GetBidPosition(binanceOrder.Price)
		} else {
			bookPosition, book = depth.GetAskPosition(binanceOrder.Price)
		}

		if bookPosition < currentPosition {
			seconds += seconds
			log.Printf(
				"[%s] Order Book position decrease [%d]->[%d] %.6f!!! Ttl has extended\n",
				binanceOrder.Symbol,
				currentPosition,
				bookPosition,
				book[0],
			)
			currentPosition = bookPosition
		}

		if (end - start) > seconds {
			break
		}

		time.Sleep(sleepDuration)
	}

	cancelOrder, err := m.Binance.CancelOrder(binanceOrder.Symbol, binanceOrder.OrderId)
	m.OrderRepository.DeleteBinanceOrder(binanceOrder)

	if err != nil {
		log.Printf("[%s] Cancel: %s", binanceOrder.Symbol, err.Error())
		return binanceOrder, err
	}

	// handle cancel error and get again

	log.Printf("Order [%d] is CANCELED [%s]", cancelOrder.OrderId, cancelOrder.Status)

	return cancelOrder, errors.New(fmt.Sprintf("Order %d was CANCELED", binanceOrder.OrderId))
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

func (m *MakerService) findOrCreateOrder(order ExchangeModel.Order, operation string) (ExchangeModel.BinanceOrder, error) {
	// todo: extra order flag...
	cached := m.OrderRepository.GetBinanceOrder(order.Symbol, operation)

	if cached != nil {
		log.Printf("[%s] Found cached %s order %d in binance", order.Symbol, operation, cached.OrderId)

		return *cached, nil
	}

	openedOrders, err := m.Binance.GetOpenedOrders()

	if err != nil {
		log.Printf("[%s] Opened: %s", order.Symbol, err.Error())
		return ExchangeModel.BinanceOrder{}, err
	}

	for _, opened := range *openedOrders {
		if opened.Side == operation && opened.Symbol == order.Symbol {
			log.Printf("[%s] Found opened %s order %d in binance", order.Symbol, operation, opened.OrderId)
			m.OrderRepository.SetBinanceOrder(opened)

			return opened, nil
		}
	}

	binanceOrder, err := m.Binance.LimitOrder(order, operation)

	if err != nil {
		log.Printf("[%s] Limit: %s", order.Symbol, err.Error())
		return binanceOrder, err
	}

	log.Printf("[%s] %s Order created %d, Price: %.6f", order.Symbol, operation, binanceOrder.OrderId, binanceOrder.Price)
	m.OrderRepository.SetBinanceOrder(binanceOrder)

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
	return m.ExchangeRepository.GetDepth(symbol)
}

func (m *MakerService) UpdateLimits() {
	symbols := make([]string, 0)

	tradeLimits := m.ExchangeRepository.GetTradeLimits()
	limitMap := make(map[string]ExchangeModel.TradeLimit)
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

func (m *MakerService) getAssetBalance(asset string) (float64, error) {
	accountInfo, err := m.Binance.GetAccountStatus()

	if err != nil {
		return 0.00, err
	}

	for _, assetBalance := range accountInfo.Balances {
		if assetBalance.Asset == asset {
			log.Printf("[%s] Free balance is: %f", asset, assetBalance.Free)
			log.Printf("[%s] Locked balance is: %f", asset, assetBalance.Locked)
			return assetBalance.Free, nil
		}
	}

	return 0.00, nil
}

func (m *MakerService) recoverCommission(order ExchangeModel.Order) {
	if order.Commission != nil {
		return
	}
	assetSymbol := order.GetAsset()

	balanceAfter, err := m.getAssetBalance(assetSymbol)

	if err != nil {
		log.Printf("[%s] Can't recover commission: %s", order.Status, err.Error())
		return
	}

	commission := order.Quantity - balanceAfter
	order.Commission = &commission
	order.CommissionAsset = &assetSymbol

	err = m.OrderRepository.Update(order)
	if err != nil {
		log.Printf("[%s] Order Commission Recover: %s", order.Symbol, err.Error())
	}
}

func (m *MakerService) UpdateCommission(balanceBefore float64, order ExchangeModel.Order) {
	assetSymbol := order.GetAsset()
	balanceAfter, err := m.getAssetBalance(assetSymbol)

	if err != nil {
		log.Printf("[%s] Can't update commission: %s", order.Status, err.Error())
		return
	}

	arrived := balanceAfter - balanceBefore

	commission := order.Quantity - arrived
	order.Commission = &commission
	order.CommissionAsset = &assetSymbol

	err = m.OrderRepository.Update(order)
	if err != nil {
		log.Printf("[%s] Order Commission Update: %s", order.Symbol, err.Error())
	}
}
