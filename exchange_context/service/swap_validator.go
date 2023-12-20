package service

import (
	"errors"
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"time"
)

type SwapValidator struct {
	SwapRepository *ExchangeRepository.SwapRepository
	Formatter      *Formatter
	SwapMinPercent float64
}

func (s SwapValidator) Validate(entity model.SwapChainEntity, order model.Order) error {
	minPercent := model.Percent(s.SwapMinPercent)

	if entity.Percent.Lt(minPercent) {
		return errors.New(fmt.Sprintf("Swap [%s] too small percent %.2f.", entity.Title, entity.Percent))
	}

	err := s.validateSwap(*entity.SwapOne, order)

	if err != nil {
		return err
	}

	err = s.validateSwap(*entity.SwapTwo, order)

	if err != nil {
		return err
	}

	err = s.validateSwap(*entity.SwapThree, order)

	if err != nil {
		return err
	}

	return nil
}

func (s SwapValidator) CalculatePercent(entity model.SwapChainEntity) model.Percent {
	initialBalance := 100.00
	balance := 0.00

	swapOnePrice, _ := s.SwapRepository.GetSwapPairBySymbol(entity.SwapOne.GetSymbol())
	if entity.SwapOne.IsSell() {
		balance = (initialBalance * swapOnePrice.SellPrice) - (initialBalance*swapOnePrice.SellPrice)*0.002
	}
	if entity.SwapOne.IsBuy() {
		balance = (initialBalance / swapOnePrice.SellPrice) - (initialBalance/swapOnePrice.SellPrice)*0.002
	}
	swapTwoPrice, _ := s.SwapRepository.GetSwapPairBySymbol(entity.SwapTwo.GetSymbol())
	if entity.SwapTwo.IsSell() {
		balance = (balance * swapTwoPrice.SellPrice) - (balance*swapTwoPrice.SellPrice)*0.002
	}
	if entity.SwapTwo.IsBuy() {
		balance = (balance / swapTwoPrice.SellPrice) - (balance/swapTwoPrice.SellPrice)*0.002
	}
	swapThreePrice, _ := s.SwapRepository.GetSwapPairBySymbol(entity.SwapThree.GetSymbol())
	if entity.SwapThree.IsBuy() {
		balance = (balance / swapThreePrice.BuyPrice) - (balance/swapThreePrice.BuyPrice)*0.002
	}
	if entity.SwapThree.IsSell() {
		balance = (balance * swapThreePrice.BuyPrice) - (balance*swapThreePrice.BuyPrice)*0.002
	}
	return s.Formatter.ComparePercentage(initialBalance, balance) - 100
}

func (s SwapValidator) validateSwap(entity model.SwapTransitionEntity, order model.Order) error {
	swapCurrentKline, err := s.SwapRepository.GetSwapPairBySymbol(entity.GetSymbol())

	if err != nil {
		return errors.New(fmt.Sprintf("Swap [%s:%s] current price is unknown", entity.Operation, entity.Symbol))
	}

	if (time.Now().Unix() - swapCurrentKline.PriceTimestamp) > 60 {
		return errors.New(fmt.Sprintf("Swap [%s:%s] price is expired", entity.Operation, entity.Symbol))
	}

	if entity.IsBuy() && s.Formatter.ComparePercentage(entity.Price, swapCurrentKline.BuyPrice).Gte(100.10) {
		return errors.New(fmt.Sprintf("Swap [%s:%s] price is too high", entity.Operation, entity.Symbol))
	}

	if entity.IsSell() && s.Formatter.ComparePercentage(entity.Price, swapCurrentKline.SellPrice).Lte(99.90) {
		return errors.New(fmt.Sprintf("Swap [%s:%s] price is too low", entity.Operation, entity.Symbol))
	}

	notional := entity.Price * order.ExecutedQuantity
	if notional < swapCurrentKline.MinNotional {
		return errors.New(fmt.Sprintf(
			"Swap [%s:%s] Notinal %f < %f",
			entity.Operation,
			entity.Symbol,
			notional,
			swapCurrentKline.MinNotional,
		))
	}

	return nil
}
