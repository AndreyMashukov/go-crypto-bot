package validator

import "github.com/AndreyMashukov/go-crypto-bot/server/src/model"

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
