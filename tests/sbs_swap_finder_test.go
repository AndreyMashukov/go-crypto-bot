package tests

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/service"
	"os"
	"testing"
	"time"
)

func TestSwapSellBuySell(t *testing.T) {
	exchangeRepoMock := new(ExchangeRepositoryMock)

	b, err := os.ReadFile("swap_pair_sbs.json") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}

	var options []model.SwapPair
	err = json.Unmarshal(b, &options)
	if err != nil {
		panic(err)
	}

	options0 := make([]model.SwapPair, 0)
	options[0].PriceTimestamp = time.Now().Unix() + 3600
	options0 = append(options0, options[0])

	options1 := make([]model.SwapPair, 0)
	options[1].PriceTimestamp = time.Now().Unix() + 3600
	options1 = append(options1, options[1])

	options2 := make([]model.SwapPair, 0)
	options[1].PriceTimestamp = time.Now().Unix() + 3600
	options[2].PriceTimestamp = time.Now().Unix() + 3600
	options2 = append(options2, options[1])
	options2 = append(options2, options[2])

	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "ETH").Return(options0)
	exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "BTC").Return(options1)
	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "XRP").Return(options2)

	sbsFinder := service.SBSSwapFinder{
		ExchangeRepository: exchangeRepoMock,
		Formatter:          &service.Formatter{},
	}

	chain := sbsFinder.Find("ETH").BestChain
	assertion := assert.New(t)
	assertion.Equal(3.68, chain.Percent.Value())
	assertion.Equal("SBS", chain.Type)
	assertion.Equal("ETH sell-> BTC buy-> XRP sell-> ETH", chain.Title)
	assertion.Equal("ETHBTC", chain.SwapOne.Symbol)
	assertion.Equal(0.0536, chain.SwapOne.Price)
	assertion.Equal("XRPBTC", chain.SwapTwo.Symbol)
	assertion.Equal(0.00001426, chain.SwapTwo.Price)
	assertion.Equal("XRPETH", chain.SwapThree.Symbol)
	assertion.Equal(0.0002775, chain.SwapThree.Price)
	// base amount is 100
	assertion.Greater(100*chain.SwapOne.Price/chain.SwapTwo.Price*chain.SwapThree.Price, 104.30)
}
