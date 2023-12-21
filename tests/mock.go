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
	swapAction model.SwapAction
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
func (s *SwapRepositoryMock) GetActiveSwapAction(order model.Order) (model.SwapAction, error) {
	args := s.Called(order)
	return args.Get(0).(model.SwapAction), args.Error(1)
}
func (s *SwapRepositoryMock) UpdateSwapAction(action model.SwapAction) error {
	s.swapAction = action
	args := s.Called(action)
	return args.Error(0)
}
func (s *SwapRepositoryMock) GetSwapChainById(id int64) (model.SwapChainEntity, error) {
	args := s.Called(id)
	return args.Get(0).(model.SwapChainEntity), args.Error(1)
}

type OrderUpdaterMock struct {
	mock.Mock
	order model.Order
}

func (s *OrderUpdaterMock) Update(order model.Order) error {
	s.order = order
	args := s.Called(order)
	return args.Error(0)
}

type BalanceServiceMock struct {
	mock.Mock
}

func (b *BalanceServiceMock) GetAssetBalance(asset string, cache bool) (float64, error) {
	args := b.Called(asset, cache)
	return args.Get(0).(float64), args.Error(1)
}
func (b *BalanceServiceMock) InvalidateBalanceCache(asset string) {
	_ = b.Called(asset)
}

type ExchangeOrderAPIMock struct {
	mock.Mock
}

func (b *ExchangeOrderAPIMock) LimitOrder(symbol string, quantity float64, price float64, operation string, timeInForce string) (model.BinanceOrder, error) {
	args := b.Called(symbol, quantity, price, operation, timeInForce)
	return args.Get(0).(model.BinanceOrder), args.Error(1)
}
func (b *ExchangeOrderAPIMock) QueryOrder(symbol string, orderId int64) (model.BinanceOrder, error) {
	args := b.Called(symbol, orderId)
	return args.Get(0).(model.BinanceOrder), args.Error(1)
}
func (b *ExchangeOrderAPIMock) CancelOrder(symbol string, orderId int64) (model.BinanceOrder, error) {
	args := b.Called(symbol, orderId)
	return args.Get(0).(model.BinanceOrder), args.Error(1)
}

type TimeoutServiceMock struct {
	mock.Mock
}

func (t *TimeoutServiceMock) WaitSeconds(seconds int64) {
	_ = t.Called(seconds)
}
