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

func TestSwapSellBuyBuy(t *testing.T) {
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

	swapManager := service.SBBSwapFinder{
		Formatter:          &service.Formatter{},
		ExchangeRepository: exchangeRepoMock,
	}

	chain := swapManager.Find("SOL").BestChain
	assertion := assert.New(t)
	assertion.Equal(4.21, chain.Percent.Value())
	assertion.Equal("SBB", chain.Type)
	assertion.Equal("SOL sell-> GBP buy-> ETH buy-> SOL", chain.Title)
	assertion.Equal("SOLGBP", chain.SwapOne.Symbol)
	assertion.Equal(58.56, chain.SwapOne.Price)
	assertion.Equal("ETHGBP", chain.SwapTwo.Symbol)
	assertion.Equal(1782.96, chain.SwapTwo.Price)
	assertion.Equal("SOLETH", chain.SwapThree.Symbol)
	assertion.Equal(0.03133, chain.SwapThree.Price)
	// base amount is 100
	assertion.Greater(100*chain.SwapOne.Price/chain.SwapTwo.Price/chain.SwapThree.Price, 104.80)

	// validate
	swapRepoMock := new(SwapRepositoryMock)

	swapRepoMock.On("GetSwapPairBySymbol", "SOLETH").Return(options0[0], nil)
	swapRepoMock.On("GetSwapPairBySymbol", "ETHGBP").Return(options2[0], nil)
	swapRepoMock.On("GetSwapPairBySymbol", "SOLGBP").Return(options2[1], nil)

	swapChainBuilder := service.SwapChainBuilder{}
	validator := service.SwapValidator{
		SwapRepository: swapRepoMock,
		Formatter:      &service.Formatter{},
		SwapMinPercent: 0.1,
	}

	order := model.Order{
		ExecutedQuantity: 100,
	}

	err = validator.Validate(swapChainBuilder.BuildEntity(*chain, chain.Percent, 0, 0, 0, 0, 0, 0), order)
	assertion.Nil(err)
}
