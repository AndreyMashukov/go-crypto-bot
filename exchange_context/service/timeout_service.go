package service

import "time"

type TimeServiceInterface interface {
	WaitSeconds(seconds int64)
	GetNowDiffMinutes(unixTime int64) float64
}

type TimeService struct {
}

func (t *TimeService) WaitSeconds(seconds int64) {
	time.Sleep(time.Second * time.Duration(seconds))
}
func (t *TimeService) GetNowDiffMinutes(unixTime int64) float64 {
	return float64(time.Now().Unix()-unixTime) / 60.00
}
