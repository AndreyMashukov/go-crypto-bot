package exchange_context

import (
	"errors"
	"fmt"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"log"
	"time"
)

type TraderService struct {
	OrderRepository *ExchangeRepository.OrderRepository

	candles []ExchangeModel.Candle
}

func (t *TraderService) CalculateSMA(candles []ExchangeModel.Candle) float64 {
	var sum float64

	for _, candle := range candles {
		sum += candle.Price
	}

	return sum / float64(len(candles))
}

func (t *TraderService) GetByAndSellVolume(candles []ExchangeModel.Candle) (float64, float64) {
	var buyVolume float64
	var sellVolume float64

	for _, candle := range candles {
		switch candle.GetOperation() {
		case "BUY":
			buyVolume += candle.Price * candle.Quantity
			break
		case "SELL":
			sellVolume += candle.Price * candle.Quantity
			break
		}
	}

	return buyVolume, sellVolume
}

func (t *TraderService) Trade(candle ExchangeModel.Candle) {
	t.candles = append(t.candles, candle)
	period := 50

	if len(t.candles) < period {
		return
	}

	last := t.candles[len(t.candles)-period:]
	t.candles = last // override to avoid memory leaks

	sma := t.CalculateSMA(last)
	buyVolume, sellVolume := t.GetByAndSellVolume(last)
	logFunction := func() {
		log.Printf("SMA: %f\n", sma)
		log.Printf("Buy: %f, Sell: %f\n", buyVolume, sellVolume)
	}

	buyIndicator := buyVolume / sellVolume

	if buyIndicator > 2 && sma < candle.Price {
		logFunction()
		log.Println("BUY!!!")
		order, err := t.OrderRepository.GetOpenedOrder(candle.Symbol, "buy")

		if err != nil {
			fmt.Println(err)
			// todo: calculate quantity by available balance...
			err = t.Buy(candle, 0.2, sellVolume, buyVolume, sma)
			if err != nil {
				log.Println(err)
			}
		} else {
			fmt.Println(order)
		}
	}

	sellIndicator := sellVolume / buyVolume

	if sellIndicator > 2 && sma > candle.Price {
		logFunction()
		log.Println("SELL!!!")

		order, err := t.OrderRepository.GetOpenedOrder(candle.Symbol, "buy")

		if err != nil {
			fmt.Println(err)
		} else {
			err = t.Sell(order, candle, order.Quantity, sellVolume, buyVolume, sma)

			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (t *TraderService) Buy(candle ExchangeModel.Candle, quantity float64, sellVolume float64, buyVolume float64, smaValue float64) error {
	var order = ExchangeModel.Order{
		Symbol:     candle.Symbol,
		Quantity:   quantity,
		Price:      candle.Price,
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

func (t *TraderService) Sell(opened ExchangeModel.Order, candle ExchangeModel.Candle, quantity float64, sellVolume float64, buyVolume float64, smaValue float64) error {
	if opened.Price > candle.Price {
		return errors.New("bad deal, wait for positive profit")
	}

	var order = ExchangeModel.Order{
		Symbol:     candle.Symbol,
		Quantity:   quantity,
		Price:      candle.Price,
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
