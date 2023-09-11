package exchange_context

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
	BuyLowestOnly      bool
	SellHighestOnly    bool
	TradeLockMutex     sync.RWMutex
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

	currentUnixTime := time.Now().Unix()

	for _, decision := range decisions {
		if decision.Timestamp < (currentUnixTime - 5) {
			log.Printf("[%s] Decision: %s is deprecated\n", symbol, decision.StrategyName)

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
			priceSum += decision.Price
			break
		case "SELL":
			sellScore += decision.Score
			priceSum += decision.Price
			break
		case "HOLD":
			holdScore += decision.Score
			priceSum += decision.Price
			break
		}
	}

	log.Printf("[%s] Maker - H:%f, S:%f, B:%f\n", symbol, holdScore, sellScore, buyScore)

	if holdScore >= 50 {
		return
	}

	if amount == 0 {
		return
	}

	price := priceSum / amount

	if sellScore > buyScore {
		tradeLimit, err := m.ExchangeRepository.GetTradeLimit(symbol)

		if err == nil {
			order, err := m.OrderRepository.GetOpenedOrder(symbol, "BUY")
			if err == nil {
				err = m.Sell(tradeLimit, order, symbol, price, order.Quantity, sellVolume, buyVolume, smaValue)
				if err != nil {
					log.Println(err)
				}
			} else {
				log.Printf("[%s] Nothing to sell\n", symbol)
			}
		}

		return
	}

	if buyScore > sellScore {
		tradeLimit, err := m.ExchangeRepository.GetTradeLimit(symbol)

		if err == nil {
			_, err := m.OrderRepository.GetOpenedOrder(symbol, "BUY")
			if err != nil {
				quantity := m._FormatQuantity(tradeLimit, tradeLimit.USDTLimit/price)
				err = m.Buy(tradeLimit, symbol, price, quantity, sellVolume, buyVolume, smaValue)
				if err != nil {
					log.Println(err)
				}
			}
		}

		return
	}
}

