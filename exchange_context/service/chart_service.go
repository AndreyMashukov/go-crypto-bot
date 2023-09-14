package service

import (
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"time"
)

type ChartService struct {
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	OrderRepository    *ExchangeRepository.OrderRepository
}

func (e *ChartService) GetChart() map[string][]model.ChartPoint {
	list := make(map[string][]model.ChartPoint, 0)

	orders := e.OrderRepository.GetList()
	orderMap := make(map[string][]model.Order)
	symbols := make([]string, 0)

	for _, order := range orders {
		_, exist := orderMap[order.Symbol]
		if !exist {
			orderMap[order.Symbol] = make([]model.Order, 0)
			symbols = append(symbols, order.Symbol)
		}
		orderMap[order.Symbol] = append(orderMap[order.Symbol], order)
	}

	for _, symbol := range symbols {
		symbolOrders := orderMap[symbol]
		kLines := e.ExchangeRepository.KLineList(symbol, true)

		for kLineIndex, kLine := range kLines {
			klinePoint := model.ChartPoint{
				XAxis: kLine.Timestamp,
				YAxis: kLine.Close,
			}
			orderPoint := model.ChartPoint{
				XAxis: kLine.Timestamp,
				YAxis: 0,
			}

			for _, symbolOrder := range symbolOrders {
				date, _ := time.Parse("2006-01-02 15:04:05", symbolOrder.CreatedAt)
				orderTimestamp := date.UnixMilli() // convert date to timestamp

				if orderTimestamp >= kLine.Timestamp && len(kLines) > kLineIndex && orderTimestamp <= kLines[kLineIndex+1].Timestamp {
					orderPoint.YAxis = symbolOrder.Price
				}
			}

			klineKey := fmt.Sprintf("kline-%s", symbol)
			orderKey := fmt.Sprintf("order-%s", symbol)
			list[klineKey] = append(list[klineKey], klinePoint)
			list[orderKey] = append(list[orderKey], orderPoint)
		}
	}

	return list
}
