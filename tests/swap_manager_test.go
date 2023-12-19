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

type ExchangeRepositoryMock struct {
	mock.Mock
}

func (m *ExchangeRepositoryMock) CreateSwapPair(swapPair model.SwapPair) (*int64, error) {
	args := m.Called(swapPair)
	id := int64(args.Int(1))
	return &id, args.Error(0)
}
func (m *ExchangeRepositoryMock) UpdateSwapPair(swapPair model.SwapPair) error {
	args := m.Called(swapPair)
	return args.Error(0)
}
func (m *ExchangeRepositoryMock) GetSwapPairs() []model.SwapPair {
	args := m.Called()
	return args.Get(0).([]model.SwapPair)
}
func (m *ExchangeRepositoryMock) GetSwapPairsByBaseAsset(baseAsset string) []model.SwapPair {
	args := m.Called(baseAsset)
	return args.Get(0).([]model.SwapPair)
}
func (m *ExchangeRepositoryMock) GetSwapPairsByQuoteAsset(quoteAsset string) []model.SwapPair {
	args := m.Called(quoteAsset)
	return args.Get(0).([]model.SwapPair)
}
func (m *ExchangeRepositoryMock) GetSwapPair(symbol string) (model.SwapPair, error) {
	args := m.Called(symbol)
	return args.Get(0).(model.SwapPair), args.Error(0)
}

type SwapRepositoryMock struct {
	mock.Mock
	savedChain model.SwapChainEntity
}

func (m *SwapRepositoryMock) GetSwapChain(hash string) (model.SwapChainEntity, error) {
	args := m.Called(hash)
	return args.Get(0).(model.SwapChainEntity), args.Error(1)
}
func (m *SwapRepositoryMock) CreateSwapChain(swapChain model.SwapChainEntity) (*int64, error) {
	m.savedChain = swapChain
	args := m.Called(swapChain)
	id := int64(args.Int(0))
	return &id, args.Error(1)
}
func (m *SwapRepositoryMock) UpdateSwapChain(swapChain model.SwapChainEntity) error {
	args := m.Called(swapChain)
	return args.Error(0)
}
func (m *SwapRepositoryMock) SaveSwapChainCache(asset string, entity model.SwapChainEntity) {
	m.Called(asset, entity)
}

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestSwapSellSellBuy(t *testing.T) {
	swapRepoMock := new(SwapRepositoryMock)
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
	exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "ETH").Return(options1)
	exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "GBP").Return(options2)

	swapRepoMock.On("GetSwapChain", "39bcff3625afb6b785a1a2b01620133c").Return(model.SwapChainEntity{}, errors.New("Not found!"))
	swapRepoMock.On("CreateSwapChain", mock.Anything).Return(99, nil)
	swapRepoMock.On("SaveSwapChainCache", "SOL", mock.Anything).Once()

	swapManager := service.SwapManager{
		SwapRepository:     swapRepoMock,
		ExchangeRepository: exchangeRepoMock,
		Formatter:          &service.Formatter{},
	}

	swapManager.CalculateSwapOptions("SOL")
	chain := swapRepoMock.savedChain
	assertion := assert.New(t)
	assertion.Equal(14.17, chain.Percent.Value())
	assertion.Equal("SSB", chain.Type)
	assertion.Equal("SOL sell-> ETH sell-> GBP buy-> SOL", chain.Title)
	assertion.Equal("SOLETH", chain.SwapOne.Symbol)
	assertion.Equal(0.03374, chain.SwapOne.Price)
	assertion.Equal("ETHGBP", chain.SwapTwo.Symbol)
	assertion.Equal(1783.08, chain.SwapTwo.Price)
	assertion.Equal("SOLGBP", chain.SwapThree.Symbol)
	assertion.Equal(52.38, chain.SwapThree.Price)
	// base amount is 100
	assertion.Greater(100*chain.SwapOne.Price*chain.SwapTwo.Price/chain.SwapThree.Price, 114.50)
}

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

	options4 := make([]model.SwapPair, 0)

	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "SOL").Return(options3)
	exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "ETH").Return(options0)
	exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "GBP").Return(options2)
	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "GBP").Return(options4)
	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "ETH").Return(options4)
	//exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "GBP").Return(options2)

	swapRepoMock.On("GetSwapChain", "f99d6ebc52c3b0ae372a6a58c4d76d66").Return(model.SwapChainEntity{}, errors.New("Not found!"))
	swapRepoMock.On("CreateSwapChain", mock.Anything).Return(99, nil)
	swapRepoMock.On("SaveSwapChainCache", "SOL", mock.Anything).Once()

	swapManager := service.SwapManager{
		SwapRepository:     swapRepoMock,
		ExchangeRepository: exchangeRepoMock,
		Formatter:          &service.Formatter{},
	}

	swapManager.CalculateSwapOptions("SOL")
	chain := swapRepoMock.savedChain
	assertion := assert.New(t)
	assertion.Equal(4.58, chain.Percent.Value())
	assertion.Equal("SBB", chain.Type)
	assertion.Equal("SOL sell-> GBP buy-> ETH buy-> SOL", chain.Title)
	assertion.Equal("SOLGBP", chain.SwapOne.Symbol)
	assertion.Equal(58.58, chain.SwapOne.Price)
	assertion.Equal("ETHGBP", chain.SwapTwo.Symbol)
	assertion.Equal(1782.94, chain.SwapTwo.Price)
	assertion.Equal("SOLETH", chain.SwapThree.Symbol)
	assertion.Equal(0.03123, chain.SwapThree.Price)
	// base amount is 100
	assertion.Greater(100*chain.SwapOne.Price/chain.SwapTwo.Price/chain.SwapThree.Price, 105.00)
}
