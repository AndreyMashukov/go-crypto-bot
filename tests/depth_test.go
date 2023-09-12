package tests

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"io/ioutil"
	"testing"
)

func TestShouldAllowToGetDepthAskAndBid(t *testing.T) {
	assert := assert.New(t)
	content, _ := ioutil.ReadFile("example/ethusdt@depth.json")
	var depth ExchangeModel.DepthEvent
	json.Unmarshal(content, &depth)

	assert.Equal(1474.64, depth.Depth.GetBestBid())
	assert.Equal(1552.26, depth.Depth.GetBestAsk())
	assert.Equal(981217.4920299998, depth.Depth.GetAskVolume())
	assert.Equal(1325228.3602399998, depth.Depth.GetBidVolume())
}
