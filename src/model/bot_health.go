package model

import (
	"github.com/rafacas/sysstats"
)

const MlStatusLearning = "learning"
const MlStatusReady = "ready"
const DbStatusOk = "ok"
const DbStatusFail = "fail"
const RedisStatusOk = "ok"
const RedisStatusFail = "fail"
const BinanceStatusOk = "ok"
const BinanceStatusBan = "ban"
const BinanceStatusDisconnected = "disconnected"
const BinanceStatusApiKeyCheck = "api_key_checking"

type BotHealth struct {
	Bot           Bot               `json:"bot"`
	MlStatus      string            `json:"mlStatus"`
	DbStatus      string            `json:"dbStatus"`
	RedisStatus   string            `json:"redisStatus"`
	BinanceStatus string            `json:"binanceStatus"`
	Cores         int               `json:"cores"`
	Memory        sysstats.MemStats `json:"memory"`
	LoadAvg       sysstats.LoadAvg  `json:"loadAvg"`
	Updates       map[string]string `json:"updates"`
}
