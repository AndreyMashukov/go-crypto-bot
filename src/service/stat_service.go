package service

import (
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
)

type StatService struct {
	ExchangeRepository *repository.ExchangeRepository
	Binance            *client.Binance
}

func (s *StatService) GetTradeStat(kLine model.KLine, cache bool, full bool) model.TradeStat {
	var tradesPastPeriod []model.Trade

	if cache {
		if kLine.TradeVolume != nil {
			// Sell
			tradesPastPeriod = append(tradesPastPeriod, model.Trade{
				Symbol:       kLine.Symbol,
				IsBuyerMaker: true,
				Quantity:     kLine.TradeVolume.SellQty,
				Price:        kLine.Close,
			})
			// Buy
			tradesPastPeriod = append(tradesPastPeriod, model.Trade{
				Symbol:       kLine.Symbol,
				IsBuyerMaker: false,
				Quantity:     kLine.TradeVolume.BuyQty,
				Price:        kLine.Close,
			})
		}
	} else {
		tradesPastPeriod = s.Binance.TradesAggregate(
			kLine.Symbol,
			1000,
			kLine.Timestamp.GetPeriodFromMinute(),
			kLine.Timestamp.GetPeriodToMinute(),
		)
	}

	sellQty := 0.00
	buyQty := 0.00
	sellVolume := 0.00
	buyVolume := 0.00
	buyCount := 0.00
	sellCount := 0.00
	buyPriceSum := 0.00
	sellPriceSum := 0.00

	for _, trade := range tradesPastPeriod {
		if trade.IsSell() {
			sellCount += 1.0
			sellQty += trade.Quantity
			sellVolume += trade.Quantity * trade.Price
			sellPriceSum += trade.Price
		}
		if trade.IsBuy() {
			buyCount += 1.0
			buyQty += trade.Quantity
			buyVolume += trade.Quantity * trade.Price
			buyPriceSum += trade.Price
		}
	}

	maxPcs := 0.00
	minPcs := 0.00
	priceChangeSpeed := s.ExchangeRepository.GetPriceChangeSpeed(kLine.Symbol, kLine.Timestamp)

	if priceChangeSpeed != nil {
		maxPcs = priceChangeSpeed.MaxChange
		minPcs = priceChangeSpeed.MinChange
	}

	orderBookModel := model.OrderBookModel{
		Symbol: kLine.Symbol,
		Asks:   make([][2]model.Number, 0),
		Bids:   make([][2]model.Number, 0),
	}

	if full {
		if cache {
			orderBookModel = s.ExchangeRepository.GetDepth(kLine.Symbol, 500)
		} else {
			if orderBookModelPointer := s.Binance.GetDepth(kLine.Symbol, 500); orderBookModelPointer != nil {
				orderBookModel = orderBookModelPointer.ToOrderBookModel(kLine.Symbol)
			}
		}
	}

	orderBookStat := orderBookModel.GetStat()

	return model.TradeStat{
		Symbol:        kLine.Symbol,
		Timestamp:     kLine.Timestamp,
		Price:         kLine.Close,
		BuyQty:        buyQty,
		SellQty:       sellQty,
		BuyVolume:     buyVolume,
		SellVolume:    sellVolume,
		AvgBuyPrice:   buyPriceSum / buyCount,
		AvgSellPrice:  sellPriceSum / sellCount,
		BuyCount:      buyCount,
		SellCount:     sellCount,
		TradeCount:    int64(len(tradesPastPeriod)),
		MaxPSC:        maxPcs,
		MinPCS:        minPcs,
		Open:          kLine.Open,
		Close:         kLine.Close,
		High:          kLine.High,
		Low:           kLine.Low,
		Volume:        kLine.Volume,
		OrderBookStat: orderBookStat,
	}
}
