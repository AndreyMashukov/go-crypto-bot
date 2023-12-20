package tests

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/service"
	"os"
	"testing"
	"time"
)

func TestSwapSellBuyBuy(t *testing.T) {
	swapRepoMock := new(SwapRepositoryMock)
	exchangeRepoMock := new(ExchangeRepositoryMock)

	b, err := os.ReadFile("swap_pair_sbb.json") // just pass the file name
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

	options3 := make([]model.SwapPair, 0)
	options3 = append(options3, options[0])
	options3 = append(options3, options[2])

	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "SOL").Return(options3)
	exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "ETH").Return(options0)
	exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "GBP").Return(options2)
	//exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "GBP").Return(options2)

	swapRepoMock.On("GetSwapChain", "f99d6ebc52c3b0ae372a6a58c4d76d66").Return(model.SwapChainEntity{}, errors.New("Not found!"))
	swapRepoMock.On("CreateSwapChain", mock.Anything).Return(99, nil)
	swapRepoMock.On("SaveSwapChainCache", "SOL", mock.Anything).Once()

	swapManager := service.SBBSwapFinder{
		Formatter:          &service.Formatter{},
		ExchangeRepository: exchangeRepoMock,
	}

	chain := swapManager.Find("SOL").BestChain
	assertion := assert.New(t)
	assertion.Equal(4.24, chain.Percent.Value())
	assertion.Equal("SBB", chain.Type)
	assertion.Equal("SOL sell-> GBP buy-> ETH buy-> SOL", chain.Title)
	assertion.Equal("SOLGBP", chain.SwapOne.Symbol)
	assertion.Equal(58.58, chain.SwapOne.Price)
	assertion.Equal("ETHGBP", chain.SwapTwo.Symbol)
	assertion.Equal(1782.94, chain.SwapTwo.Price)
	assertion.Equal("SOLETH", chain.SwapThree.Symbol)
	assertion.Equal(0.03133, chain.SwapThree.Price)
	// base amount is 100
	assertion.Greater(100*chain.SwapOne.Price/chain.SwapTwo.Price/chain.SwapThree.Price, 104.80)
}
