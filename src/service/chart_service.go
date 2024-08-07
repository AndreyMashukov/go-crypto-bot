package service

import (
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"slices"
	"strings"
	"sync"
	"time"
)

type ChartService struct {
	ExchangeRepository *repository.ExchangeRepository
	OrderRepository    *repository.OrderRepository
	Formatter          *utils.Formatter
	StatRepository     *repository.StatRepository
	StatService        *StatService
}

type ChartResult struct {
	Symbol string
	Charts map[string][]any
}

func (e *ChartService) GetCharts(symbolFilter []string) []map[string][]any {
	symbols := make([]string, 0)

	tradeLimits := e.ExchangeRepository.GetTradeLimits()

	for _, tradeLimit := range tradeLimits {
		if len(symbolFilter) > 0 && !slices.Contains(symbolFilter, tradeLimit.Symbol) {
			continue
		}

		symbols = append(symbols, tradeLimit.Symbol)
	}

	resultChannel := make(chan ChartResult)

	for _, symbol := range symbols {
		go func(symbol string) {
			resultChannel <- ChartResult{
				Symbol: symbol,
				Charts: e.ProcessSymbol(symbol),
			}
		}(symbol)
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

func (e *ChartService) ProcessSymbol(symbol string) map[string][]any {
	kLines := e.ExchangeRepository.KLineList(symbol, true, 200)

	symbolOrders := make([]model.Order, 0)
	orderMap := sync.Map{}

	if len(kLines) > 1 {
		from := time.UnixMilli(kLines[0].Timestamp.GetPeriodFromMinute())
		to := time.UnixMilli(kLines[len(kLines)-1].Timestamp.GetPeriodToMinute())

		// todo: use cache for history list
		symbolOrders = e.OrderRepository.GetHistoryList(symbol, from, to)
		for _, symbolOrder := range symbolOrders {
			date, _ := time.Parse("2006-01-02 15:04:05", symbolOrder.CreatedAt)
			orderTimestamp := model.TimestampMilli(date.UnixMilli()).GetPeriodToMinute() // convert date to timestamp
			orderMap.Store(orderTimestamp, symbolOrder)
		}
	}

	tradeLimit := e.ExchangeRepository.GetTradeLimitCached(symbol)
	cummulativeTradeQuantity := 0.00
	statMap := e.StatRepository.GetStatRange(symbol, 200)

	cKLine := e.ExchangeRepository.GetCurrentKline(symbol)
	if cKLine != nil {
		cStat := e.StatService.GetTradeStat(*cKLine, true, false)
		statMap.Store(cKLine.Timestamp.GetPeriodToMinute(), cStat)
	}

	list := make(map[string][]any, 0)

	for _, kLine := range kLines {
		klinePoint := model.FinancialPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			High:  kLine.High.Value(),
			Close: kLine.Close.Value(),
			Open:  kLine.Open.Value(),
			Low:   kLine.Low.Value(),
		}
		kLinePredict, _ := e.ExchangeRepository.GetKLinePredict(kLine)
		interpolation, _ := e.ExchangeRepository.GetInterpolation(kLine)

		tradeVolume := e.ExchangeRepository.GetTradeVolume(kLine.Symbol, kLine.Timestamp)
		tradeVolumeSellVal := kLine.GetTradeVolumeSell()
		tradeVolumeBuyVal := kLine.GetTradeVolumeBuy()

		if tradeVolume != nil {
			tradeVolumeSellVal = tradeVolume.SellQty
			tradeVolumeBuyVal = tradeVolume.BuyQty
		}

		cummulativeTradeQuantity += tradeVolumeBuyVal
		cummulativeTradeQuantity -= tradeVolumeSellVal

		capitalization := 0.00
		capitalizationPrice := 0.00
		capitalizationValue := e.ExchangeRepository.GetCapitalization(kLine.Symbol, kLine.Timestamp)
		if capitalizationValue != nil {
			capitalization = e.Formatter.ToFixed(capitalizationValue.Capitalization, 2)
			capitalizationPrice = e.Formatter.FormatPrice(tradeLimit, capitalizationValue.Price)
		}

		icebergQtyBuyVal := 0.00
		icebergQtySellVal := 0.00
		icebergPriceBuyVal := 0.00
		icebergPriceSellVal := 0.00

		if tradeStatObj, ok := statMap.Load(kLine.Timestamp.GetPeriodToMinute()); ok {
			if tradeStat, ok := tradeStatObj.(model.TradeStat); ok {
				if tradeStat.OrderBookStat.BuyIceberg.Price != 0.00 {
					icebergPriceBuyVal = e.Formatter.FormatPrice(tradeLimit, tradeStat.OrderBookStat.BuyIceberg.Price)
				}
				if tradeStat.OrderBookStat.SellIceberg.Price != 0.00 {
					icebergPriceSellVal = e.Formatter.FormatPrice(tradeLimit, tradeStat.OrderBookStat.SellIceberg.Price)
				}
				if tradeStat.OrderBookStat.BuyIceberg.Quantity != 0.00 {
					icebergQtyBuyVal = e.Formatter.FormatQuantity(tradeLimit, tradeStat.OrderBookStat.BuyIceberg.Quantity)
				}
				if tradeStat.OrderBookStat.SellIceberg.Quantity != 0.00 {
					icebergQtySellVal = e.Formatter.FormatQuantity(tradeLimit, tradeStat.OrderBookStat.SellIceberg.Quantity)
				}
			}
		}

		icebergPriceBuyPoint := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: icebergPriceBuyVal,
		}
		icebergPriceSellPoint := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: icebergPriceSellVal,
		}
		icebergQtyBuyPoint := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: icebergQtyBuyVal,
		}
		icebergQtySellPoint := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: icebergQtySellVal,
		}

		capitalizationValuePoint := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: capitalization,
		}
		capitalizationPricePoint := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: capitalizationPrice,
		}
		tradeVolumeSell := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: tradeVolumeSellVal,
		}
		tradeVolumeBuy := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: tradeVolumeBuyVal,
		}
		cummulativeTradeQtyPoint := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: cummulativeTradeQuantity,
		}
		if kLinePredict > 0.00 {
			kLinePredict = e.Formatter.FormatPrice(tradeLimit, kLinePredict)
		}

		kLinePredictPoint := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: kLinePredict,
		}

		priceChangeSpeed := e.ExchangeRepository.GetPriceChangeSpeed(kLine.Symbol, kLine.Timestamp)
		if priceChangeSpeed != nil {
			kLine.PriceChangeSpeed = priceChangeSpeed
		}

		kLineAvgChangeSpeedPoint := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: e.Formatter.ToFixed(kLine.GetPriceChangeSpeedAvg(), 2),
		}
		kLineMinChangeSpeedPoint := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: e.Formatter.ToFixed(kLine.GetPriceChangeSpeedMin(), 2),
		}
		kLineMaxChangeSpeedPoint := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: e.Formatter.ToFixed(kLine.GetPriceChangeSpeedMax(), 2),
		}

		if interpolation.BtcInterpolationUsdt > 0.00 {
			interpolation.BtcInterpolationUsdt = e.Formatter.FormatPrice(tradeLimit, interpolation.BtcInterpolationUsdt)
		}

		interpolationBtcPoint := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: interpolation.BtcInterpolationUsdt,
		}
		if interpolation.EthInterpolationUsdt > 0.00 {
			interpolation.EthInterpolationUsdt = e.Formatter.FormatPrice(tradeLimit, interpolation.EthInterpolationUsdt)
		}
		interpolationEthPoint := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: interpolation.EthInterpolationUsdt,
		}
		openedBuyPoint := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: 0,
		}
		sellPoint := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: 0,
		}
		buyPoint := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: 0,
		}
		sellPendingPoint := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: 0,
		}
		buyPendingPoint := model.ChartPoint{
			XAxis: kLine.Timestamp.GetPeriodToMinute(),
			YAxis: 0,
		}

		if symbolOrderRaw, ok := orderMap.Load(kLine.Timestamp.GetPeriodToMinute()); ok {
			if symbolOrder, ok := symbolOrderRaw.(model.Order); ok {
				if strings.ToUpper(symbolOrder.Operation) == "BUY" {
					buyPoint.YAxis = symbolOrder.Price
				} else {
					sellPoint.YAxis = symbolOrder.Price
				}
			}
		}

		openedBuyOrder := e.OrderRepository.GetOpenedOrderCached(symbol, "BUY")
		if openedBuyOrder != nil && openedBuyOrder.IsOpened() {
			date, _ := time.Parse("2006-01-02 15:04:05", openedBuyOrder.CreatedAt)
			openedOrderTimestamp := date.UnixMilli() // convert date to timestamp
			if openedOrderTimestamp <= kLine.Timestamp.GetPeriodToMinute() {
				openedBuyPoint.YAxis = openedBuyOrder.Price
			}
		}

		binanceBuyOrder := e.OrderRepository.GetBinanceOrder(symbol, "BUY")
		if binanceBuyOrder != nil && (binanceBuyOrder.IsNew() || binanceBuyOrder.IsPartiallyFilled()) {
			buyPendingPoint.YAxis = binanceBuyOrder.Price
		}

		binanceSellOrder := e.OrderRepository.GetBinanceOrder(symbol, "SELL")
		if binanceSellOrder != nil && (binanceSellOrder.IsNew() || binanceSellOrder.IsPartiallyFilled()) {
			sellPendingPoint.YAxis = binanceSellOrder.Price
		}

		klineKey := fmt.Sprintf("kline-%s", symbol)
		icebergPriceBuyValueKey := fmt.Sprintf("iceberg-price-buy-value-%s", symbol)
		icebergPriceSellValueKey := fmt.Sprintf("iceberg-price-sell-value-%s", symbol)
		icebergQtyBuyValueKey := fmt.Sprintf("iceberg-qty-buy-value-%s", symbol)
		icebergQtySellValueKey := fmt.Sprintf("iceberg-qty-sell-value-%s", symbol)
		capitalizationValueKey := fmt.Sprintf("capitalization-value-%s", symbol)
		capitalizationPriceKey := fmt.Sprintf("capitalization-price-%s", symbol)
		cummulativeTradeQtyKey := fmt.Sprintf("cummulative-trade-qty-%s", symbol)
		klineTradeVolumeBuyKey := fmt.Sprintf("trade-volume-buy-%s", symbol)
		klineTradeVolumeSellKey := fmt.Sprintf("trade-volume-sell-%s", symbol)
		klineAvgChangeSpeedKey := fmt.Sprintf("avg-change-speed-%s", symbol)
		klineMinChangeSpeedKey := fmt.Sprintf("min-change-speed-%s", symbol)
		klineMaxChangeSpeedKey := fmt.Sprintf("max-change-speed-%s", symbol)
		klinePredictKey := fmt.Sprintf("predict-%s", symbol)
		interpolationBtcKey := fmt.Sprintf("interpolation-btc-%s", symbol)
		interpolationEthKey := fmt.Sprintf("interpolation-eth-%s", symbol)
		orderBuyKey := fmt.Sprintf("order-buy-%s", symbol)
		orderSellKey := fmt.Sprintf("order-sell-%s", symbol)
		orderBuyPendingKey := fmt.Sprintf("order-buy-pending-%s", symbol)
		orderSellPendingKey := fmt.Sprintf("order-sell-pending-%s", symbol)
		openedOrderBuyKey := fmt.Sprintf("order-buy-opened-%s", symbol)
		list[klineKey] = append(list[klineKey], klinePoint)
		list[klinePredictKey] = append(list[klinePredictKey], kLinePredictPoint)
		list[interpolationBtcKey] = append(list[interpolationBtcKey], interpolationBtcPoint)
		list[interpolationEthKey] = append(list[interpolationEthKey], interpolationEthPoint)
		list[orderBuyKey] = append(list[orderBuyKey], buyPoint)
		list[orderSellKey] = append(list[orderSellKey], sellPoint)
		list[orderBuyPendingKey] = append(list[orderBuyPendingKey], buyPendingPoint)
		list[orderSellPendingKey] = append(list[orderSellPendingKey], sellPendingPoint)
		list[openedOrderBuyKey] = append(list[openedOrderBuyKey], openedBuyPoint)
		list[klineAvgChangeSpeedKey] = append(list[klineAvgChangeSpeedKey], kLineAvgChangeSpeedPoint)
		list[klineMinChangeSpeedKey] = append(list[klineMinChangeSpeedKey], kLineMinChangeSpeedPoint)
		list[klineMaxChangeSpeedKey] = append(list[klineMaxChangeSpeedKey], kLineMaxChangeSpeedPoint)
		list[klineTradeVolumeBuyKey] = append(list[klineTradeVolumeBuyKey], tradeVolumeBuy)
		list[klineTradeVolumeSellKey] = append(list[klineTradeVolumeSellKey], tradeVolumeSell)
		list[capitalizationValueKey] = append(list[capitalizationValueKey], capitalizationValuePoint)
		list[capitalizationPriceKey] = append(list[capitalizationPriceKey], capitalizationPricePoint)
		list[cummulativeTradeQtyKey] = append(list[cummulativeTradeQtyKey], cummulativeTradeQtyPoint)
		list[icebergPriceBuyValueKey] = append(list[icebergPriceBuyValueKey], icebergPriceBuyPoint)
		list[icebergPriceSellValueKey] = append(list[icebergPriceSellValueKey], icebergPriceSellPoint)
		list[icebergQtyBuyValueKey] = append(list[icebergQtyBuyValueKey], icebergQtyBuyPoint)
		list[icebergQtySellValueKey] = append(list[icebergQtySellValueKey], icebergQtySellPoint)
	}

	return list
}
