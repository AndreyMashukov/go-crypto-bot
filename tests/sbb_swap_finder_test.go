package tests

import (
	"encoding/json"
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

	swapChain := swapChainBuilder.BuildEntity(*chain, chain.Percent, 0, 0, 0, 0, 0, 0)
	err = validator.Validate(swapChain, order)
	assertion.Nil(err)

	// execute

	balanceServiceMock := new(BalanceServiceMock)
	orderRepositoryMock := new(OrderUpdaterMock)
	binanceMock := new(ExchangeOrderAPIMock)

	assetBalance := order.ExecutedQuantity

	balanceServiceMock.On("GetAssetBalance", "SOL", false).Times(1).Return(order.ExecutedQuantity, nil)
	swapRepoMock.On("UpdateSwapAction", mock.Anything).Return(nil)
	swapRepoMock.On("GetActiveSwapAction", order).Return(model.SwapAction{
		Id:              990,
		OrderId:         order.Id,
		BotId:           1,
		SwapChainId:     swapChain.Id,
		Asset:           swapChain.SwapOne.BaseAsset,
		Status:          model.SwapActionStatusPending,
		StartTimestamp:  time.Now().Unix(),
		StartQuantity:   assetBalance,
		SwapOneSymbol:   swapChain.SwapOne.GetSymbol(),
		SwapOnePrice:    swapChain.SwapOne.Price,
		SwapTwoSymbol:   swapChain.SwapTwo.GetSymbol(),
		SwapTwoPrice:    swapChain.SwapTwo.Price,
		SwapThreeSymbol: swapChain.SwapThree.GetSymbol(),
		SwapThreePrice:  swapChain.SwapThree.Price,
	}, nil)
	swapRepoMock.On("GetSwapChainById", swapChain.Id).Return(swapChain, nil)

	binanceMock.On("LimitOrder", "SOLGBP", 100.00, 58.56, "SELL", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             int64(19),
		Symbol:              "SOLGBP",
		ExecutedQty:         0.00,
		OrigQty:             100.00,
		Price:               58.56,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("QueryOrder", "SOLGBP", int64(19)).Times(1).Return(model.BinanceOrder{
		Status:              "PARTIALLY_FILLED",
		OrderId:             int64(19),
		ExecutedQty:         80.00,
		OrigQty:             100.00,
		Symbol:              "SOLGBP",
		Price:               58.56,
		Side:                "SELL",
		CummulativeQuoteQty: 80 * 58.56,
	}, nil)
	binanceMock.On("QueryOrder", "SOLGBP", int64(19)).Times(2).Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             int64(19),
		ExecutedQty:         100.00,
		OrigQty:             100.00,
		Symbol:              "SOLGBP",
		Price:               58.56,
		Side:                "SELL",
		CummulativeQuoteQty: 100 * 58.56,
	}, nil)

	gbpInitialBalance := 50.99
	balanceServiceMock.On("GetAssetBalance", "GBP", false).Return(5856.00+gbpInitialBalance, nil)

	binanceMock.On("LimitOrder", "ETHGBP", 3.2844, 1782.96, "BUY", "GTC").Return(model.BinanceOrder{
		Status:      "NEW",
		OrderId:     int64(20),
		Symbol:      "ETHGBP",
		ExecutedQty: 0.00,
		OrigQty:     3.284,
		Price:       1782.96,
	}, nil)
	binanceMock.On("QueryOrder", "ETHGBP", int64(20)).Times(1).Return(model.BinanceOrder{
		Status:      "PARTIALLY_FILLED",
		OrderId:     int64(20),
		Symbol:      "ETHGBP",
		ExecutedQty: 1.272,
		OrigQty:     3.284,
		Price:       1782.96,
	}, nil)
	binanceMock.On("QueryOrder", "ETHGBP", int64(20)).Times(2).Return(model.BinanceOrder{
		Status:      "FILLED",
		OrderId:     int64(20),
		Symbol:      "ETHGBP",
		ExecutedQty: 3.284,
		OrigQty:     3.284,
		Price:       1782.96,
	}, nil)
	ethInitialBalance := 2.99
	balanceServiceMock.On("GetAssetBalance", "ETH", false).Return(3.282+ethInitialBalance, nil)

	binanceMock.On("LimitOrder", "SOLETH", 104.819, 0.03133, "BUY", "GTC").Return(model.BinanceOrder{
		Status:      "NEW",
		OrderId:     int64(21),
		Symbol:      "SOLETH",
		ExecutedQty: 0.00,
		OrigQty:     104.755,
		Price:       0.03133,
	}, nil)
	binanceMock.On("QueryOrder", "SOLETH", int64(21)).Times(1).Return(model.BinanceOrder{
		Status:      "PARTIALLY_FILLED",
		OrderId:     int64(21),
		Symbol:      "SOLETH",
		ExecutedQty: 12.00,
		OrigQty:     104.755,
		Price:       0.03133,
	}, nil)
	binanceMock.On("QueryOrder", "SOLETH", int64(21)).Times(2).Return(model.BinanceOrder{
		Status:      "FILLED",
		OrderId:     int64(21),
		Symbol:      "SOLETH",
		ExecutedQty: 104.755,
		OrigQty:     104.755,
		Price:       0.03133,
	}, nil)

	orderRepositoryMock.On("Update", mock.Anything).Once().Return(nil)
	swapRepoMock.On("UpdateSwapAction", mock.Anything).Return(nil)
	balanceServiceMock.On("InvalidateBalanceCache", "SOL").Once()
	solInitialBalance := 50.00
	balanceServiceMock.On("GetAssetBalance", "SOL", false).Times(2).Return(104.72+solInitialBalance, nil)

	timeServiceMock := new(TimeServiceMock)
	timeServiceMock.On("WaitSeconds", int64(5)).Times(3)
	timeServiceMock.On("WaitSeconds", int64(7)).Times(3)
	timeServiceMock.On("GetNowDiffMinutes", mock.Anything).Return(0.50)

	executor := service.SwapExecutor{
		SwapRepository:  swapRepoMock,
		OrderRepository: orderRepositoryMock,
		BalanceService:  balanceServiceMock,
		Binance:         binanceMock,
		TimeoutService:  timeServiceMock,
		Formatter:       &service.Formatter{},
	}

	executor.Execute(order)

	assertion.Equal(104.755, *swapRepoMock.swapAction.EndQuantity)
	assertion.Equal(int64(19), *swapRepoMock.swapAction.SwapOneExternalId)
	assertion.Equal("SOLGBP", swapRepoMock.swapAction.SwapOneSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapOneExternalStatus)
	assertion.Equal(int64(20), *swapRepoMock.swapAction.SwapTwoExternalId)
	assertion.Equal("ETHGBP", swapRepoMock.swapAction.SwapTwoSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapTwoExternalStatus)
	assertion.Equal(int64(21), *swapRepoMock.swapAction.SwapThreeExternalId)
	assertion.Equal("SOLETH", swapRepoMock.swapAction.SwapThreeSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapThreeExternalStatus)
}