package exchange

import (
	"errors"
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/service"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"strings"
)

type PriceCalculatorInterface interface {
	CalculateBuy(tradeLimit model.TradeLimit) (float64, error)
	CalculateSell(tradeLimit model.TradeLimit, order model.Order) (float64, error)
	GetDepth(symbol string, limit int64) model.OrderBookModel
}

type PriceCalculator struct {
	ExchangeRepository repository.ExchangePriceStorageInterface
	OrderRepository    repository.OrderCachedReaderInterface
	FrameService       FrameServiceInterface
	Formatter          *utils.Formatter
	LossSecurity       LossSecurityInterface
	ProfitService      ProfitServiceInterface
	BotService         service.BotServiceInterface
}

func (m *PriceCalculator) CalculateBuy(tradeLimit model.TradeLimit) (float64, error) {
	lastKline := m.ExchangeRepository.GetCurrentKline(tradeLimit.Symbol)

	if lastKline == nil || lastKline.IsPriceExpired() {
		return 0.00, errors.New(fmt.Sprintf("[%s] Current price is unknown, wait...", tradeLimit.Symbol))
	}

	minPrice := m.ExchangeRepository.GetPeriodMinPrice(tradeLimit.Symbol, tradeLimit.MinPriceMinutesPeriod)
	order := m.OrderRepository.GetOpenedOrderCached(tradeLimit.Symbol, "BUY")

	// Extra charge by current price
	if order != nil && order.GetProfitPercent(lastKline.Close, m.BotService.UseSwapCapital()).Lte(tradeLimit.GetBuyOnFallPercent(*order, *lastKline, m.BotService.UseSwapCapital())) {
		return m.LossSecurity.BuyPriceCorrection(lastKline.Close, tradeLimit), nil
	}

	frame := m.FrameService.GetFrame(tradeLimit.Symbol, tradeLimit.FrameInterval, tradeLimit.FramePeriod)
	buyPrice := minPrice

	potentialOpenPrice := lastKline.Close
	for {
		closePrice := m.ProfitService.GetMinClosePrice(tradeLimit, potentialOpenPrice)

		if closePrice <= frame.AvgHigh {
			break
		}

		step := tradeLimit.MinPrice * potentialOpenPrice
		if step < tradeLimit.MinPrice {
			step = tradeLimit.MinPrice
		}

		potentialOpenPrice -= step
		if potentialOpenPrice < 0.00 {
			break
		}
	}

	if potentialOpenPrice > 0.00 && buyPrice > potentialOpenPrice {
		buyPrice = potentialOpenPrice
	}

	if buyPrice > lastKline.Close {
		buyPrice = lastKline.Close
	}

	buyPrice = m.LossSecurity.CheckBuyPriceOnHistory(tradeLimit, buyPrice)

	return m.LossSecurity.BuyPriceCorrection(buyPrice, tradeLimit), nil
}

func (m *PriceCalculator) CalculateSell(tradeLimit model.TradeLimit, order model.Order) (float64, error) {
	lastKline := m.ExchangeRepository.GetCurrentKline(tradeLimit.Symbol)

	if lastKline == nil {
		return 0.00, errors.New("price is unknown")
	}

	minPrice := m.Formatter.FormatPrice(tradeLimit, m.ProfitService.GetMinClosePrice(order, order.Price))
	currentPrice := lastKline.Close

	// todo: Only we do not have active order
	if currentPrice > minPrice {
		minPrice = currentPrice
	}

	if minPrice <= order.Price {
		return 0.00, errors.New("price is less than order price")
	}

	return m.Formatter.FormatPrice(tradeLimit, minPrice), nil
}

func (m *PriceCalculator) GetDepth(symbol string, limit int64) model.OrderBookModel {
	return m.ExchangeRepository.GetDepth(symbol, limit)
}

func (m *PriceCalculator) GetBestFrameBuy(limit model.TradeLimit, marketDepth model.OrderBookModel, frame model.Frame) ([2]float64, error) {
	openPrice := 0.00
	closePrice := 0.00
	potentialOpenPrice := 0.00

	for _, bid := range marketDepth.GetBids() {
		potentialOpenPrice = bid[0].Value
		closePrice = m.ProfitService.GetMinClosePrice(limit, potentialOpenPrice)

		if potentialOpenPrice <= frame.Low {
			break
		}

		if closePrice <= frame.AvgHigh {
			openPrice = potentialOpenPrice
			break
		}
	}

	if openPrice == 0.00 {
		return [2]float64{0.00, 0.00}, errors.New(fmt.Sprintf(
			"Order OrderBookModel is out of Frame [low:%f - high:%f] [must close = %f, if open = %f]",
			frame.AvgLow,
			frame.AvgHigh,
			closePrice,
			potentialOpenPrice,
		))
	}

	return [2]float64{frame.AvgLow, openPrice}, nil
}

func (m *PriceCalculator) InterpolatePrice(symbol string) model.Interpolation {
	asset := strings.ReplaceAll(symbol, "USDT", "")
	btcPair, err := m.ExchangeRepository.GetSwapPairsByAssets("BTC", asset)

	interpolation := model.Interpolation{
		Asset:                asset,
		BtcInterpolationUsdt: 0.00,
		EthInterpolationUsdt: 0.00,
	}

	if err == nil {
		priceXBtc := btcPair.BuyPrice
		lastKlineBtc := m.ExchangeRepository.GetCurrentKline("BTCUSDT")
		if lastKlineBtc != nil && !lastKlineBtc.IsPriceExpired() && !btcPair.IsPriceExpired() {
			interpolation.BtcInterpolationUsdt = priceXBtc * lastKlineBtc.Close
		}
	}

	ethPair, err := m.ExchangeRepository.GetSwapPairsByAssets("ETH", asset)

	if err == nil {
		priceXEth := ethPair.BuyPrice
		lastKlineEth := m.ExchangeRepository.GetCurrentKline("ETHUSDT")
		if lastKlineEth != nil && !lastKlineEth.IsPriceExpired() && !ethPair.IsPriceExpired() {
			interpolation.EthInterpolationUsdt = priceXEth * lastKlineEth.Close
		}
	}

	return interpolation
}
