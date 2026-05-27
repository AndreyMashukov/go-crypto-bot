package model

import "math"

type Percent float64

func (p Percent) Gte(percent Percent) bool {
	return p.Value() >= percent.Value()
}

func (p Percent) Lt(percent Percent) bool {
	return p.Value() < percent.Value()
}

func (p Percent) Lte(percent Percent) bool {
	return p.Value() <= percent.Value()
}

func (p Percent) Division(divider float64) Percent {
	return Percent(p.Value() / divider)
}

func (p Percent) Multiply(multiplier float64) Percent {
	return Percent(p.Value() * multiplier)
}

func (p Percent) Abs() Percent {
	return Percent(math.Abs(p.Value()))
}

func (p Percent) Value() float64 {
	return float64(p)
}
