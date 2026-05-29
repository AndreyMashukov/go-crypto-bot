package utils

import (
	"context"
	"time"
)

type TimeServiceInterface interface {
	WaitSeconds(seconds int64)
	WaitSecondsCtx(ctx context.Context, seconds int64) error
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

// WaitSecondsCtx blocks for the given number of seconds OR until ctx
// cancels, returning ctx.Err in the latter case. Use this in any loop
// that must observe SIGTERM during a sleep.
func (t *TimeHelper) WaitSecondsCtx(ctx context.Context, seconds int64) error {
	timer := time.NewTimer(time.Second * time.Duration(seconds))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
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
