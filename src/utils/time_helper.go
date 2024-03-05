package utils

import "time"

type TimeServiceInterface interface {
	WaitSeconds(seconds int64)
	WaitMilliseconds(milliseconds int64)
	GetNowUnix() int64
	GetNowDateTimeString() string
	GetNowDiffMinutes(unixTime int64) float64
}

type TimeHelper struct {
}

func (t *TimeHelper) WaitMilliseconds(milliseconds int64) {
	time.Sleep(time.Millisecond * time.Duration(milliseconds))
}
func (t *TimeHelper) WaitSeconds(seconds int64) {
	time.Sleep(time.Second * time.Duration(seconds))
}
func (t *TimeHelper) GetNowDiffMinutes(unixTime int64) float64 {
	return float64(time.Now().Unix()-unixTime) / 60.00
}
func (t *TimeHelper) GetNowUnix() int64 {
	return time.Now().Unix()
}
func (t *TimeHelper) GetNowDateTimeString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
