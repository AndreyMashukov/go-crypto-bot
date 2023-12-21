package tests

import (
	"github.com/stretchr/testify/mock"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
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
func (m *SwapRepositoryMock) GetSwapPairBySymbol(symbol string) (model.SwapPair, error) {
	args := m.Called(symbol)
	return args.Get(0).(model.SwapPair), args.Error(1)
}
