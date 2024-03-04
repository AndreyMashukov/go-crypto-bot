package validator

import "gitlab.com/open-soft/go-crypto-bot/src/model"

type TradeLimitValidator struct {
	ProfitOptionsValidator *ProfitOptionsValidator
}

func (v *TradeLimitValidator) Validate(limit model.TradeLimit) error {
	violation := v.ProfitOptionsValidator.Validate(limit.ProfitOptions)

	if violation != nil {
		return violation
	}

	return nil
}
