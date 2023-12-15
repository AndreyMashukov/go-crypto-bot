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
}

func (s SwapValidator) Validate(entity model.SwapChainEntity) error {
	if !entity.IsBBS() {
		return errors.New(fmt.Sprintf("Swap [%s] unsupported type given.", entity.Title))
	}

	if entity.Percent.Lt(0.3) {
		return errors.New(fmt.Sprintf("Swap [%s] too small percent %.2f.", entity.Title, entity.Percent))
	}

	err := s.validateSwap(*entity.SwapOne)

	if err != nil {
		return err
	}

	err = s.validateSwap(*entity.SwapTwo)

	if err != nil {
		return err
	}

	err = s.validateSwap(*entity.SwapThree)

	if err != nil {
		return err
	}

	percent := s.CalculatePercent(entity)

	if percent.Lte(entity.Percent) {
		return errors.New(fmt.Sprintf("Swap [%s] percent is decreased: %.2f of %.2f", entity.Title, percent, entity.Percent))
	}

	return nil
}

func (s SwapValidator) CalculatePercent(entity model.SwapChainEntity) model.Percent {
	initialBalance := 100.00
	swapOnePrice, _ := s.SwapRepository.GetSwapPairBySymbol(entity.SwapOne.GetSymbol())
	balance := (initialBalance * swapOnePrice.LastPrice) - (initialBalance*swapOnePrice.LastPrice)*0.002
	swapTwoPrice, _ := s.SwapRepository.GetSwapPairBySymbol(entity.SwapTwo.GetSymbol())
	balance = (balance * swapTwoPrice.LastPrice) - (balance*swapTwoPrice.LastPrice)*0.002
	swapThreePrice, _ := s.SwapRepository.GetSwapPairBySymbol(entity.SwapThree.GetSymbol())
	balance = (balance / swapThreePrice.LastPrice) - (balance/swapThreePrice.LastPrice)*0.002

	return s.Formatter.ComparePercentage(initialBalance, balance) - 100
}

func (s SwapValidator) validateSwap(entity model.SwapTransitionEntity) error {
	swapCurrentKline, err := s.SwapRepository.GetSwapPairBySymbol(entity.GetSymbol())

	if err != nil {
		return errors.New(fmt.Sprintf("Swap [%s:%s] current price is unknown", entity.Operation, entity.Symbol))
	}

	if (time.Now().Unix() - swapCurrentKline.PriceTimestamp) > 60 {
		return errors.New(fmt.Sprintf("Swap [%s:%s] price is expired", entity.Operation, entity.Symbol))
	}

	if entity.IsBuy() && s.Formatter.ComparePercentage(entity.Price, swapCurrentKline.LastPrice).Gte(100.15) {
		return errors.New(fmt.Sprintf("Swap [%s:%s] price is too high", entity.Operation, entity.Symbol))
	}

	if entity.IsSell() && s.Formatter.ComparePercentage(entity.Price, swapCurrentKline.LastPrice).Lte(99.85) {
		return errors.New(fmt.Sprintf("Swap [%s:%s] price is too low", entity.Operation, entity.Symbol))
	}

	return nil
}
