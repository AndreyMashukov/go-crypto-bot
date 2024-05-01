package model

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
)

const TradeFilterConditionTypeAnd = "and"
const TradeFilterConditionTypeOr = "or"

const TradeFilterParameterPrice = "price"
const TradeFilterParameterDailyPercent = "daily_percent"
const TradeFilterParameterPositionTimeMinutes = "position_time_minutes"
const TradeFilterParameterExtraOrdersToday = "extra_orders_today"

const TradeFilterConditionLt = "lt"
const TradeFilterConditionLte = "lte"
const TradeFilterConditionGt = "gt"
const TradeFilterConditionGte = "gte"
const TradeFilterConditionEq = "eq"
const TradeFilterConditionNeq = "neq"

type TradeFilters []TradeFilter

type TradeFilter struct {
	Symbol    string       `json:"symbol"`
	Parameter string       `json:"parameter"`
	Condition string       `json:"condition"`
	Value     string       `json:"value"`
	Type      string       `json:"type"`
	Children  TradeFilters `json:"children"`
}

func (t *TradeFilters) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), &t)
}
func (t TradeFilters) Value() (driver.Value, error) {
	jsonV, err := json.Marshal(t)
	return string(jsonV), err
}

func (t *TradeFilter) Or() bool {
	return strings.ToLower(t.Type) == TradeFilterConditionTypeOr
}

func (t *TradeFilter) And() bool {
	return strings.ToLower(t.Type) == TradeFilterConditionTypeAnd
}
