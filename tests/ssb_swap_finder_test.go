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

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestSwapSellSellBuy(t *testing.T) {
	exchangeRepoMock := new(ExchangeRepositoryMock)

	b, err := os.ReadFile("swap_pair_ssb.json") // just pass the file name
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
	options[2].PriceTimestamp = time.Now().Unix() + 3600
	options2 = append(options2, options[2])

	options3 := make([]model.SwapPair, 0)
	options3 = append(options3, options[0])
	options3 = append(options3, options[2])

	options4 := make([]model.SwapPair, 0)

	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "SOL").Return(options3)
	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "ETH").Return(options1)
	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "GBP").Return(options4)

	swapManager := &service.SSBSwapFinder{
		Formatter:          &service.Formatter{},
		ExchangeRepository: exchangeRepoMock,
	}

	chain := swapManager.Find("SOL").BestChain
	assertion := assert.New(t)
	assertion.Equal(13.88, chain.Percent.Value())
	assertion.Equal("SSB", chain.Type)
	assertion.Equal("SOL sell-> ETH sell-> GBP buy-> SOL", chain.Title)
	assertion.Equal("SOLETH", chain.SwapOne.Symbol)
	assertion.Equal(0.03372, chain.SwapOne.Price)
	assertion.Equal("ETHGBP", chain.SwapTwo.Symbol)
	assertion.Equal(1783.06, chain.SwapTwo.Price)
	assertion.Equal("SOLGBP", chain.SwapThree.Symbol)
	assertion.Equal(52.48, chain.SwapThree.Price)
	// base amount is 100
	assertion.Greater(100*chain.SwapOne.Price*chain.SwapTwo.Price/chain.SwapThree.Price, 114.50)

	// validate
	swapRepoMock := new(SwapRepositoryMock)

	swapRepoMock.On("GetSwapPairBySymbol", "SOLETH").Return(options0[0], nil)
	swapRepoMock.On("GetSwapPairBySymbol", "ETHGBP").Return(options1[0], nil)
	swapRepoMock.On("GetSwapPairBySymbol", "SOLGBP").Return(options2[0], nil)

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
