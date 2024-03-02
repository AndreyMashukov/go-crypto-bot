package model

import "time"

type KlineSpeed struct {
	Time  time.Time `json:"time"`
	KLine KLine
}
