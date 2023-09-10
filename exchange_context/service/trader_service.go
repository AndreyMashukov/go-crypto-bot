package exchange_context

import (
	"errors"
	"fmt"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"log"
	"math"
	"time"
)

type TraderService struct {
	OrderRepository    *ExchangeRepository.OrderRepository
	ExchangeRepository *ExchangeRepository.ExchangeRepository

	trades map[string][]ExchangeModel.Trade
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
	if t.trades == nil {
		t.trades = make(map[string][]ExchangeModel.Trade)
	}

	t.trades[trade.Symbol] = append(t.trades[trade.Symbol], trade)
	sellPeriod := 15
	buyPeriod := 60
	maxPeriod := int(math.Max(float64(sellPeriod), float64(buyPeriod)))

	if len(t.trades[trade.Symbol]) < maxPeriod {
		return
	}

	tradeSlice := t.trades[trade.Symbol][len(t.trades[trade.Symbol])-maxPeriod:]
	t.trades[trade.Symbol] = tradeSlice // override to avoid memory leaks

	sellSma := t.CalculateSMA(tradeSlice[len(tradeSlice)-sellPeriod:])
	buySma := t.CalculateSMA(tradeSlice[len(tradeSlice)-buyPeriod:])

	buyVolumeS, sellVolumeS := t.GetByAndSellVolume(tradeSlice[len(tradeSlice)-sellPeriod:])
	buyVolumeB, sellVolumeB := t.GetByAndSellVolume(tradeSlice[len(tradeSlice)-buyPeriod:])
	logFunction := func() {
		log.Printf("Sell SMA: %f\n", sellSma)
		log.Printf("Buy SMA: %f\n", buySma)
		log.Printf("Buy Volume S: %f, Sell Volume S: %f\n", buyVolumeS, sellVolumeS)
		log.Printf("Buy Volume B: %f, Sell Volume B: %f\n", buyVolumeB, sellVolumeB)
	}
	buyIndicator := buyVolumeB / sellVolumeB

	if buyIndicator > 50 && buySma < trade.Price {
		logFunction()
		log.Println("BUY!!!")
		order, err := t.OrderRepository.GetOpenedOrder(trade.Symbol, "buy")

		if err != nil {
			fmt.Println(err)
			quantity := t.GetBuyQuantity(trade)
			err = t.Buy(trade, quantity, sellVolumeB, buyVolumeB, buySma)
			if err != nil {
				log.Println(err)
			}
		} else {
			fmt.Println(order)
		}
	}

	sellIndicator := sellVolumeS / buyVolumeS

	if sellIndicator > 5 && sellSma > trade.Price {
		logFunction()
		log.Println("SELL!!!")

		order, err := t.OrderRepository.GetOpenedOrder(trade.Symbol, "buy")

		if err != nil {
			fmt.Println(err)
		} else {
			err = t.Sell(order, trade, order.Quantity, sellVolumeS, buyVolumeS, sellSma)

			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (t *TraderService) Buy(trade ExchangeModel.Trade, quantity float64, sellVolume float64, buyVolume float64, smaValue float64) error {
	if quantity <= 0.00 {
		return errors.New(fmt.Sprintf("Available quantity is %f", quantity))
	}

	// todo: check min quantity

	var order = ExchangeModel.Order{
		Symbol:     trade.Symbol,
		Quantity:   quantity,
		Price:      trade.Price,
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

	// todo: Call API and Create() only on success!

	_, err := t.OrderRepository.Create(order)

	if err != nil {
		log.Printf("Can't create order: %s", order)

		return err
	}

	return nil
}

func (t *TraderService) Sell(opened ExchangeModel.Order, trade ExchangeModel.Trade, quantity float64, sellVolume float64, buyVolume float64, smaValue float64) error {
	if opened.Price >= trade.Price {
		return errors.New("bad deal, wait for positive profit")
	}

	var order = ExchangeModel.Order{
		Symbol:     trade.Symbol,
		Quantity:   quantity,
		Price:      trade.Price,
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

	// todo: Call API and Create() only on success!

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

func (t *TraderService) GetBuyQuantity(trade ExchangeModel.Trade) float64 {
	limit := 0.00

	// todo: use cache
	tradeLimits := t.ExchangeRepository.GetTradeLimits()
	for _, tradeLimit := range tradeLimits {
		if tradeLimit.Symbol == trade.Symbol {
			limit = tradeLimit.USDTLimit
			break
		}
	}

	if limit > 0 {
		return limit / trade.Price
	}

	return 0.00
}
