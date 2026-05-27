package validator

import "github.com/AndreyMashukov/go-crypto-bot/server/src/model"

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
