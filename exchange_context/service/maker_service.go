package service

import (
	"errors"
	"fmt"
	ExchangeClient "gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MakerService struct {
	OrderRepository    *ExchangeRepository.OrderRepository
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	Binance            *ExchangeClient.Binance
	LockChannel        *chan ExchangeModel.Lock
	Lock               map[string]bool
	TradeLockMutex     sync.RWMutex
	MinDecisions       float64
	HoldScore          float64
	DepthMap           map[string]ExchangeModel.Depth
}

func (m *MakerService) Make(symbol string, decisions []ExchangeModel.Decision) {
	buyScore := 0.00
	sellScore := 0.00
	holdScore := 0.00

	sellVolume := 0.00
	buyVolume := 0.00
	smaValue := 0.00
	amount := 0.00

	currentUnixTime := time.Now().Unix()

	for _, decision := range decisions {
		if decision.Timestamp < (currentUnixTime - 5) {
			continue
		}

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
	}

	if amount != m.MinDecisions {
		return
	}

	//marketDepth := m.GetDepth(symbol)
	//if marketDepth != nil {
	//	log.Printf(
	//		"[%s] Bid: %.6f[%.4f] Ask: %.6f[%.4f]\n",
	//		symbol,
	//		marketDepth.GetBestBid(),
	//		marketDepth.GetBidVolume(),
	//		marketDepth.GetBestAsk(),
	//		marketDepth.GetAskVolume(),
	//	)
	//}

	if holdScore >= m.HoldScore {
		return
	}

	if sellScore >= buyScore {
		log.Printf("[%s] Maker - H:%f, S:%f, B:%f\n", symbol, holdScore, sellScore, buyScore)
		tradeLimit, err := m.ExchangeRepository.GetTradeLimit(symbol)

		if err == nil {
			order, err := m.OrderRepository.GetOpenedOrder(symbol, "BUY")
			if err == nil {
				price := m.calculateSellPrice(tradeLimit)

				if price > 0 {
					err = m.Sell(tradeLimit, order, symbol, price, order.Quantity, sellVolume, buyVolume, smaValue)
					if err != nil {
						log.Println(err)
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

		if err == nil {
			_, err := m.OrderRepository.GetOpenedOrder(symbol, "BUY")
			if err != nil {
				price := m.calculateBuyPrice(tradeLimit)

				if price > 0 {
					quantity := m.formatQuantity(tradeLimit, tradeLimit.USDTLimit/price)
					err = m.Buy(tradeLimit, symbol, price, quantity, sellVolume, buyVolume, smaValue)
					if err != nil {
						log.Println(err)
					}
				} else {
					log.Printf("[%s] No ASKs on the market", symbol)
				}
			}
		}

		return
	}
}

func (m *MakerService) calculateSellPrice(tradeLimit ExchangeModel.TradeLimit) float64 {
	marketDepth := m.GetDepth(tradeLimit.Symbol)
	bestPrice := 0.00

	if marketDepth != nil {
		bestPrice = marketDepth.GetBestAsk()
	}

	if 0.00 == bestPrice {
		return bestPrice
	}

	return m.formatPrice(tradeLimit, bestPrice*1.005) // 0.5% higher than best Bid
}

func (m *MakerService) calculateBuyPrice(tradeLimit ExchangeModel.TradeLimit) float64 {
	marketDepth := m.GetDepth(tradeLimit.Symbol)
	bestPrice := 0.00

	if marketDepth != nil {
		bestPrice = marketDepth.GetBestBid()
	}

	if 0.00 == bestPrice {
		return bestPrice
	}

	return m.formatPrice(tradeLimit, bestPrice*0.995) // 0.5% lower than best Ask
}

func (m *MakerService) BuyExtra(tradeLimit ExchangeModel.TradeLimit, order ExchangeModel.Order, price float64) error {
	// todo: buy extra
	// todo: validate extra buy
	// todo: merge new order with existing
	// todo: calculate average price for merged order
	return nil
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

	binanceOrder, err := m.tryLimitOrder(order, "BUY")

	if err != nil {
		return err
	}

	// fill from API
	order.ExternalId = &binanceOrder.OrderId
	order.Quantity = binanceOrder.ExecutedQty
	order.Price = binanceOrder.Price

	_, err = m.OrderRepository.Create(order)

	if err != nil {
		log.Printf("Can't create order: %s", order)

		return err
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

	profitPercent := (price * 100 / opened.Price) - 100

	if profitPercent < tradeLimit.MinProfitPercent {
		return errors.New(fmt.Sprintf(
			"[%s] Minimum profit is not reached: %.6f of %.6f [o:%.6f, c:%.6f]",
			symbol,
			profitPercent,
			tradeLimit.MinProfitPercent,
			opened.Price,
			price,
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

	binanceOrder, err := m.tryLimitOrder(order, "SELL")

	if err != nil {
		return err
	}

	// fill from API
	order.ExternalId = &binanceOrder.OrderId
	order.Quantity = binanceOrder.ExecutedQty
	order.Price = binanceOrder.Price

	lastId, err := m.OrderRepository.Create(order)

	if err != nil {
		log.Printf("Can't create order: %s", order)

		return err
	}

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

	return nil
}

// todo: order has to be Interface
func (m *MakerService) tryLimitOrder(order ExchangeModel.Order, operation string) (ExchangeModel.BinanceOrder, error) {
	binanceOrder, err := m.findOrCreateOrder(order, operation)

	if err != nil {
		return binanceOrder, err
	}

	binanceOrder, err = m.waitExecution(binanceOrder, 10)

	if err != nil {
		return binanceOrder, err
	}

	return binanceOrder, nil
}

func (m *MakerService) waitExecution(binanceOrder ExchangeModel.BinanceOrder, seconds int) (ExchangeModel.BinanceOrder, error) {
	for i := 0; i <= seconds; i++ {
		queryOrder, err := m.Binance.QueryOrder(binanceOrder.Symbol, binanceOrder.OrderId)
		log.Printf("[%s] Wait order execution %d, current status is [%s]", binanceOrder.Symbol, binanceOrder.OrderId, queryOrder.Status)

		if err == nil && queryOrder.Status == "PARTIALLY_FILLED" {
			time.Sleep(time.Second)
			seconds++
			continue
		}

		if err == nil && queryOrder.Status == "FILLED" {
			log.Printf("[%s] Order [%d] is executed [%s]", binanceOrder.Symbol, queryOrder.OrderId, queryOrder.Status)

			return queryOrder, nil
		}

		time.Sleep(time.Second)
	}

	cancelOrder, err := m.Binance.CancelOrder(binanceOrder.Symbol, binanceOrder.OrderId)

	if err != nil {
		log.Println(err)
		return binanceOrder, err
	}

	// handle cancel error and get again

	log.Printf("Order [%d] is cancelled [%s]", cancelOrder.OrderId, cancelOrder.Status)

	return cancelOrder, errors.New(fmt.Sprintf("Order %d was cancelled", binanceOrder.OrderId))
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
	openedOrders, err := m.Binance.GetOpenedOrders()

	if err != nil {
		return ExchangeModel.BinanceOrder{}, err
	}

	for _, opened := range *openedOrders {
		if opened.Side == operation && opened.Symbol == order.Symbol {
			log.Printf("[%s] Found opened %s order %d in binance", order.Symbol, operation, opened.OrderId)
			return opened, nil
		}
	}

	binanceOrder, err := m.Binance.LimitOrder(order, operation)

	if err != nil {
		return binanceOrder, err
	}

	log.Printf("[%s] %s Order created %d, price: %.6f", order.Symbol, operation, binanceOrder.OrderId, binanceOrder.Price)

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

func (m *MakerService) formatPrice(limit ExchangeModel.TradeLimit, price float64) float64 {
	if price < limit.MinPrice {
		return limit.MinPrice
	}

	split := strings.Split(fmt.Sprintf("%s", strconv.FormatFloat(limit.MinPrice, 'f', -1, 64)), ".")
	precision := len(split[1])
	ratio := math.Pow(10, float64(precision))
	return math.Round(price*ratio) / ratio
}

func (m *MakerService) formatQuantity(limit ExchangeModel.TradeLimit, quantity float64) float64 {
	if quantity < limit.MinQuantity {
		return limit.MinQuantity
	}

	split := strings.Split(fmt.Sprintf("%s", strconv.FormatFloat(limit.MinQuantity, 'f', -1, 64)), ".")
	precision := len(split[1])
	ratio := math.Pow(10, float64(precision))
	return math.Round(quantity*ratio) / ratio
}

func (m *MakerService) SetDepth(depth ExchangeModel.Depth) {
	m.TradeLockMutex.Lock()
	m.DepthMap[depth.Symbol] = depth
	m.TradeLockMutex.Unlock()
}

func (m *MakerService) GetDepth(symbol string) *ExchangeModel.Depth {
	m.TradeLockMutex.Lock()
	depth, exists := m.DepthMap[symbol]
	m.TradeLockMutex.Unlock()

	if !exists {
		return nil
	}

	return &depth
}
