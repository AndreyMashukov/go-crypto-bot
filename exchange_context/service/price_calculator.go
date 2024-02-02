package service

import (
	"errors"
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"log"
	"strings"
)

type PriceCalculatorInterface interface {
	CalculateBuy(tradeLimit model.TradeLimit) (float64, error)
	CalculateSell(tradeLimit model.TradeLimit, order model.Order) float64
	GetDepth(symbol string) model.Depth
}

type PriceCalculator struct {
	ExchangeRepository repository.ExchangePriceStorageInterface
	OrderRepository    repository.OrderCachedReaderInterface
	FrameService       FrameServiceInterface
	Binance            client.ExchangePriceAPIInterface
	Formatter          *Formatter
	LossSecurity       LossSecurityInterface
}

func (m *PriceCalculator) CalculateBuy(tradeLimit model.TradeLimit) (float64, error) {
	marketDepth := m.GetDepth(tradeLimit.Symbol)
	lastKline := m.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)

	if lastKline == nil {
		return 0.00, errors.New(fmt.Sprintf("[%s] Current price is unknown, wait...", tradeLimit.Symbol))
	}

	minPrice := m.ExchangeRepository.GetPeriodMinPrice(tradeLimit.Symbol, tradeLimit.MinPriceMinutesPeriod)
	order, err := m.OrderRepository.GetOpenedOrderCached(tradeLimit.Symbol, "BUY")

	// Extra charge by current price
	if err == nil && order.GetProfitPercent(lastKline.Close).Lte(tradeLimit.GetBuyOnFallPercent(order, *lastKline)) {
		extraBuyPrice := minPrice
		if order.GetHoursOpened() >= 24 {
			extraBuyPrice = lastKline.Close
			log.Printf(
				"[%s] Extra buy price is %f (more than 24 hours), profit: %.2f",
				tradeLimit.Symbol,
				extraBuyPrice,
				order.GetProfitPercent(lastKline.Close).Value(),
			)
		} else {
			extraBuyPrice = minPrice
			log.Printf(
				"[%s] Extra buy price is %f (less than 24 hours), profit: %.2f",
				tradeLimit.Symbol,
				extraBuyPrice,
				order.GetProfitPercent(lastKline.Close),
			)
		}

		if extraBuyPrice > lastKline.Close {
			extraBuyPrice = lastKline.Close
		}

		return m.LossSecurity.BuyPriceCorrection(extraBuyPrice, tradeLimit), nil
	}

	frame := m.FrameService.GetFrame(tradeLimit.Symbol, tradeLimit.FrameInterval, tradeLimit.FramePeriod)
	bestFramePrice, err := m.GetBestFrameBuy(tradeLimit, marketDepth, frame)
	buyPrice := minPrice

	if err == nil {
		if buyPrice > bestFramePrice[1] {
			buyPrice = bestFramePrice[1]
		}
	} else {
		log.Printf("[%s] Buy Frame Error: %s, current = %f", tradeLimit.Symbol, err.Error(), lastKline.Close)
		potentialOpenPrice := lastKline.Close
		for {
			closePrice := tradeLimit.GetClosePrice(potentialOpenPrice)

			if closePrice <= frame.AvgHigh {
				break
			}

			potentialOpenPrice -= tradeLimit.MinPrice
		}

		if buyPrice > potentialOpenPrice {
			buyPrice = potentialOpenPrice
			log.Printf("[%s] Choosen potential open price = %f", tradeLimit.Symbol, buyPrice)
		}
	}

	if buyPrice > lastKline.Close {
		buyPrice = lastKline.Close
	}

	log.Printf("[%s] buy price history check", tradeLimit.Symbol)
	buyPrice = m.LossSecurity.CheckBuyPriceOnHistory(tradeLimit, buyPrice)
	closePrice := tradeLimit.GetClosePrice(buyPrice)

	log.Printf(
		"[%s] Trade Frame [low:%f - high:%f](%.2f%s/%.2f%s): BUY Price = %f [min(200) = %f, current = %f, close = %f]",
		tradeLimit.Symbol,
		frame.AvgLow,
		frame.AvgHigh,
		frame.GetMediumVolatilityPercent(),
		"%",
		frame.GetVolatilityPercent(),
		"%",
		buyPrice,
		minPrice,
		lastKline.Close,
		closePrice,
	)

	return m.LossSecurity.BuyPriceCorrection(buyPrice, tradeLimit), nil
}

