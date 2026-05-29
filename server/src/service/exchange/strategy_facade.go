package exchange

import (
	"errors"
	"fmt"
	"time"

	"github.com/AndreyMashukov/go-crypto-bot/server/shared/tickstore"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/model"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/repository"
	"github.com/AndreyMashukov/go-crypto-bot/server/src/service"
)

// priceFreshnessWindow caps how stale a tick may be before the facade
// treats it as missing. The legacy IsPriceExpired heuristic on
// model.KLine used a similar threshold; matching it here preserves
// existing strategy behaviour while removing the repo hop.
const priceFreshnessWindow = 30 * time.Second

type StrategyFacadeInterface interface {
	Decide(symbol string) (model.FacadeResponse, error)
}

type StrategyFacade struct {
	DecisionReadStorage repository.DecisionReadStorageInterface
	ExchangeRepository  repository.ExchangeTradeInfoInterface
	OrderRepository     repository.OrderStorageInterface
	BotService          service.BotServiceInterface
	// TickStore is the in-process latest-tick map written by the
	// watcher's emit path. Phase D replaces the legacy
	// ExchangeRepository.GetCurrentKline call with TickStore.Latest so
	// the strategy hot path is no longer coupled to the kline repo.
	TickStore    tickstore.Store
	MinDecisions float64
}

func (s *StrategyFacade) Decide(symbol string) (model.FacadeResponse, error) {
	decisions := s.DecisionReadStorage.GetDecisions(symbol)

	buyScore := 0.00
	sellScore := 0.00
	holdScore := 0.00
	decisionAmount := 0.00
	priceSum := 0.00

	for _, decision := range decisions {
		decisionAmount++
		switch decision.Operation {
		case "BUY":
			buyScore += decision.Score
		case "SELL":
			sellScore += decision.Score
		case "HOLD":
			holdScore += decision.Score
		}
		priceSum += decision.Price
	}

	manualOrder := s.OrderRepository.GetManualOrder(symbol)

	if decisionAmount < s.MinDecisions && manualOrder == nil {
		return model.FacadeResponse{
			Hold: model.DecisionHighestPriorityScore,
			Buy:  0.00,
			Sell: 0.00,
		}, fmt.Errorf("[%s] Not enough decision amount %d of %d", symbol, int64(decisionAmount), int64(s.MinDecisions))
	}

	tradeLimit, err := s.ExchangeRepository.GetTradeLimit(symbol)
	if err != nil {
		return model.FacadeResponse{
			Hold: model.DecisionHighestPriorityScore,
			Buy:  0.00,
			Sell: 0.00,
		}, fmt.Errorf("[%s] %s", symbol, err.Error())
	}

	if s.TickStore == nil {
		return model.FacadeResponse{
			Hold: model.DecisionHighestPriorityScore,
			Buy:  0.00,
			Sell: 0.00,
		}, errors.New("strategy facade: TickStore not wired")
	}

	tick, ok := s.TickStore.Latest(tradeLimit.Symbol)
	if !ok || len(tick.Candles.Series) == 0 {
		return model.FacadeResponse{
			Hold: model.DecisionHighestPriorityScore,
			Buy:  0.00,
			Sell: 0.00,
		}, fmt.Errorf("[%s] no candles on tick", symbol)
	}

	if buyScore > sellScore && time.Since(tick.EventTime) > priceFreshnessWindow {
		return model.FacadeResponse{
			Hold: model.DecisionHighestPriorityScore,
			Buy:  0.00,
			Sell: 0.00,
		}, fmt.Errorf("[%s] tick is stale", symbol)
	}

	if sellScore == model.DecisionHighestPriorityScore || buyScore == model.DecisionHighestPriorityScore {
		holdScore = 0.00
	}

	return model.FacadeResponse{
		Sell: sellScore,
		Buy:  buyScore,
		Hold: holdScore,
	}, nil
}
