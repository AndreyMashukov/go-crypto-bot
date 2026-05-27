package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type Volume float64

func (v *Volume) UnmarshalJSON(b []byte) error {
	var strValue string
	err := json.Unmarshal(b, &strValue)
	if err == nil {
		floatValue, _ := strconv.ParseFloat(strValue, 64)
		*v = Volume(floatValue)
		return nil
	}

	var floatValue float64
	err = json.Unmarshal(b, &floatValue)

	if err == nil {
		*v = Volume(floatValue)
		return nil
	}

	return errors.New(fmt.Sprintf("Volume: unsupported data type given, %s", err.Error()))
}

func (v Volume) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%.12f", v.Value()))
}

func (v Volume) Value() float64 {
	return float64(v)
}

type TimestampMilli int64

func (t *TimestampMilli) UnmarshalJSON(b []byte) error {
	var strValue string
	err := json.Unmarshal(b, &strValue)
	if err == nil {
		intValue, _ := strconv.ParseInt(strValue, 10, 64)
		*t = TimestampMilli(intValue)
		return nil
	}

	var intValue int64
	err = json.Unmarshal(b, &intValue)

	if err == nil {
		*t = TimestampMilli(intValue)
		return nil
	}

	return errors.New(fmt.Sprintf("TimestampMilli: unsupported data type given, %s", err.Error()))
}

func (t TimestampMilli) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Value())
}

func (t TimestampMilli) Value() int64 {
	return int64(t)
}

func (t TimestampMilli) GetPeriodFromMinute() int64 {
	dateTime := time.Unix(0, t.Value()*int64(time.Millisecond))
	newDate := time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), dateTime.Hour(), dateTime.Minute(), 0, 0, dateTime.Location())
	return newDate.UnixMilli()
}

func (t TimestampMilli) GetPeriodToMinute() int64 {
	dateTime := time.Unix(0, t.Value()*int64(time.Millisecond))
	newDate := time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), dateTime.Hour(), dateTime.Minute(), 59, 0, dateTime.Location())
	return newDate.UnixMilli() + 999
}

func (t TimestampMilli) Neq(milli TimestampMilli) bool {
	return t.Value() != milli.Value()
}

func (t TimestampMilli) Eq(milli TimestampMilli) bool {
	return t.Value() == milli.Value()
}

func (t TimestampMilli) PeriodToEq(milli TimestampMilli) bool {
	return t.GetPeriodToMinute() == milli.GetPeriodToMinute()
}

func (t TimestampMilli) Gt(milli TimestampMilli) bool {
	return t.Value() > milli.Value()
}

func (t TimestampMilli) Gte(milli TimestampMilli) bool {
	return t.Value() >= milli.Value()
}

func (t TimestampMilli) Lt(milli TimestampMilli) bool {
	return t.Value() < milli.Value()
}

func (t TimestampMilli) Lte(milli TimestampMilli) bool {
	return t.Value() <= milli.Value()
}

type Price float64

func (p *Price) UnmarshalJSON(b []byte) error {
	var strValue string
	err := json.Unmarshal(b, &strValue)
	if err == nil {
		floatValue, _ := strconv.ParseFloat(strValue, 64)
		*p = Price(floatValue)
		return nil
	}

	var floatValue float64
	err = json.Unmarshal(b, &floatValue)

	if err == nil {
		*p = Price(floatValue)
		return nil
	}

	return errors.New(fmt.Sprintf("Price: unsupported data type given, %s", err.Error()))
}

func (p Price) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%.12f", p.Value()))
}

func (p Price) Value() float64 {
	return float64(p)
}
