package service

import "time"

type TimeoutServiceInterface interface {
	WaitSeconds(seconds int64)
}

type TimeoutService struct {
}

func (t *TimeoutService) WaitSeconds(seconds int64) {
	time.Sleep(time.Second * time.Duration(seconds))
}
