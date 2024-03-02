package service

import (
	"errors"
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/src/repository"
	"strings"
	"time"
)

type SwapValidatorInterface interface {
	Validate(entity model.SwapChainEntity, order model.Order) error
	CalculatePercent(entity model.SwapChainEntity) model.Percent
}

type SwapValidator struct {
	Binance        client.ExchangePriceAPIInterface
	SwapRepository ExchangeRepository.SwapBasicRepositoryInterface
	Formatter      *Formatter
	SwapMinPercent float64
}

func (v *SwapValidator) Validate(entity model.SwapChainEntity, order model.Order) error {
	minPercent := model.Percent(v.SwapMinPercent)

	if entity.Percent.Lt(minPercent) {
		return errors.New(fmt.Sprintf("Swap [%s] too small percent %.2f.", entity.Title, entity.Percent))
	}

	err := v.validateSwap(entity, order, 0)

	if err != nil {
		return err
	}

	err = v.validateSwap(entity, order, 1)

	if err != nil {
		return err
	}

	err = v.validateSwap(entity, order, 2)

	if err != nil {
		return err
	}

	return nil
}

func (v *SwapValidator) CalculatePercent(entity model.SwapChainEntity) model.Percent {
	initialBalance := 100.00
	balance := 0.00

	swapOnePrice, _ := v.SwapRepository.GetSwapPairBySymbol(entity.SwapOne.GetSymbol())
	if entity.SwapOne.IsSell() {
		balance = (initialBalance * swapOnePrice.SellPrice) - (initialBalance*swapOnePrice.SellPrice)*0.002
	}
	if entity.SwapOne.IsBuy() {
		balance = (initialBalance / swapOnePrice.SellPrice) - (initialBalance/swapOnePrice.SellPrice)*0.002
	}
	swapTwoPrice, _ := v.SwapRepository.GetSwapPairBySymbol(entity.SwapTwo.GetSymbol())
	if entity.SwapTwo.IsSell() {
		balance = (balance * swapTwoPrice.SellPrice) - (balance*swapTwoPrice.SellPrice)*0.002
	}
	if entity.SwapTwo.IsBuy() {
		balance = (balance / swapTwoPrice.SellPrice) - (balance/swapTwoPrice.SellPrice)*0.002
	}
	swapThreePrice, _ := v.SwapRepository.GetSwapPairBySymbol(entity.SwapThree.GetSymbol())
	if entity.SwapThree.IsBuy() {
		balance = (balance / swapThreePrice.BuyPrice) - (balance/swapThreePrice.BuyPrice)*0.002
	}
	if entity.SwapThree.IsSell() {
		balance = (balance * swapThreePrice.BuyPrice) - (balance*swapThreePrice.BuyPrice)*0.002
	}
	return v.Formatter.ComparePercentage(initialBalance, balance) - 100.00
}

func (v *SwapValidator) validateSwap(chain model.SwapChainEntity, order model.Order, index int64) error {
	var entity model.SwapTransitionEntity

	if 0 == index {
		entity = *chain.SwapOne
	}
	if 1 == index {
		entity = *chain.SwapTwo
	}
	if 2 == index {
		entity = *chain.SwapThree
	}

	swapCurrentKline, err := v.SwapRepository.GetSwapPairBySymbol(entity.GetSymbol())

	if err != nil {
		return errors.New(fmt.Sprintf("Swap [%s:%s] current price is unknown (%s)", entity.Operation, entity.Symbol, err.Error()))
	}

	if (time.Now().Unix() - swapCurrentKline.PriceTimestamp) > 60 {
		return errors.New(fmt.Sprintf("Swap [%s:%s] price is expired", entity.Operation, entity.Symbol))
	}

	if entity.IsBuy() && v.Formatter.ComparePercentage(entity.Price, swapCurrentKline.BuyPrice).Gte(100.10) {
		return errors.New(fmt.Sprintf("Swap [%s:%s] price is too high", entity.Operation, entity.Symbol))
	}

	if entity.IsSell() && v.Formatter.ComparePercentage(entity.Price, swapCurrentKline.SellPrice).Lte(99.90) {
		return errors.New(fmt.Sprintf("Swap [%s:%s] price is too low", entity.Operation, entity.Symbol))
	}

	notional := chain.GetNotional(order.ExecutedQuantity, index)

	if notional < swapCurrentKline.MinNotional {
		return errors.New(fmt.Sprintf(
			"Swap [%s:%s] Notional index = %d | %f < %f",
			entity.Operation,
			entity.Symbol,
			index,
			notional,
			swapCurrentKline.MinNotional,
		))
	}

	err = v.checkHistoryData(entity.Symbol, entity.Price, entity.Operation)

	if err != nil {
		return errors.New(fmt.Sprintf("Swap %s", err.Error()))
	}

	return nil
}

func (v *SwapValidator) checkHistoryData(symbol string, price float64, operation string) error {
	period := int64(14)
	history := v.Binance.GetKLinesCached(symbol, "1d", period)

	seenTimes := int64(0)

	for _, record := range history {
		if "SELL" == strings.ToUpper(operation) && record.High > price {
			seenTimes++
		}

		if "BUY" == strings.ToUpper(operation) && record.Low < price {
			seenTimes++
		}
	}

	if seenTimes > (period / 2) {
		return nil
	}

	return errors.New(fmt.Sprintf(
		"[%s] %s price %f seen just %d times",
		symbol,
		strings.ToUpper(operation),
		price,
		seenTimes,
	))
}
