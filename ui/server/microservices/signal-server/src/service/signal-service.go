package service

import (
	"github.com/AndreyMashukov/go-crypto-bot/ui/server/microservices/signal-server/src/client"
	"github.com/AndreyMashukov/go-crypto-bot/ui/server/microservices/signal-server/src/model"
	"github.com/AndreyMashukov/go-crypto-bot/ui/server/microservices/signal-server/src/repository"
	"github.com/AndreyMashukov/go-crypto-bot/ui/server/microservices/signal-server/src/utils"
	"log"
	"sync"
	"time"
)

type SignalService struct {
	TradeRepository         *repository.TradeRepository
	AutoTradeClient         *client.AutoTradeClient
	Formatter               *utils.Formatter
	TradingIntervalMinutes  int64
	MinAllowedProfitPercent model.Percent
}

// GenerateSignals todo generate signals for separate exchanges???
// choose the best price???
func (s *SignalService) GenerateSignals() {
	topWg := sync.WaitGroup{}

	for _, day := range []int64{1, 7, 14, 30} {
		for _, exchange := range []string{"bybit", "binance"} {
			topWg.Add(1)
			go func(e string, d int64) {
				defer topWg.Done()
				statMap := s.TradeRepository.GetTradeStat(s.TradingIntervalMinutes, d, s.MinAllowedProfitPercent, e)
				wg := sync.WaitGroup{}
				statMap.Range(func(key, val any) bool {
					wg.Add(1)
					go func(v any) {
						defer wg.Done()
						stat, _ := v.(model.TradeStat)
						extraChargeOptions := []model.ExtraChargeOption{{
							Index:            0,
							BuyPrice:         stat.ExtraChargePrice,
							Percent:          stat.ExtraChargePercent,
							BudgetPercentage: 100.00,
						}}
						log.Printf("[%s:%s|%dd] BUY %.10f -> SELL %.10f | percent = %.2f", stat.Symbol, e, d, stat.BuyPrice, stat.SellPrice, stat.Percent)

						// signal period in minutes
						tradePeriodMinutes := float64(d) * 24.00 * 60.00

						signal := model.Signal{
							Exchange:        e,
							PeriodDays:      stat.PeriodDays,
							IntervalMinutes: stat.IntervalMinutes,
							Symbol:          stat.Symbol,
							BuyPrice:        stat.BuyPrice,
							Percent:         stat.Percent,
							ProfitOptions: []model.ProfitOption{
								{
									Index:           0,
									IsTriggerOption: true,
									OptionValue:     tradePeriodMinutes / 2.00,
									OptionUnit:      model.ProfitOptionUnitMinute,
									OptionPercent:   stat.Percent,
									SellPrice:       stat.SellPrice,
								},
								{
									Index:           1,
									IsTriggerOption: false,
									OptionValue:     tradePeriodMinutes,
									OptionUnit:      model.ProfitOptionUnitMinute,
									OptionPercent:   stat.PercentHalf,
									SellPrice:       stat.HalfSellPrice,
								},
								{
									Index:           2,
									IsTriggerOption: false,
									OptionValue:     tradePeriodMinutes * 2,
									OptionUnit:      model.ProfitOptionUnitMinute,
									OptionPercent:   stat.PercentNegative,
									SellPrice:       stat.SellPriceNegative,
								},
							},
							ExtraChargeOptions: extraChargeOptions,
							ExpireTimestamp:    time.Now().Add(time.Minute * 10).UnixMilli(),
						}

						// send signal to AutoTrade
						s.AutoTradeClient.SendSignal(signal)
					}(val)

					return true
				})

				wg.Wait()
			}(exchange, day)
		}
	}

	topWg.Wait()
}
