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
	"time"
)

type TraderService struct {
	OrderRepository    *ExchangeRepository.OrderRepository
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	Binance            *ExchangeClient.Binance
	LockChannel        *chan ExchangeModel.Lock
	Lock               map[string]bool
	BuyLowestOnly      bool
	SellHighestOnly    bool
	Trades             map[string][]ExchangeModel.Trade
}

func (t *TraderService) CalculateSMA(trades []ExchangeModel.Trade) float64 {
	var sum float64

	slice := trades

	for _, trade := range slice {
		sum += trade.Price
	}

	return sum / float64(len(slice))
}

func (t *TraderService) GetByAndSellVolume(trades []ExchangeModel.Trade) (float64, float64) {
	var buyVolume float64
	var sellVolume float64

	for _, trade := range trades {
		switch trade.GetOperation() {
		case "BUY":
			buyVolume += trade.Price * trade.Quantity
			break
		case "SELL":
			sellVolume += trade.Price * trade.Quantity
			break
		}
	}

	return buyVolume, sellVolume
}

func (t *TraderService) Trade(trade ExchangeModel.Trade) {
	sellPeriod := 15
	buyPeriod := 60
	maxPeriod := int(math.Max(float64(sellPeriod), float64(buyPeriod)))

	if len(t.Trades[trade.Symbol]) < maxPeriod {
		return
	}

	tradeSlice := t.Trades[trade.Symbol][len(t.Trades[trade.Symbol])-maxPeriod:]
	t.Trades[trade.Symbol] = tradeSlice // override to avoid memory leaks

	sellSma := t.CalculateSMA(tradeSlice[len(tradeSlice)-sellPeriod:])
	buySma := t.CalculateSMA(tradeSlice[len(tradeSlice)-buyPeriod:])

	buyVolumeS, sellVolumeS := t.GetByAndSellVolume(tradeSlice[len(tradeSlice)-sellPeriod:])
	buyVolumeB, sellVolumeB := t.GetByAndSellVolume(tradeSlice[len(tradeSlice)-buyPeriod:])
	logFunction := func(trade ExchangeModel.Trade) {
		log.Printf("[%s] Sell SMA: %f\n", trade.Symbol, sellSma)
		log.Printf("[%s] Buy SMA: %f\n", trade.Symbol, buySma)
		log.Printf("[%s] Buy Volume S: %f, Sell Volume S: %f\n", trade.Symbol, buyVolumeS, sellVolumeS)
		log.Printf("[%s] Buy Volume B: %f, Sell Volume B: %f\n", trade.Symbol, buyVolumeB, sellVolumeB)
	}

	buyIndicator := buyVolumeB / sellVolumeB

	if buyIndicator > 50 && buySma < trade.Price {
		log.Printf("[%s] BUY!!!", trade.Symbol)
		tradeLimit := t._GetLimit(trade.Symbol)
		if tradeLimit != nil {
			logFunction(trade)
			order, err := t.OrderRepository.GetOpenedOrder(trade.Symbol, "buy")

			if err != nil {
				fmt.Println(err)
				quantity := t._FormatQuantity(*tradeLimit, tradeLimit.USDTLimit/trade.Price)
				log.Printf("[%s] Quantity is %f", trade.Symbol, quantity)
				err = t.Buy(*tradeLimit, trade, quantity, sellVolumeB, buyVolumeB, buySma)
				if err != nil {
					log.Println(err)
				}
			} else {
				fmt.Printf("[%s] Order already opened: %d, price: %f/%f", order.Symbol, order.Id, order.Price, order.Quantity)
			}
		} else {
			fmt.Printf("[%s] Trade limits is not configured\n", trade.Symbol)
		}
	}

	sellIndicator := sellVolumeS / buyVolumeS

	if sellIndicator > 5 && sellSma > trade.Price {
		log.Printf("[%s] SELL!!!", trade.Symbol)
		tradeLimit := t._GetLimit(trade.Symbol)
		if tradeLimit != nil {
			logFunction(trade)

			order, err := t.OrderRepository.GetOpenedOrder(trade.Symbol, "buy")

			if err != nil {
				fmt.Println(err)
			} else {
				err = t.Sell(*tradeLimit, order, trade, order.Quantity, sellVolumeS, buyVolumeS, sellSma)

				if err != nil {
					log.Println(err)
				}
			}
		} else {
			fmt.Printf("[%s] Trade limits is not configured\n", trade.Symbol)
		}
	}
}