func (m *MakerService) Buy(tradeLimit ExchangeModel.TradeLimit, symbol string, price float64, quantity float64, sellVolume float64, buyVolume float64, smaValue float64) error {
	if m._IsTradeLocked(symbol) {
		return errors.New(fmt.Sprintf("Operation Buy is Locked %s", symbol))
	}

	if quantity <= 0.00 {
		return errors.New(fmt.Sprintf("Available quantity is %f", quantity))
	}

	// to avoid concurrent map writes
	m._AcquireLock(symbol)
	defer m._ReleaseLock(symbol)

	marketDepth, _ := m.Binance.GetDepth(symbol)
	lowestPrice := 0.00
	finalQuantity := quantity
	finalPrice := price

	// todo: validate with Asks...
	if marketDepth != nil {
		for _, ask := range marketDepth.Asks {
			if ask[1].Value >= quantity && (0.00 == lowestPrice || ask[0].Value <= lowestPrice) {
				lowestPrice = ask[0].Value
			}
		}
	}

	if 0.00 == lowestPrice {
		if m.BuyLowestOnly {
			return errors.New(fmt.Sprintf("[%s] No ASKs on the market", symbol))
		} else {
			lowestPrice = finalPrice
		}
	}

	// apply allowed correction
	if lowestPrice > finalPrice {
		if m.BuyLowestOnly {
			return errors.New(fmt.Sprintf("[%s] Can't buy price %f the lowest price is %f", symbol, finalPrice, lowestPrice))
		} else {
			finalPrice = m._FormatPrice(tradeLimit, (finalPrice+lowestPrice)/2)
		}
	}

	// todo: check min quantity

	var order = ExchangeModel.Order{
		Symbol:     symbol,
		Quantity:   finalQuantity,
		Price:      finalPrice,
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

	binanceOrder, err := m._TryLimitOrder(order, "BUY")

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
	if m._IsTradeLocked(symbol) {
		return errors.New(fmt.Sprintf("Operation Sell is Locked %s", symbol))
	}

	m._AcquireLock(symbol)
	defer m._ReleaseLock(symbol)

	marketDepth, _ := m.Binance.GetDepth(symbol)
	highestPrice := 0.00
	finalPrice := price

	// todo: validate with Bids...
	if marketDepth != nil {
		for _, bid := range marketDepth.Bids {
			if bid[1].Value >= quantity && (0.00 == highestPrice || bid[0].Value >= highestPrice) {
				highestPrice = bid[0].Value
			}
		}
	}

	if 0.00 == highestPrice {
		return errors.New(fmt.Sprintf("[%s] No BIDs on the market", symbol))
	}

	if highestPrice < finalPrice {
		if m.SellHighestOnly {
			return errors.New(fmt.Sprintf("[%s] Can't sell price %f the highest price is %f", symbol, finalPrice, highestPrice))
		} else {
			finalPrice = m._FormatPrice(tradeLimit, (finalPrice+highestPrice)/2)
		}
	}

	// loose money control
	if opened.Price >= finalPrice {
		return errors.New(fmt.Sprintf("[%s] Bad deal, wait for positive profit (%.6f)", symbol, (finalPrice-opened.Price)*quantity))
	}

	var order = ExchangeModel.Order{
		Symbol:     symbol,
		Quantity:   quantity,
		Price:      finalPrice,
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

	binanceOrder, err := m._TryLimitOrder(order, "SELL")

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
func (m *MakerService) _TryLimitOrder(order ExchangeModel.Order, operation string) (ExchangeModel.BinanceOrder, error) {
	binanceOrder, err := m._FindOrCreateOrder(order, operation)

	if err != nil {
		return binanceOrder, err
	}

	binanceOrder, err = m._WaitExecution(binanceOrder, 15)

	if err != nil {
		return binanceOrder, err
	}

	return binanceOrder, nil
}

func (m *MakerService) _WaitExecution(binanceOrder ExchangeModel.BinanceOrder, seconds int) (ExchangeModel.BinanceOrder, error) {
	for i := 0; i <= seconds; i++ {
		queryOrder, err := m.Binance.QueryOrder(binanceOrder.Symbol, binanceOrder.OrderId)
		log.Printf("[%s] Wait order execution %d", binanceOrder.Symbol, binanceOrder.OrderId)

		if err == nil && queryOrder.Status == "FILLED" {
			log.Printf("Order [%d] is executed [%s]", queryOrder.OrderId, queryOrder.Status)

			return queryOrder, nil
		}

		time.Sleep(time.Second)
	}

	cancelOrder, err := m.Binance.CancelOrder(binanceOrder.Symbol, binanceOrder.OrderId)

	if err != nil {
		log.Println(err)
		return binanceOrder, err
	}

	log.Printf("Order [%d] is cancelled [%s]", cancelOrder.OrderId, cancelOrder.Status)

	return cancelOrder, errors.New(fmt.Sprintf("Order %d was cancelled", binanceOrder.OrderId))
}

func (m *MakerService) _IsTradeLocked(symbol string) bool {
	m.TradeLockMutex.Lock()
	isLocked, _ := m.Lock[symbol]
	m.TradeLockMutex.Unlock()

	return isLocked
}

func (m *MakerService) _AcquireLock(symbol string) {
	*m.LockChannel <- ExchangeModel.Lock{IsLocked: true, Symbol: symbol}
}

func (m *MakerService) _ReleaseLock(symbol string) {
	*m.LockChannel <- ExchangeModel.Lock{IsLocked: false, Symbol: symbol}
}

func (m *MakerService) _FindOrCreateOrder(order ExchangeModel.Order, operation string) (ExchangeModel.BinanceOrder, error) {
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
	log.Printf("[%s] %s Order created %d", order.Symbol, operation, binanceOrder.OrderId)

	if err != nil {
		return binanceOrder, err
	}

	return binanceOrder, nil
}

func (m *MakerService) _GetLimit(symbol string) *ExchangeModel.TradeLimit {
	tradeLimits := m.ExchangeRepository.GetTradeLimits()
	for _, tradeLimit := range tradeLimits {
		if tradeLimit.Symbol == symbol {
			return &tradeLimit
		}
	}

	return nil
}

func (m *MakerService) _FormatPrice(limit ExchangeModel.TradeLimit, price float64) float64 {
	if price < limit.MinPrice {
		return limit.MinPrice
	}

	split := strings.Split(fmt.Sprintf("%s", strconv.FormatFloat(limit.MinPrice, 'f', -1, 64)), ".")
	precision := len(split[1])
	ratio := math.Pow(10, float64(precision))
	return math.Round(price*ratio) / ratio
}

func (m *MakerService) _FormatQuantity(limit ExchangeModel.TradeLimit, quantity float64) float64 {
	if quantity < limit.MinQuantity {
		return limit.MinQuantity
	}

	split := strings.Split(fmt.Sprintf("%s", strconv.FormatFloat(limit.MinQuantity, 'f', -1, 64)), ".")
	precision := len(split[1])
	ratio := math.Pow(10, float64(precision))
	return math.Round(quantity*ratio) / ratio
}
