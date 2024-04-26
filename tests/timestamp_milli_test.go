package tests

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"testing"
)

func TestShouldAllowToGetTimestampToFrom(t *testing.T) {
	assertion := assert.New(t)
	timestampVal := model.TimestampMilli(1714147241054)
	assertion.Equal(int64(1714147200000), timestampVal.GetPeriodFromMinute())
	assertion.Equal(int64(1714147259999), timestampVal.GetPeriodToMinute())
}
