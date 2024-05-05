package model

import (
	"math"
	"time"
)

type SignalProfitOption struct {
	Index           int64   `json:"index"`
	IsTriggerOption bool    `json:"isTriggerOption"`
	OptionValue     float64 `json:"optionValue"`
	OptionUnit      string  `json:"optionUnit"`
	OptionPercent   Percent `json:"optionPercent"`
	SellPrice       float64 `json:"sellPrice"`
}

type SignalExtraChargeOption struct {
	Index            int64   `json:"index"`
	BuyPrice         float64 `json:"buyPrice"`
	Percent          Percent `json:"percent"`
	BudgetPercentage Percent `json:"budgetPercentage"`
}

type Signal struct {
	Symbol             string                    `json:"symbol"`
	BuyPrice           float64                   `json:"buyPrice"`
	Percent            Percent                   `json:"percent"`
	ProfitOptions      []SignalProfitOption      `json:"profitOptions"`
	ExtraChargeOptions []SignalExtraChargeOption `json:"extraChargeOptions"`
	ExpireTimestamp    int64                     `json:"expireTimestamp"`
}

func (s *Signal) GetTTLMilli() time.Duration {
	return time.Duration(int64(math.Max(1000, float64(s.ExpireTimestamp-time.Now().UnixMilli()))))
}

func (s *Signal) IsExpired() bool {
	return time.Now().UnixMilli() >= s.ExpireTimestamp
}

func (s *Signal) GetProfitOptions() ProfitOptions {
	profitOptions := make(ProfitOptions, 0)

	for _, option := range s.ProfitOptions {
		profitOptions = append(profitOptions, ProfitOption{
			Index:           option.Index,
			IsTriggerOption: option.IsTriggerOption,
			OptionValue:     option.OptionValue,
			OptionUnit:      option.OptionUnit,
			OptionPercent:   option.OptionPercent,
		})
	}

	return profitOptions
}

func (s *Signal) GetProfitExtraChargeOptions(limit TradeLimit) ExtraChargeOptions {
	profitOptions := make(ExtraChargeOptions, 0)

	budget := limit.USDTLimit

	for _, option := range s.ExtraChargeOptions {
		profitOptions = append(profitOptions, ExtraChargeOption{
			Index:      option.Index,
			Percent:    option.Percent,
			AmountUsdt: budget * (option.BudgetPercentage.Value() / 100),
		})
	}

	return profitOptions
}
