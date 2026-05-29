package tests

import (
	"github.com/AndreyMashukov/go-crypto-bot/server/src/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShouldAllowToGetTimestampToFrom(t *testing.T) {
	assertion := assert.New(t)
	timestampVal := model.TimestampMilli(1714147241054)
	assertion.Equal(int64(1714147200000), timestampVal.GetPeriodFromMinute())
	assertion.Equal(int64(1714147259999), timestampVal.GetPeriodToMinute())
}
