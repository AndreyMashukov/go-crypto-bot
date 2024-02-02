package service

import (
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"log"
)

type LossSecurityInterface interface {
	IsRiskyBuy(binanceOrder model.BinanceOrder, limit model.TradeLimit) bool
	BuyPriceCorrection(price float64, limit model.TradeLimit) float64
	CheckBuyPriceOnHistory(limit model.TradeLimit, buyPrice float64) float64
}

type LossSecurity struct {
	MlEnabled            bool
	InterpolationEnabled bool
	Formatter            *Formatter
	ExchangeRepository   repository.ExchangeTradeInfoInterface
	Binance              client.ExchangePriceAPIInterface
}

func (l *LossSecurity) IsRiskyBuy(binanceOrder model.BinanceOrder, limit model.TradeLimit) bool {
	kline := l.ExchangeRepository.GetLastKLine(binanceOrder.Symbol)

	if kline != nil && binanceOrder.IsBuy() && binanceOrder.IsNew() {
		if l.MlEnabled {
			predict, predictErr := l.ExchangeRepository.GetPredict(kline.Symbol)
			if predictErr == nil && binanceOrder.Price > l.Formatter.FormatPrice(limit, predict) {
				log.Printf(
					"[%s] ML RISK detected: %f > %f",
					binanceOrder.Symbol,
					binanceOrder.Price,
					l.Formatter.FormatPrice(limit, predict),
				)

				return true
			}
		}

		if binanceOrder.Price > l.Formatter.FormatPrice(limit, kline.Close) {
			fallPercent := model.Percent(100.00 - l.Formatter.ComparePercentage(binanceOrder.Price, kline.Close).Value())
			minPrice := l.ExchangeRepository.GetPeriodMinPrice(binanceOrder.Symbol, 200)

			cancelFallPercent := model.Percent(0.50)

			// If falls more than (min - 0.5%) cancel current
			if fallPercent.Gte(cancelFallPercent) && minPrice-(minPrice*0.005) > kline.Close {
				log.Printf(
					"[%s] Close price RISK detected: %f > %f",
					binanceOrder.Symbol,
					binanceOrder.Price,
					l.Formatter.FormatPrice(limit, kline.Close),
				)

				return true
			}
		}

		if l.InterpolationEnabled {
			interpolation, err := l.ExchangeRepository.GetInterpolation(*kline)
			if err == nil && interpolation.HasBtc() && binanceOrder.Price > l.Formatter.FormatPrice(limit, interpolation.BtcInterpolationUsdt) {
				log.Printf(
					"[%s] BTC Interpolation RISK detected: %f > %f",
					binanceOrder.Symbol,
					binanceOrder.Price,
					l.Formatter.FormatPrice(limit, interpolation.BtcInterpolationUsdt),
				)

				return true
			}

			if err == nil && interpolation.HasEth() && binanceOrder.Price > l.Formatter.FormatPrice(limit, interpolation.EthInterpolationUsdt) {
				log.Printf(
					"[%s] ETH Interpolation RISK detected: %f > %f",
					binanceOrder.Symbol,
					binanceOrder.Price,
					l.Formatter.FormatPrice(limit, interpolation.EthInterpolationUsdt),
				)

				return true
			}
		}
	}

	return false
}

func (l *LossSecurity) BuyPriceCorrection(price float64, limit model.TradeLimit) float64 {
	kline := l.ExchangeRepository.GetLastKLine(limit.Symbol)

	if kline != nil {
		if price > kline.Low {
			log.Printf("[%s] Buy price Low correction %.8f -> %.8f", limit.Symbol, price, kline.Low)
			price = kline.Low
		}
	}

	if l.MlEnabled {
		predict, predictErr := l.ExchangeRepository.GetPredict(limit.Symbol)
		if predictErr == nil && price > predict {
			log.Printf("[%s] Buy price ML correction %.8f -> %.8f", limit.Symbol, price, predict)
			price = predict
		}
	}

	if l.InterpolationEnabled && kline != nil {
		interpolation, err := l.ExchangeRepository.GetInterpolation(*kline)
		if err == nil && interpolation.HasBtc() && price > interpolation.BtcInterpolationUsdt {
			log.Printf("[%s] Buy price BTC Index correction %.8f -> %.8f", limit.Symbol, price, interpolation.BtcInterpolationUsdt)
			price = interpolation.BtcInterpolationUsdt
		}

		if err == nil && interpolation.HasEth() && price > interpolation.EthInterpolationUsdt {
			log.Printf("[%s] Buy price ETH Index correction %.8f -> %.8f", limit.Symbol, price, interpolation.EthInterpolationUsdt)
			price = interpolation.EthInterpolationUsdt
		}
	}

	return l.Formatter.FormatPrice(limit, price)
}

func (l *LossSecurity) CheckBuyPriceOnHistory(limit model.TradeLimit, buyPrice float64) float64 {
	kLines := l.Binance.GetKLinesCached(limit.Symbol, limit.BuyPriceHistoryCheckInterval, limit.BuyPriceHistoryCheckPeriod)

	priceBefore := buyPrice

	for {
		closePrice := limit.GetClosePrice(buyPrice)
		var closePriceMetTimes int64 = 0
		for _, kline := range kLines {
			if kline.High >= closePrice {
				closePriceMetTimes++
			}
		}

		if float64(closePriceMetTimes) > (float64(limit.BuyPriceHistoryCheckPeriod) * 0.8) {
			break
		}

		buyPrice -= limit.MinPrice
	}

	log.Printf("[%s] Price history check completed: %f -> %f", limit.Symbol, priceBefore, buyPrice)

	return buyPrice
}
