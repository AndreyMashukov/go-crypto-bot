package model

const ProfitOptionUnitMinute = "i"
const ProfitOptionUnitHour = "h"
const ProfitOptionUnitDay = "d"
const ProfitOptionUnitMonth = "m"

type ProfitOption struct {
	Index           int64   `json:"index"`
	IsTriggerOption bool    `json:"isTriggerOption"`
	OptionValue     float64 `json:"optionValue"`
	OptionUnit      string  `json:"optionUnit"`
	OptionPercent   Percent `json:"optionPercent"`
	SellPrice       float64 `json:"sellPrice"`
}

type ExtraChargeOption struct {
	Index            int64   `json:"index"`
	BuyPrice         float64 `json:"buyPrice"`
	Percent          Percent `json:"percent"`
	BudgetPercentage Percent `json:"budgetPercentage"`
}

type Signal struct {
	Exchange           string              `json:"exchange"`
	Symbol             string              `json:"symbol"`
	BuyPrice           float64             `json:"buyPrice"`
	Percent            Percent             `json:"percent"`
	ProfitOptions      []ProfitOption      `json:"profitOptions"`
	ExtraChargeOptions []ExtraChargeOption `json:"extraChargeOptions"`
	ExpireTimestamp    int64               `json:"expireTimestamp"`
	PeriodDays         int64               `json:"periodDays"`
	IntervalMinutes    int64               `json:"intervalMinutes"`
}
