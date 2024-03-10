package exchange

import (
	"errors"
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/service"
)

type StrategyFacade struct {
	ExchangeRepository repository.ExchangeRepositoryInterface
	OrderRepository    repository.OrderStorageInterface
	BotService         service.BotServiceInterface
	MinDecisions       float64
}

func (s *StrategyFacade) Decide(symbol string) (model.FacadeResponse, error) {
	decisions := s.ExchangeRepository.GetDecisions(symbol)

	buyScore := 0.00
	sellScore := 0.00
	holdScore := 0.00
	amount := 0.00
	priceSum := 0.00

	for _, decision := range decisions {
		amount = amount + 1.00
		switch decision.Operation {
		case "BUY":
			buyScore += decision.Score
			break
		case "SELL":
			sellScore += decision.Score
			break
		case "HOLD":
			holdScore += decision.Score
			break
		}
		priceSum += decision.Price
	}

	manualOrder := s.OrderRepository.GetManualOrder(symbol)

	if amount != s.MinDecisions && manualOrder == nil {
		return model.FacadeResponse{
			Hold: 999,
			Buy:  0,
			Sell: 0,
		}, errors.New(fmt.Sprintf("[%s] Not enough decision amount %d of %d", symbol, int64(amount), int64(s.MinDecisions)))
	}

	tradeLimit, err := s.ExchangeRepository.GetTradeLimit(symbol)

	if err != nil {
		return model.FacadeResponse{
			Hold: 999,
			Buy:  0,
			Sell: 0,
		}, errors.New(fmt.Sprintf("[%s] %s", symbol, err.Error()))
	}

	lastKline := s.ExchangeRepository.GetLastKLine(tradeLimit.Symbol)

	if lastKline == nil {
		return model.FacadeResponse{
			Hold: 999,
			Buy:  0,
			Sell: 0,
		}, errors.New(fmt.Sprintf("[%s] Last price is unknown... skip!", symbol))
	}

	return model.FacadeResponse{
		Sell: sellScore,
		Buy:  buyScore,
		Hold: holdScore,
	}, nil
}
