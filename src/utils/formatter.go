package utils

import (
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"log"
	"math"
	"strconv"
	"strings"
)

type Formatter struct {
}

func (m *Formatter) FormatPrice(limit model.TradeLimitInterface, price float64) float64 {
	if price < limit.GetMinPrice() {
		return limit.GetMinPrice()
	}

	split := strings.Split(fmt.Sprintf("%s", strconv.FormatFloat(limit.GetMinPrice(), 'f', -1, 64)), ".")
	precision := 0
	if len(split) > 1 {
		precision = len(split[1])
	}
	ratio := math.Pow(10, float64(precision))
	return math.Round(price*ratio) / ratio
}

func (m *Formatter) FormatQuantity(limit model.TradeLimitInterface, quantity float64) float64 {
	if quantity < limit.GetMinQuantity() {
		return limit.GetMinQuantity()
	}

	splitQty := strings.Split(fmt.Sprintf("%s", strconv.FormatFloat(quantity, 'f', -1, 64)), ".")
	split := strings.Split(fmt.Sprintf("%s", strconv.FormatFloat(limit.GetMinQuantity(), 'f', -1, 64)), ".")
	precision := 0
	if len(split) > 1 {
		precision = len(split[1])
	}

	second := "00"
	if precision > 0 && len(splitQty) > 1 {
		substr := precision
		if len(splitQty[1]) < substr {
			substr = len(splitQty[1])
		}

		second = splitQty[1][0:substr]
	}
	quantity, _ = strconv.ParseFloat(fmt.Sprintf("%s.%s", splitQty[0], second), 64)

	return quantity
}

func (m *Formatter) ComparePercentage(first float64, second float64) model.Percent {
	return model.Percent(second * 100.00 / first)
}

func (m *Formatter) Round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func (m *Formatter) ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(m.Round(num*output)) / output
}

func (m *Formatter) Floor(num float64) int64 {
	return int64(math.Floor(num))
}

func (m *Formatter) BinanceIntervalToByBitInterval(interval string) string {
	// ByBit:
	// 1 3 5 15 30 60 120 240 360 720 minute
	// D day
	// W week
	// M month
	// Binance:
	// 1m 3m 5m 15m 30m 1h 2h 4h 6h 8h 12h
	// 1d 3d 1w 1M
	switch interval {
	case "1m":
		return "1"
	case "3m":
		return "3"
	case "5m":
		return "5"
	case "15m":
		return "15"
	case "30m":
		return "30"
	case "1h":
		return "60"
	case "2h":
		return "120"
	case "4h":
		return "240"
	case "6h":
		return "360"
	case "12h":
		return "720"
	case "1d":
		return "D"
	case "1w":
		return "W"
	case "1M":
		return "M"
	default:
		log.Panicf("Interval %s is not supported by ByBitIntervalToBinanceInterval", interval)
	}

	return ""
}

func (m *Formatter) ByBitIntervalToBinanceInterval(interval string) string {
	// ByBit:
	// 1 3 5 15 30 60 120 240 360 720 minute
	// D day
	// W week
	// M month
	// Binance:
	// 1m 3m 5m 15m 30m 1h 2h 4h 6h 8h 12h
	// 1d 3d 1w 1M
	switch interval {
	case "1":
		return "1m"
	case "3":
		return "3m"
	case "5":
		return "5m"
	case "15":
		return "15m"
	case "30":
		return "30m"
	case "60":
		return "1h"
	case "120":
		return "2h"
	case "240":
		return "4h"
	case "360":
		return "6h"
	case "720":
		return "12h"
	case "D":
		return "1d"
	case "W":
		return "1w"
	case "M":
		return "1M"
	default:
		log.Panicf("Interval %s is not supported by ByBitIntervalToBinanceInterval", interval)
	}

	return ""
}

func (m *Formatter) ByBitStatusToBinanceStatus(status string) string {
	// ByBit:
	// - New
	// - PartiallyFilled
	// - Untriggered
	// - Rejected
	// - PartiallyFilledCanceled
	// - Filled
	// - Cancelled
	// - Triggered
	// - Deactivated
	// Binance:
	// - NEW
	// - PARTIALLY_FILLED
	// - FILLED
	// - CANCELED
	// - PENDING_CANCEL
	// - REJECTED
	// - EXPIRED
	switch status {
	case "New":
		return "NEW"
	case "PartiallyFilled":
	case "PartiallyFilledCanceled":
		return "PARTIALLY_FILLED"
	case "Rejected":
		return "REJECTED"
	case "Filled":
		return "FILLED"
	case "Canceled":
	case "Cancelled":
		return "CANCELED"
	default:
		log.Panicf("Status %s is not supported by ByBitStatusToBinanceStatus", status)
	}

	return ""
}

