package validator

import (
	"errors"
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"slices"
	"strings"
)

type ProfitOptionsValidator struct {
}

func (v *ProfitOptionsValidator) Validate(ProfitOptions model.ProfitOptions) error {
	validProfitUnits := []string{
		model.ProfitOptionUnitMinute,
		model.ProfitOptionUnitHour,
		model.ProfitOptionUnitDay,
		model.ProfitOptionUnitMonth,
	}

	hasInvalidProfitPercent := false

	invalidUnits := make([]string, 0)
	for _, option := range ProfitOptions {
		if !slices.Contains(validProfitUnits, option.OptionUnit) {
			invalidUnits = append(invalidUnits, option.OptionUnit)
		}

		if option.OptionPercent < model.MinProfitPercent {
			hasInvalidProfitPercent = true
		}
	}

	if len(invalidUnits) > 0 {
		return errors.New(fmt.Sprintf("ProfitOptions units: %s are invalid", strings.Join(invalidUnits, ", ")))
	}

	if hasInvalidProfitPercent {
		return errors.New(fmt.Sprintf("ProfitOptions `optionProfitPercent` should be greater or equal %.2f", model.MinProfitPercent))
	}

	return nil
}