func (m *PriceCalculator) CalculateSell(tradeLimit model.TradeLimit, order model.Order) float64 {
	marketDepth := m.GetDepth(tradeLimit.Symbol)
	avgPrice := marketDepth.GetBestAvgAsk()

	if 0.00 == avgPrice {
		return m.Formatter.FormatPrice(tradeLimit, avgPrice)
	}

	minPrice := m.Formatter.FormatPrice(tradeLimit, order.GetMinClosePrice(tradeLimit))
	openedOrder, err := m.OrderRepository.GetOpenedOrderCached(tradeLimit.Symbol, "BUY")

	if err != nil {
		return 0.00
	}

	lastKline := m.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)
	if lastKline == nil {
		return 0.00
	}

	currentPrice := lastKline.Close

	if avgPrice > minPrice {
		log.Printf("[%s] Choosen AVG sell price %f", tradeLimit.Symbol, avgPrice)
		minPrice = avgPrice
	}

	var frame model.Frame
	orderHours := openedOrder.GetHoursOpened()

	if orderHours >= 48.00 {
		log.Printf("[%s] Order is opened for %d hours, will be used 8-hours frame", tradeLimit.Symbol, orderHours)
		frame = m.FrameService.GetFrame(tradeLimit.Symbol, "2h", 4)
	} else {
		frame = m.FrameService.GetFrame(tradeLimit.Symbol, "2h", 8)
	}

	bestFrameSell, err := frame.GetBestFrameSell(marketDepth)

	if err == nil {
		if bestFrameSell[0] > minPrice {
			minPrice = bestFrameSell[0]
			log.Printf(
				"[%s] Choosen Frame [low:%f - high:%f] Sell price = %f",
				tradeLimit.Symbol,
				frame.AvgLow,
				frame.AvgHigh,
				minPrice,
			)
		}
	} else {
		log.Printf("[%s] Sell Frame: %s, current = %f", tradeLimit.Symbol, err.Error(), currentPrice)
	}

	if currentPrice > minPrice {
		minPrice = currentPrice
		log.Printf("[%s] Choosen Current sell price %f", tradeLimit.Symbol, minPrice)
	}

	profit := (minPrice - order.Price) * order.ExecutedQuantity

	log.Printf("[%s] Sell price = %f, expected profit = %f$", order.Symbol, minPrice, profit)

	return m.Formatter.FormatPrice(tradeLimit, minPrice)
}

func (m *PriceCalculator) GetDepth(symbol string) model.Depth {
	depth := m.ExchangeRepository.GetDepth(symbol)

	if len(depth.Asks) == 0 && len(depth.Bids) == 0 {
		book, err := m.Binance.GetDepth(symbol)
		if err == nil {
			depth = book.ToDepth(symbol)
			m.ExchangeRepository.SetDepth(depth)
		}
	}

	return depth
}

func (m *PriceCalculator) GetBestFrameBuy(limit model.TradeLimit, marketDepth model.Depth, frame model.Frame) ([2]float64, error) {
	openPrice := 0.00
	closePrice := 0.00
	potentialOpenPrice := 0.00

	for _, bid := range marketDepth.GetBids() {
		potentialOpenPrice = bid[0].Value
		closePrice = limit.GetClosePrice(potentialOpenPrice)

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
			"Order Depth is out of Frame [low:%f - high:%f] [must close = %f, if open = %f]",
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
		lastKlineBtc := m.ExchangeRepository.GetLastKLine("BTCUSDT")
		if lastKlineBtc != nil && !lastKlineBtc.IsPriceExpired() && !btcPair.IsPriceExpired() {
			interpolation.BtcInterpolationUsdt = priceXBtc * lastKlineBtc.Close
		}
	}

	ethPair, err := m.ExchangeRepository.GetSwapPairsByAssets("ETH", asset)

	if err == nil {
		priceXEth := ethPair.BuyPrice
		lastKlineEth := m.ExchangeRepository.GetLastKLine("ETHUSDT")
		if lastKlineEth != nil && !lastKlineEth.IsPriceExpired() && !ethPair.IsPriceExpired() {
			interpolation.EthInterpolationUsdt = priceXEth * lastKlineEth.Close
		}
	}

	return interpolation
}