func (m *Formatter) ByBitSideToBinanceSide(side string) string {
	switch side {
	case "Sell":
		return "SELL"
	case "Buy":
		return "BUY"
	default:
		log.Panicf("Side %s is not supported by ByBitSideToBinanceSide", side)
	}

	return ""
}

func (m *Formatter) BinanceSideToByBitSide(side string) string {
	switch side {
	case "SELL":
		return "Sell"
	case "BUY":
		return "Buy"
	default:
		log.Panicf("Side %s is not supported by BinanceSideToByBitSide", side)
	}

	return ""
}

func (m *Formatter) ByBitTypeToBinanceType(orderType string) string {
	switch orderType {
	case "Limit":
		return "LIMIT"
	default:
		log.Panicf("Order type %s is not supported by ByBitTypeToBinanceType", orderType)
	}

	return ""
}

func (m *Formatter) ByBitOrderToBinanceOrder(byBitOrder model.ByBitOrder) model.BinanceOrder {
	return model.BinanceOrder{
		OrderId:     byBitOrder.OrderId,
		Symbol:      strings.ToUpper(byBitOrder.Symbol),
		Price:       byBitOrder.Price,
		OrigQty:     byBitOrder.OrigQty,
		ExecutedQty: byBitOrder.ExecutedQty,
		Status:      m.ByBitStatusToBinanceStatus(byBitOrder.Status),
		Type:        m.ByBitTypeToBinanceType(byBitOrder.Type),
		Side:        m.ByBitSideToBinanceSide(byBitOrder.Side),
		Timestamp:   byBitOrder.Timestamp,
	}
}

func (m *Formatter) ByBitHistoryKlineToBinanceHistoryKline(kLine model.ByBitKLineHistory) model.KLineHistory {
	openTime, _ := strconv.ParseInt(kLine.OpenTime, 10, 64)

	return model.KLineHistory{
		OpenTime:         model.TimestampMilli(openTime),
		Open:             kLine.Open,
		High:             kLine.High,
		Low:              kLine.Low,
		Close:            kLine.Close,
		Volume:           kLine.Volume,
		CloseTime:        model.TimestampMilli(model.TimestampMilli(openTime).GetPeriodToMinute()),
		QuoteAssetVolume: kLine.Turnover,
	}
}

func (m *Formatter) ByBitTradeToBinanceTrade(trade model.ByBitTrade) model.Trade {
	return model.Trade{
		AggregateTradeId: trade.ExecId,
		Price:            trade.Price,
		Symbol:           trade.Symbol,
		Quantity:         trade.Size,
		IsBuyerMaker:     trade.Side == model.ByBitTradeSideSell,
		Timestamp:        trade.Time,
	}
}

func (m *Formatter) ByBitSymbolStatusToBinanceSymbolStatus(status string) string {
	switch status {
	case "Trading":
		return "TRADING"
	}

	return ""
}

func (m *Formatter) ByBitTickerToBinanceTicker(ticker model.ByBitTicker) model.WSTickerPrice {
	return model.WSTickerPrice{
		Symbol: ticker.Symbol,
		Price:  ticker.LastPrice,
	}
}

func (m *Formatter) ByBitExchangeSymbolToBinanceExchangeSymbol(symbol model.ByBitExchangeSymbol) model.ExchangeSymbol {
	filters := make([]model.ExchangeFilter, 0)

	filters = append(filters, model.ExchangeFilter{
		FilterType: model.BinanceExchangeFilterTypePrice,
		MinPrice:   &symbol.PriceFilter.TickSize,
		MaxPrice:   nil,
		TickSize:   &symbol.PriceFilter.TickSize,
	})

	filters = append(filters, model.ExchangeFilter{
		FilterType:  model.BinanceExchangeFilterTypeLotSize,
		MinQuantity: &symbol.LotSizeFilter.MinOrderQty,
		MaxQuantity: &symbol.LotSizeFilter.MaxOrderQty,
	})

	filters = append(filters, model.ExchangeFilter{
		FilterType:  model.BinanceExchangeFilterTypeNotional,
		MinNotional: &symbol.LotSizeFilter.MinOrderAmt,
		MaxNotional: &symbol.LotSizeFilter.MaxOrderAmt,
	})

	return model.ExchangeSymbol{
		Symbol:             symbol.Symbol,
		Status:             m.ByBitSymbolStatusToBinanceSymbolStatus(symbol.Status),
		BaseAsset:          symbol.BaseCoin,
		QuoteAsset:         symbol.QuoteCoin,
		BaseAssetPrecision: symbol.LotSizeFilter.BasePrecision,
		QuotePrecision:     symbol.LotSizeFilter.QuotePrecision,
		Filters:            filters,
	}
}
