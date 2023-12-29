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

	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "SOL", "USDT").Return(options3)
	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "ETH", "USDT").Return(options1)
	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "GBP", "USDT").Return(options4)

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

	binance := new(ExchangePriceAPIMock)

	binance.On("GetKLinesCached", "SOLETH", "1d", int64(14)).Return([]model.KLine{
		{
			High: 0.03373,
		},
		{
			High: 0.03373,
		},
		{
			High: 0.03373,
		},
		{
			High: 0.03373,
		},
		{
			High: 0.03373,
		},
		{
			High: 0.03373,
		},
		{
			High: 0.03373,
		},
		{
			High: 0.03373,
		},
	})
	binance.On("GetKLinesCached", "ETHGBP", "1d", int64(14)).Return([]model.KLine{
		{
			High: 1783.07,
		},
		{
			High: 1783.07,
		},
		{
			High: 1783.07,
		},
		{
			High: 1783.07,
		},
		{
			High: 1783.07,
		},
		{
			High: 1783.07,
		},
		{
			High: 1783.07,
		},
		{
			High: 1783.07,
		},
	})
	binance.On("GetKLinesCached", "SOLGBP", "1d", int64(14)).Return([]model.KLine{
		{
			Low: 52.47,
		},
		{
			Low: 52.47,
		},
		{
			Low: 52.47,
		},
		{
			Low: 52.47,
		},
		{
			Low: 52.47,
		},
		{
			Low: 52.47,
		},
		{
			Low: 52.47,
		},
		{
			Low: 52.47,
		},
	})

	swapChainBuilder := service.SwapChainBuilder{}
	validator := service.SwapValidator{
		Binance:        binance,
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
		Id:              995,
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

	binanceMock.On("LimitOrder", "SOLETH", 100.00, 0.03372, "SELL", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             int64(16),
		Symbol:              "SOLETH",
		ExecutedQty:         0.00,
		OrigQty:             100.00,
		Price:               0.03372,
		Side:                "SELL",
		CummulativeQuoteQty: 00.00,
	}, nil)
	binanceMock.On("QueryOrder", "SOLETH", int64(16)).Times(1).Return(model.BinanceOrder{
		Status:              "PARTIALLY_FILLED",
		OrderId:             int64(16),
		ExecutedQty:         80.00,
		OrigQty:             100.00,
		Symbol:              "SOLETH",
		Price:               0.03372,
		Side:                "SELL",
		CummulativeQuoteQty: 80 * 0.03372,
	}, nil)
	binanceMock.On("QueryOrder", "SOLETH", int64(16)).Times(2).Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             int64(16),
		ExecutedQty:         100.00,
		OrigQty:             100.00,
		Symbol:              "SOLETH",
		Price:               0.03372,
		Side:                "SELL",
		CummulativeQuoteQty: 100 * 0.03372,
	}, nil)

	ethInitialBalance := 2.99
	balanceServiceMock.On("GetAssetBalance", "ETH", false).Return(3.372+ethInitialBalance, nil)

	binanceMock.On("LimitOrder", "ETHGBP", 3.372, 1783.06, "SELL", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             int64(17),
		Symbol:              "ETHGBP",
		ExecutedQty:         0.00,
		OrigQty:             3.372,
		Price:               1783.06,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("QueryOrder", "ETHGBP", int64(17)).Times(1).Return(model.BinanceOrder{
		Status:              "PARTIALLY_FILLED",
		OrderId:             int64(17),
		Symbol:              "ETHGBP",
		ExecutedQty:         1.272,
		OrigQty:             3.372,
		Price:               1783.06,
		Side:                "SELL",
		CummulativeQuoteQty: 1.272 * 1783.06,
	}, nil)
	binanceMock.On("QueryOrder", "ETHGBP", int64(17)).Times(2).Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             int64(17),
		Symbol:              "ETHGBP",
		ExecutedQty:         3.372,
		OrigQty:             3.372,
		Price:               1783.06,
		Side:                "SELL",
		CummulativeQuoteQty: 3.372 * 1783.06,
	}, nil)
	gbpInitialBalance := 2300.99
	balanceServiceMock.On("GetAssetBalance", "GBP", false).Return(6012.476+gbpInitialBalance, nil)

	binanceMock.On("LimitOrder", "SOLGBP", 114.56, 52.48, "BUY", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             int64(18),
		Symbol:              "SOLGBP",
		ExecutedQty:         0.00,
		OrigQty:             114.56,
		Price:               52.48,
		Side:                "BUY",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("QueryOrder", "SOLGBP", int64(18)).Times(1).Return(model.BinanceOrder{
		Status:              "PARTIALLY_FILLED",
		OrderId:             int64(18),
		Symbol:              "SOLGBP",
		ExecutedQty:         12.00,
		OrigQty:             114.56,
		Price:               52.48,
		Side:                "BUY",
		CummulativeQuoteQty: 12.00 * 52.48,
	}, nil)
	binanceMock.On("QueryOrder", "SOLGBP", int64(18)).Times(2).Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             int64(18),
		Symbol:              "SOLGBP",
		ExecutedQty:         114.56,
		OrigQty:             114.56,
		Price:               52.48,
		Side:                "BUY",
		CummulativeQuoteQty: 114.56 * 52.48,
	}, nil)

	orderRepositoryMock.On("Update", mock.Anything).Once().Return(nil)
	swapRepoMock.On("UpdateSwapAction", mock.Anything).Return(nil)
	balanceServiceMock.On("InvalidateBalanceCache", "SOL").Once()
	solInitialBalance := 9.00
	balanceServiceMock.On("GetAssetBalance", "SOL", false).Times(2).Return(114.54+solInitialBalance, nil)

	timeServiceMock := new(TimeServiceMock)
	timeServiceMock.On("WaitSeconds", int64(5)).Times(3)
	timeServiceMock.On("WaitSeconds", int64(7)).Times(3)
	timeServiceMock.On("GetNowDiffMinutes", mock.Anything).Return(0.50)

	executor := service.SwapExecutor{
		SwapRepository:  swapRepoMock,
		OrderRepository: orderRepositoryMock,
		BalanceService:  balanceServiceMock,
		Binance:         binanceMock,
		TimeService:     timeServiceMock,
		Formatter:       &service.Formatter{},
	}

	executor.Execute(order)

	assertion.Equal(114.56, *swapRepoMock.swapAction.EndQuantity)
	assertion.Equal(int64(16), *swapRepoMock.swapAction.SwapOneExternalId)
	assertion.Equal("SOLETH", swapRepoMock.swapAction.SwapOneSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapOneExternalStatus)
	assertion.Equal(int64(17), *swapRepoMock.swapAction.SwapTwoExternalId)
	assertion.Equal("ETHGBP", swapRepoMock.swapAction.SwapTwoSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapTwoExternalStatus)
	assertion.Equal(int64(18), *swapRepoMock.swapAction.SwapThreeExternalId)
	assertion.Equal("SOLGBP", swapRepoMock.swapAction.SwapThreeSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapThreeExternalStatus)
}
