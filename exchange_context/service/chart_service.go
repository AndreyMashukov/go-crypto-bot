package service

import (
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"slices"
	"strings"
	"time"
)

type ChartService struct {
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	OrderRepository    *ExchangeRepository.OrderRepository
}

type ChartResult struct {
	Symbol string
	Charts map[string][]any
}

func (e *ChartService) GetCharts(symbolFilter []string) []map[string][]any {
	orders := e.OrderRepository.GetList()
	orderMap := make(map[string][]model.Order)
	symbols := make([]string, 0)

	tradeLimits := e.ExchangeRepository.GetTradeLimits()

	for _, tradeLimit := range tradeLimits {
		if !tradeLimit.IsEnabled {
			continue
		}

		if len(symbolFilter) > 0 && !slices.Contains(symbolFilter, tradeLimit.Symbol) {
			continue
		}

		symbols = append(symbols, tradeLimit.Symbol)
	}

	for _, order := range orders {
		_, exist := orderMap[order.Symbol]
		if !exist {
			orderMap[order.Symbol] = make([]model.Order, 0)
		}
		orderMap[order.Symbol] = append(orderMap[order.Symbol], order)
	}

	resultChannel := make(chan ChartResult)

	for _, symbol := range symbols {
		go func(symbol string, orderMap map[string][]model.Order) {
			resultChannel <- ChartResult{
				Symbol: symbol,
				Charts: e.processSymbol(symbol, orderMap),
			}
		}(symbol, orderMap)
	}

	charts := make([]map[string][]any, 0)
	mapped := make(map[string]map[string][]any)
	processed := 0

	for {
		result := <-resultChannel
		mapped[result.Symbol] = result.Charts
		processed++

		if processed == len(symbols) {
			break
		}
	}

	for _, symbol := range symbols {
		charts = append(charts, mapped[symbol])
	}

	return charts
}

func (e *ChartService) processSymbol(symbol string, orderMap map[string][]model.Order) map[string][]any {
	list := make(map[string][]any, 0)
	symbolOrders := orderMap[symbol]
	kLines := e.ExchangeRepository.KLineList(symbol, true, 200)

	previousPredict := 0.00

	for kLineIndex, kLine := range kLines {
		klinePoint := model.FinancialPoint{
			XAxis: kLine.Timestamp,
			High:  kLine.High,
			Close: kLine.Close,
			Open:  kLine.Open,
			Low:   kLine.Low,
		}
		kLinePredict, _ := e.ExchangeRepository.GetKLinePredict(kLine)

		if kLinePredict > 0.00 {
			previousPredict = kLinePredict
		} else {
			kLinePredict = previousPredict
		}

		kLinePredictPoint := model.ChartPoint{
			XAxis: kLine.Timestamp,
			YAxis: kLinePredict,
		}
		openedBuyPoint := model.ChartPoint{
			XAxis: kLine.Timestamp,
			YAxis: 0,
		}
		sellPoint := model.ChartPoint{
			XAxis: kLine.Timestamp,
			YAxis: 0,
		}
		buyPoint := model.ChartPoint{
			XAxis: kLine.Timestamp,
			YAxis: 0,
		}
		sellPendingPoint := model.ChartPoint{
			XAxis: kLine.Timestamp,
			YAxis: 0,
		}
		buyPendingPoint := model.ChartPoint{
			XAxis: kLine.Timestamp,
			YAxis: 0,
		}

		// todo: add current sell and buy limit orders...

		for _, symbolOrder := range symbolOrders {
			date, _ := time.Parse("2006-01-02 15:04:05", symbolOrder.CreatedAt)
			orderTimestamp := date.UnixMilli() // convert date to timestamp

			if orderTimestamp >= kLine.Timestamp && len(kLines) > kLineIndex+1 && orderTimestamp < kLines[kLineIndex+1].Timestamp {
				if strings.ToUpper(symbolOrder.Operation) == "BUY" {
					buyPoint.YAxis = symbolOrder.Price
				} else {
					sellPoint.YAxis = symbolOrder.Price
				}
			}
		}

		openedBuyOrder, err := e.OrderRepository.GetOpenedOrderCached(symbol, "BUY")
		if err == nil {
			date, _ := time.Parse("2006-01-02 15:04:05", openedBuyOrder.CreatedAt)
			openedOrderTimestamp := date.UnixMilli() // convert date to timestamp
			if openedOrderTimestamp <= kLine.Timestamp {
				openedBuyPoint.YAxis = openedBuyOrder.Price
			}
		}

		binanceBuyOrder := e.OrderRepository.GetBinanceOrder(symbol, "BUY")
		if binanceBuyOrder != nil {
			buyPendingPoint.YAxis = binanceBuyOrder.Price
		}

		binanceSellOrder := e.OrderRepository.GetBinanceOrder(symbol, "SELL")
		if binanceSellOrder != nil {
			sellPendingPoint.YAxis = binanceSellOrder.Price
		}

		klineKey := fmt.Sprintf("kline-%s", symbol)
		klinePredictKey := fmt.Sprintf("predict-%s", symbol)
		orderBuyKey := fmt.Sprintf("order-buy-%s", symbol)
		orderSellKey := fmt.Sprintf("order-sell-%s", symbol)
		orderBuyPendingKey := fmt.Sprintf("order-buy-pending-%s", symbol)
		orderSellPendingKey := fmt.Sprintf("order-sell-pending-%s", symbol)
		openedOrderBuyKey := fmt.Sprintf("order-buy-opened-%s", symbol)
		list[klineKey] = append(list[klineKey], klinePoint)
		list[klinePredictKey] = append(list[klinePredictKey], kLinePredictPoint)
		list[orderBuyKey] = append(list[orderBuyKey], buyPoint)
		list[orderSellKey] = append(list[orderSellKey], sellPoint)
		list[orderBuyPendingKey] = append(list[orderBuyPendingKey], buyPendingPoint)
		list[orderSellPendingKey] = append(list[orderSellPendingKey], sellPendingPoint)
		list[openedOrderBuyKey] = append(list[openedOrderBuyKey], openedBuyPoint)
	}

	return list
}