func (t *TraderService) Buy(tradeLimit ExchangeModel.TradeLimit, trade ExchangeModel.Trade, quantity float64, sellVolume float64, buyVolume float64, smaValue float64) error {
	if t._IsTradeLocked(trade.Symbol) {
		return errors.New(fmt.Sprintf("Operation Buy is Locked %s", trade.Symbol))
	}

	if quantity <= 0.00 {
		return errors.New(fmt.Sprintf("Available quantity is %f", quantity))
	}

	// to avoid concurrent map writes
	t._AcquireLock(trade.Symbol)
	defer t._ReleaseLock(trade.Symbol)

	marketDepth, _ := t.Binance.GetDepth(trade.Symbol)
	lowestPrice := 0.00
	finalQuantity := quantity
	finalPrice := trade.Price

	// todo: validate with Asks...
	if marketDepth != nil {
		for _, ask := range marketDepth.Asks {
			if ask[1].Value >= quantity && (0.00 == lowestPrice || ask[0].Value <= lowestPrice) {
				lowestPrice = ask[0].Value
			}
		}
	}

	if 0.00 == lowestPrice {
		return errors.New(fmt.Sprintf("[%s] No ASKs on the market", trade.Symbol))
	}

	// apply allowed correction
	if lowestPrice > finalPrice {
		if t.BuyLowestOnly {
			return errors.New(fmt.Sprintf("[%s] Can't buy price %f the lowest price is %f", trade.Symbol, finalPrice, lowestPrice))
		} else {
			finalPrice = t._FormatPrice(tradeLimit, (finalPrice+lowestPrice)/2)
		}
	}

	// todo: check min quantity

	var order = ExchangeModel.Order{
		Symbol:     trade.Symbol,
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

	binanceOrder, err := t._TryLimitOrder(order, "BUY")

	if err != nil {
		return err
	}

	// fill from API
	order.ExternalId = &binanceOrder.OrderId
	order.Quantity = binanceOrder.ExecutedQty
	order.Price = binanceOrder.Price

	_, err = t.OrderRepository.Create(order)

	if err != nil {
		log.Printf("Can't create order: %s", order)

		return err
	}

	return nil
}

func (t *TraderService) Sell(tradeLimit ExchangeModel.TradeLimit, opened ExchangeModel.Order, trade ExchangeModel.Trade, quantity float64, sellVolume float64, buyVolume float64, smaValue float64) error {
	if t._IsTradeLocked(trade.Symbol) {
		return errors.New(fmt.Sprintf("Operation Sell is Locked %s", trade.Symbol))
	}

	t._AcquireLock(trade.Symbol)
	defer t._ReleaseLock(trade.Symbol)

	marketDepth, _ := t.Binance.GetDepth(trade.Symbol)
	highestPrice := 0.00
	finalPrice := trade.Price

	// todo: validate with Bids...
	if marketDepth != nil {
		for _, bid := range marketDepth.Bids {
			if bid[1].Value >= quantity && (0.00 == highestPrice || bid[0].Value >= highestPrice) {
				highestPrice = bid[0].Value
			}
		}
	}

	if 0.00 == highestPrice {
		return errors.New(fmt.Sprintf("[%s] No BIDs on the market", trade.Symbol))
	}

	if highestPrice < finalPrice {
		if t.SellHighestOnly {
			return errors.New(fmt.Sprintf("[%s] Can't sell price %f the highest price is %f", trade.Symbol, finalPrice, highestPrice))
		} else {
			finalPrice = t._FormatPrice(tradeLimit, (finalPrice+highestPrice)/2)
		}
	}

	// loose money control
	if opened.Price >= finalPrice {
		return errors.New(fmt.Sprintf("[%s] Bad deal, wait for positive profit", trade.Symbol))
	}

	var order = ExchangeModel.Order{
		Symbol:     trade.Symbol,
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

	binanceOrder, err := t._TryLimitOrder(order, "SELL")

	if err != nil {
		*t.LockChannel <- ExchangeModel.Lock{IsLocked: false, Symbol: trade.Symbol}

		return err
	}

	// fill from API
	order.ExternalId = &binanceOrder.OrderId
	order.Quantity = binanceOrder.ExecutedQty
	order.Price = binanceOrder.Price

	lastId, err := t.OrderRepository.Create(order)

	if err != nil {
		log.Printf("Can't create order: %s", order)

		return err
	}

	created, err := t.OrderRepository.Find(*lastId)

	if err != nil {
		log.Printf("Can't get created order [%d]: %s", lastId, order)

		return err
	}

	opened.Status = "closed"
	opened.ClosedBy = &created.Id
	err = t.OrderRepository.Update(opened)

	if err != nil {
		log.Printf("Can't udpdate order [%d]: %s", order.Id, order)

		return err
	}

	return nil
}

// todo: order has to be Interface
func (t *TraderService) _TryLimitOrder(order ExchangeModel.Order, operation string) (*ExchangeModel.BinanceOrder, error) {
	binanceOrder, err := t._FindOrCreateOrder(order, operation)

	if err != nil {
		return nil, err
	}

	binanceOrder, err = t._WaitExecution(binanceOrder, 15)

	if err != nil {
		return nil, err
	}

	return binanceOrder, nil
}

func (t *TraderService) _WaitExecution(binanceOrder *ExchangeModel.BinanceOrder, seconds int) (*ExchangeModel.BinanceOrder, error) {
	for i := 0; i <= seconds; i++ {
		queryOrder, err := t.Binance.QueryOrder(binanceOrder.Symbol, binanceOrder.OrderId)
		log.Printf("[%s] Wait order execution %d", binanceOrder.Symbol, binanceOrder.OrderId)

		if err == nil && queryOrder != nil && queryOrder.Status == "FILLED" {
			log.Printf("Order [%d] is executed [%s]", queryOrder.OrderId, queryOrder.Status)

			return queryOrder, nil
		}

		time.Sleep(time.Second)
	}

	cancelOrder, err := t.Binance.CancelOrder(binanceOrder.Symbol, binanceOrder.OrderId)

	if err != nil {
		log.Println(err)
		return binanceOrder, err
	}

	log.Printf("Order [%d] is cancelled [%s]", cancelOrder.OrderId, cancelOrder.Status)

	return cancelOrder, errors.New(fmt.Sprintf("Order %d was cancelled", binanceOrder.OrderId))
}

func (t *TraderService) _IsTradeLocked(symbol string) bool {
	isLocked, _ := t.Lock[symbol]

	return isLocked
}

func (t *TraderService) _AcquireLock(symbol string) {
	*t.LockChannel <- ExchangeModel.Lock{IsLocked: true, Symbol: symbol}
}

func (t *TraderService) _ReleaseLock(symbol string) {
	*t.LockChannel <- ExchangeModel.Lock{IsLocked: false, Symbol: symbol}
}

func (t *TraderService) _FindOrCreateOrder(order ExchangeModel.Order, operation string) (*ExchangeModel.BinanceOrder, error) {
	openedOrders, err := t.Binance.GetOpenedOrders()

	if err != nil {
		return nil, err
	}

	for _, opened := range *openedOrders {
		if opened.Side == operation && opened.Symbol == order.Symbol {
			log.Printf("[%s] Found opened %s order %d in binance", order.Symbol, operation, opened.OrderId)
			return &opened, nil
		}
	}

	binanceOrder, err := t.Binance.LimitOrder(order, operation)
	log.Printf("[%s] %s Order created %d", order.Symbol, operation, binanceOrder.OrderId)

	if err != nil {
		return nil, err
	}

	return binanceOrder, nil
}

func (t *TraderService) _GetLimit(symbol string) *ExchangeModel.TradeLimit {
	tradeLimits := t.ExchangeRepository.GetTradeLimits()
	for _, tradeLimit := range tradeLimits {
		if tradeLimit.Symbol == symbol {
			return &tradeLimit
		}
	}

	return nil
}

func (t *TraderService) _FormatPrice(limit ExchangeModel.TradeLimit, price float64) float64 {
	if price < limit.MinPrice {
		return limit.MinPrice
	}

	split := strings.Split(fmt.Sprintf("%s", strconv.FormatFloat(limit.MinPrice, 'f', -1, 64)), ".")
	precision := len(split[1])
	ratio := math.Pow(10, float64(precision))
	return math.Round(price*ratio) / ratio
}

func (t *TraderService) _FormatQuantity(limit ExchangeModel.TradeLimit, quantity float64) float64 {
	if quantity < limit.MinQuantity {
		return limit.MinQuantity
	}

	split := strings.Split(fmt.Sprintf("%s", strconv.FormatFloat(limit.MinQuantity, 'f', -1, 64)), ".")
	precision := len(split[1])
	ratio := math.Pow(10, float64(precision))
	return math.Round(quantity*ratio) / ratio
}
