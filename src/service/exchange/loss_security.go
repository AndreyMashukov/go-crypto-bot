package exchange

import (
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
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
	Formatter            *utils.Formatter
	ExchangeRepository   repository.ExchangeTradeInfoInterface
	ProfitService        ProfitServiceInterface
}

func (l *LossSecurity) IsRiskyBuy(binanceOrder model.BinanceOrder, limit model.TradeLimit) bool {
	kline := l.ExchangeRepository.GetCurrentKline(binanceOrder.Symbol)

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

		if binanceOrder.Price > l.Formatter.FormatPrice(limit, kline.Close.Value()) {
			fallPercent := model.Percent(100.00 - l.Formatter.ComparePercentage(binanceOrder.Price, kline.Close.Value()).Value())
			minPrice := l.ExchangeRepository.GetPeriodMinPrice(binanceOrder.Symbol, 200)

			cancelFallPercent := model.Percent(model.MinProfitPercent)

			// If falls more than (min - 0.5%) cancel current
			if fallPercent.Gte(cancelFallPercent) && minPrice-(minPrice*0.005) > kline.Close.Value() {
				log.Printf(
					"[%s] Close price RISK detected: %f > %f",
					binanceOrder.Symbol,
					binanceOrder.Price,
					l.Formatter.FormatPrice(limit, kline.Close.Value()),
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
	kline := l.ExchangeRepository.GetCurrentKline(limit.Symbol)

	if kline != nil {
		if price > kline.Low.Value() {
			price = kline.Low.Value()
		}
	}

	if l.MlEnabled {
		predict, predictErr := l.ExchangeRepository.GetPredict(limit.Symbol)
		// todo: in the future is should be `predict.Min` value for current time interval
		if predictErr == nil && price > predict {
			price = predict
		}
	}

	if l.InterpolationEnabled && kline != nil {
		interpolation, err := l.ExchangeRepository.GetInterpolation(*kline)
		if err == nil && interpolation.HasBtc() && price > interpolation.BtcInterpolationUsdt {
			price = interpolation.BtcInterpolationUsdt
		}

		if err == nil && interpolation.HasEth() && price > interpolation.EthInterpolationUsdt {
			price = interpolation.EthInterpolationUsdt
		}
	}

	return l.Formatter.FormatPrice(limit, price)
}

func (l *LossSecurity) CheckBuyPriceOnHistory(limit model.TradeLimit, buyPrice float64) float64 {
	return l.ProfitService.CheckBuyPriceOnHistory(limit, buyPrice)
}
