package tests

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/src/model"
	"io/ioutil"
	"testing"
)

func TestShouldAllowToGetDepthAskAndBid(t *testing.T) {
	assert := assert.New(t)
	content, _ := ioutil.ReadFile("example/ethusdt@depth.json")
	var depth ExchangeModel.DepthEvent
	json.Unmarshal(content, &depth)

	assert.Equal(1552.26, depth.Depth.GetBestBid())
	assert.Equal(1552.26, depth.Depth.GetBestAsk())
	assert.Equal(1490.8960000000004, depth.Depth.GetAvgBid())
	assert.Equal(1571.989268292683, depth.Depth.GetAvgAsk())
	assert.Equal(1542.5830303030307, depth.Depth.GetBestAvgBid())
	assert.Equal(1561.28275, depth.Depth.GetBestAvgAsk())
	assert.Equal(1552.25, depth.Depth.GetMaxQtyBid())
	assert.Equal(1552.73, depth.Depth.GetMaxQtyAsk())
	assert.Equal(1555.0616926763482, depth.Depth.GetAvgVolAsk())
	assert.Equal(1337.9738996076624, depth.Depth.GetAvgVolBid())
	assert.Equal(981217.4920299998, depth.Depth.GetAskVolume())
	assert.Equal(1325228.3602399998, depth.Depth.GetBidVolume())

	bidIndex, bid := depth.Depth.GetBidPosition(1552.17)
	assert.Equal(5, bidIndex)
	assert.Equal([2]ExchangeModel.Number{{Value: 1552.19}, {Value: 18.699}}, bid)

	askIndex, ask := depth.Depth.GetAskPosition(1598.95)
	assert.Equal(37, askIndex)
	assert.Equal([2]ExchangeModel.Number{{Value: 1598.9}, {Value: 0.059}}, ask)

	bidIndex, bid = depth.Depth.GetBidPosition(100.00)
	assert.Equal(34, bidIndex)
	assert.Equal([2]ExchangeModel.Number{{Value: 500}, {Value: 199.277}}, bid)

	askIndex, ask = depth.Depth.GetAskPosition(999999999.99)
	assert.Equal(40, askIndex)
	assert.Equal([2]ExchangeModel.Number{{Value: 2000.25}, {Value: 2.038}}, ask)
}
