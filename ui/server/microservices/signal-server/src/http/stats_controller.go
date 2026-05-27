package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/AndreyMashukov/go-crypto-bot/ui/server/microservices/signal-server/src/model"
	"github.com/AndreyMashukov/go-crypto-bot/ui/server/microservices/signal-server/src/repository"
	"strconv"
	"sync"
)

type StatsController struct {
	TradeRepository *repository.TradeRepository
}

func (s *StatsController) GetStatsGrid(context *gin.Context) {
	exchange := context.Query("exchange")
	if "" == exchange {
		context.JSON(400, "Parameter 'exchange' can not be blank.")
		return
	}

	resultMap := sync.Map{}

	wg := sync.WaitGroup{}

	for _, day := range []int64{1, 7, 14, 30, 60, 90} {
		wg.Add(1)
		go func(d int64, e string, r *sync.Map) {
			result := s.TradeRepository.GetTradeStat(30, d, 0.00, e)
			result.Range(func(key, value any) bool {
				if stat, ok := value.(model.TradeStat); ok {
					if listRaw, ok := r.Load(key); ok {
						if list, ok := listRaw.([]model.TradeStat); ok {
							list = append(list, stat)
							r.Store(key, list)
						}
					} else {
						list := make([]model.TradeStat, 0)
						list = append(list, stat)
						r.Store(key, list)
					}
				}

				return true
			})

			wg.Done()
		}(day, exchange, &resultMap)
	}

	wg.Wait()

	jsonMap := make(map[string]any)
	resultMap.Range(func(key, value any) bool {
		if stringKey, ok := key.(string); ok {
			jsonMap[stringKey] = value
		}

		return true
	})

	context.JSON(200, jsonMap)
}

func (s *StatsController) GetStats(context *gin.Context) {
	periodDays, err := strconv.ParseInt(context.Query("periodDays"), 10, 64)
	if err != nil {
		context.JSON(400, fmt.Sprintf("Parameter 'periodDays' is invalid: %s", err.Error()))
		return
	}

	interval, err := strconv.ParseInt(context.Query("interval"), 10, 64)
	if err != nil {
		context.JSON(400, fmt.Sprintf("Parameter 'interval' is invalid: %s", err.Error()))
		return
	}

	exchange := context.Query("exchange")
	if "" == exchange {
		context.JSON(400, "Parameter 'exchange' can not be blank.")
		return
	}

	symbol := context.Param("symbol")
	if "" == symbol {
		context.JSON(400, "Parameter 'symbol' can not be blank.")
		return
	}

	stats := s.TradeRepository.GetTradeStat(interval, periodDays, model.Percent(0.00), "binance")
	if value, ok := stats.Load(symbol); ok {
		context.JSON(200, value)
		return
	}

	context.JSON(404, fmt.Sprintf("Stats for symbol '%s' is not found.", symbol))
}
