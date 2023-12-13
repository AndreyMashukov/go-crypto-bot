package service

import (
	"errors"
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"time"
)

type SwapValidator struct {
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	Formatter          *Formatter
}

func (s SwapValidator) Validate(entity model.SwapChainEntity) error {
	if !entity.IsBBS() {
		return errors.New(fmt.Sprintf("Swap [%s] unsupported type given.", entity.Title))
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
	swapOnePrice := s.ExchangeRepository.GetLastKLine(entity.SwapOne.GetSymbol())
	balance := (initialBalance * swapOnePrice.Close) - (initialBalance*swapOnePrice.Close)*0.002
	swapTwoPrice := s.ExchangeRepository.GetLastKLine(entity.SwapTwo.GetSymbol())
	balance = (balance * swapTwoPrice.Close) - (balance*swapTwoPrice.Close)*0.002
	swapThreePrice := s.ExchangeRepository.GetLastKLine(entity.SwapThree.GetSymbol())
	balance = (balance / swapThreePrice.Close) - (balance/swapThreePrice.Close)*0.002

	return s.Formatter.ComparePercentage(initialBalance, balance) - 100
}

func (s SwapValidator) validateSwap(entity model.SwapTransitionEntity) error {
	swapCurrentKline := s.ExchangeRepository.GetLastKLine(entity.GetSymbol())

	if swapCurrentKline == nil {
		return errors.New(fmt.Sprintf("Swap [%s:%s] current price is unknown", entity.Operation, entity.Symbol))
	}

	timestampDeadline := time.Now().Unix() - 30

	if (swapCurrentKline.Timestamp / 1000) < timestampDeadline {
		return errors.New(fmt.Sprintf("Swap [%s:%s] price is expired", entity.Operation, entity.Symbol))
	}

	if entity.IsBuy() && s.Formatter.ComparePercentage(entity.Price, swapCurrentKline.Close).Gte(100.15) {
		return errors.New(fmt.Sprintf("Swap [%s:%s] price is too high", entity.Operation, entity.Symbol))
	}

	if entity.IsSell() && s.Formatter.ComparePercentage(entity.Price, swapCurrentKline.Close).Lte(99.85) {
		return errors.New(fmt.Sprintf("Swap [%s:%s] price is too low", entity.Operation, entity.Symbol))
	}

	return nil
}
