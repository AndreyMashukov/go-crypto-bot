package validator

import "gitlab.com/open-soft/go-crypto-bot/src/model"

type OrderValidator struct {
	ProfitOptionsValidator *ProfitOptionsValidator
}

func (v *OrderValidator) Validate(order model.Order) error {
	violation := v.ProfitOptionsValidator.Validate(order.ProfitOptions)

	if violation != nil {
		return violation
	}

	return nil
}
