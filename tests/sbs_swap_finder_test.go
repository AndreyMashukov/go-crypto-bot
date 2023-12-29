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

	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "ETH", "USDT").Return(options0)
	exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "BTC", "USDT").Return(options1)
	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "XRP", "USDT").Return(options2)

	sbsFinder := service.SBSSwapFinder{
		ExchangeRepository: exchangeRepoMock,
		Formatter:          &service.Formatter{},
	}

	chain := sbsFinder.Find("ETH").BestChain
	assertion := assert.New(t)
	assertion.Equal(3.5, chain.Percent.Value())
	assertion.Equal("SBS", chain.Type)
	assertion.Equal("ETH sell-> BTC buy-> XRP sell-> ETH", chain.Title)
	assertion.Equal("ETHBTC", chain.SwapOne.Symbol)
	assertion.Equal(0.05358, chain.SwapOne.Price)
	assertion.Equal("XRPBTC", chain.SwapTwo.Symbol)
	assertion.Equal(0.00001428, chain.SwapTwo.Price)
	assertion.Equal("XRPETH", chain.SwapThree.Symbol)
	assertion.Equal(0.0002775, chain.SwapThree.Price)
	// base amount is 100
	assertion.Greater(100*chain.SwapOne.Price/chain.SwapTwo.Price*chain.SwapThree.Price, 104.10)

	// validate
	swapRepoMock := new(SwapRepositoryMock)

	swapRepoMock.On("GetSwapPairBySymbol", "ETHBTC").Return(options0[0], nil)
	swapRepoMock.On("GetSwapPairBySymbol", "XRPBTC").Return(options1[0], nil)
	swapRepoMock.On("GetSwapPairBySymbol", "XRPETH").Return(options2[1], nil)

	binance := new(ExchangePriceAPIMock)

	binance.On("GetKLinesCached", "ETHBTC", "1d", int64(14)).Return([]model.KLine{
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
	})
	binance.On("GetKLinesCached", "XRPBTC", "1d", int64(14)).Return([]model.KLine{
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
	})
	binance.On("GetKLinesCached", "XRPETH", "1d", int64(14)).Return([]model.KLine{
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
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

	balanceServiceMock.On("GetAssetBalance", "ETH", false).Times(1).Return(order.ExecutedQuantity, nil)
	swapRepoMock.On("UpdateSwapAction", mock.Anything).Return(nil)
	swapRepoMock.On("GetActiveSwapAction", order).Return(model.SwapAction{
		Id:              999,
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

	binanceMock.On("LimitOrder", "ETHBTC", 100.00, 0.05358, "SELL", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             int64(12),
		Symbol:              "ETHBTC",
		ExecutedQty:         0.00,
		OrigQty:             100.00,
		Price:               0.05358,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("QueryOrder", "ETHBTC", int64(12)).Times(1).Return(model.BinanceOrder{
		Status:              "PARTIALLY_FILLED",
		OrderId:             int64(12),
		ExecutedQty:         80.00,
		OrigQty:             100.00,
		Symbol:              "ETHBTC",
		Price:               0.05358,
		Side:                "SELL",
		CummulativeQuoteQty: 80.00 * 0.05358,
	}, nil)
	binanceMock.On("QueryOrder", "ETHBTC", int64(12)).Times(2).Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             int64(12),
		ExecutedQty:         100.00,
		OrigQty:             100.00,
		Symbol:              "ETHBTC",
		Price:               0.05358,
		Side:                "SELL",
		CummulativeQuoteQty: 100.00 * 0.05358,
	}, nil)

	btcInitialBalance := 1.3455
	balanceServiceMock.On("GetAssetBalance", "BTC", false).Return(5.358+btcInitialBalance, nil)

	binanceMock.On("LimitOrder", "XRPBTC", 375210.00, 0.00001428, "BUY", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             int64(13),
		Symbol:              "XRPBTC",
		ExecutedQty:         0.00,
		OrigQty:             375210.00,
		Price:               0.00001428,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("QueryOrder", "XRPBTC", int64(13)).Times(1).Return(model.BinanceOrder{
		Status:              "PARTIALLY_FILLED",
		OrderId:             int64(13),
		Symbol:              "XRPBTC",
		ExecutedQty:         125210.00,
		OrigQty:             375210.00,
		Price:               0.00001428,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00001428 * 125210.00,
	}, nil)
	binanceMock.On("QueryOrder", "XRPBTC", int64(13)).Times(2).Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             int64(13),
		Symbol:              "XRPBTC",
		ExecutedQty:         375210.00,
		OrigQty:             375210.00,
		Price:               0.00001428,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00001428 * 375210.00,
	}, nil)

	xrpInitialBalance := 4000.00
	balanceServiceMock.On("GetAssetBalance", "XRP", false).Return(375212.00+xrpInitialBalance, nil)

	binanceMock.On("LimitOrder", "XRPETH", 375210.00, 0.0002775, "SELL", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             int64(14),
		Symbol:              "XRPETH",
		ExecutedQty:         0.00,
		OrigQty:             375210.00,
		Price:               0.0002775,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("QueryOrder", "XRPETH", int64(14)).Times(1).Return(model.BinanceOrder{
		Status:              "PARTIALLY_FILLED",
		OrderId:             int64(14),
		Symbol:              "XRPETH",
		ExecutedQty:         125210.00,
		OrigQty:             375210.00,
		Price:               0.0002775,
		Side:                "SELL",
		CummulativeQuoteQty: 0.0002775 * 125210.00,
	}, nil)
	binanceMock.On("QueryOrder", "XRPETH", int64(14)).Times(2).Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             int64(14),
		Symbol:              "XRPETH",
		ExecutedQty:         375210.00,
		OrigQty:             375210.00,
		Price:               0.0002775,
		Side:                "SELL",
		CummulativeQuoteQty: 0.0002775 * 375210.00,
	}, nil)

	orderRepositoryMock.On("Update", mock.Anything).Once().Return(nil)
	swapRepoMock.On("UpdateSwapAction", mock.Anything).Return(nil)
	balanceServiceMock.On("InvalidateBalanceCache", "ETH").Once()
	ethInitialBalance := 10.99
	balanceServiceMock.On("GetAssetBalance", "ETH", false).Times(2).Return(104.00+ethInitialBalance, nil)

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

	assertion.Equal(104.120775, *swapRepoMock.swapAction.EndQuantity)
	assertion.Equal(int64(12), *swapRepoMock.swapAction.SwapOneExternalId)
	assertion.Equal("ETHBTC", swapRepoMock.swapAction.SwapOneSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapOneExternalStatus)
	assertion.Equal(int64(13), *swapRepoMock.swapAction.SwapTwoExternalId)
	assertion.Equal("XRPBTC", swapRepoMock.swapAction.SwapTwoSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapTwoExternalStatus)
	assertion.Equal(int64(14), *swapRepoMock.swapAction.SwapThreeExternalId)
	assertion.Equal("XRPETH", swapRepoMock.swapAction.SwapThreeSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapThreeExternalStatus)
}

func TestSwapSellBuySellForceSwap(t *testing.T) {
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

	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "ETH", "USDT").Return(options0)
	exchangeRepoMock.On("GetSwapPairsByQuoteAsset", "BTC", "USDT").Return(options1)
	exchangeRepoMock.On("GetSwapPairsByBaseAsset", "XRP", "USDT").Return(options2)

	sbsFinder := service.SBSSwapFinder{
		ExchangeRepository: exchangeRepoMock,
		Formatter:          &service.Formatter{},
	}

	chain := sbsFinder.Find("ETH").BestChain
	assertion := assert.New(t)
	assertion.Equal(3.5, chain.Percent.Value())
	assertion.Equal("SBS", chain.Type)
	assertion.Equal("ETH sell-> BTC buy-> XRP sell-> ETH", chain.Title)
	assertion.Equal("ETHBTC", chain.SwapOne.Symbol)
	assertion.Equal(0.05358, chain.SwapOne.Price)
	assertion.Equal("XRPBTC", chain.SwapTwo.Symbol)
	assertion.Equal(0.00001428, chain.SwapTwo.Price)
	assertion.Equal("XRPETH", chain.SwapThree.Symbol)
	assertion.Equal(0.0002775, chain.SwapThree.Price)
	// base amount is 100
	assertion.Greater(100*chain.SwapOne.Price/chain.SwapTwo.Price*chain.SwapThree.Price, 104.10)

	// validate
	swapRepoMock := new(SwapRepositoryMock)

	swapRepoMock.On("GetSwapPairBySymbol", "ETHBTC").Return(options0[0], nil)
	swapRepoMock.On("GetSwapPairBySymbol", "XRPBTC").Return(options1[0], nil)
	swapRepoMock.On("GetSwapPairBySymbol", "XRPETH").Return(options2[1], nil)

	binance := new(ExchangePriceAPIMock)

	binance.On("GetKLinesCached", "ETHBTC", "1d", int64(14)).Return([]model.KLine{
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
		{
			High: 0.05359,
		},
	})
	binance.On("GetKLinesCached", "XRPBTC", "1d", int64(14)).Return([]model.KLine{
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
		{
			Low: 0.00001425,
		},
	})
	binance.On("GetKLinesCached", "XRPETH", "1d", int64(14)).Return([]model.KLine{
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
		},
		{
			High: 0.0002776,
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

	swapRepoMock.On("GetSwapPairBySymbol", "ETHBTC").Times(2).Return(model.SwapPair{
		Symbol:      "ETHBTC",
		SellPrice:   0.0536,
		MinNotional: 0.0001,
		MinQuantity: 0.0001,
		MinPrice:    0.00001,
	}, nil)
	swapRepoMock.On("GetSwapPairBySymbol", "XRPBTC").Times(2).Return(model.SwapPair{
		Symbol:      "XRPBTC",
		BuyPrice:    0.00001426,
		MinNotional: 0.0001,
		MinQuantity: 1,
		MinPrice:    0.00000001,
	}, nil)
	swapRepoMock.On("GetSwapPairBySymbol", "XRPETH").Times(2).Return(model.SwapPair{
		Symbol:      "XRPETH",
		SellPrice:   0.0002785,
		MinNotional: 0.001,
		MinQuantity: 1,
		MinPrice:    0.0000001,
	}, nil)

	balanceServiceMock.On("GetAssetBalance", "ETH", false).Times(1).Return(order.ExecutedQuantity, nil)
	swapRepoMock.On("UpdateSwapAction", mock.Anything).Return(nil)
	swapRepoMock.On("GetActiveSwapAction", order).Return(model.SwapAction{
		Id:              999,
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

	binanceMock.On("LimitOrder", "ETHBTC", 100.00, 0.05358, "SELL", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             int64(12),
		Symbol:              "ETHBTC",
		ExecutedQty:         0.00,
		OrigQty:             100.00,
		Price:               0.05358,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("QueryOrder", "ETHBTC", int64(12)).Times(1).Return(model.BinanceOrder{
		Status:              "PARTIALLY_FILLED",
		OrderId:             int64(12),
		ExecutedQty:         80.00,
		OrigQty:             100.00,
		Symbol:              "ETHBTC",
		Price:               0.05358,
		Side:                "SELL",
		CummulativeQuoteQty: 80.00 * 0.05358,
	}, nil)
	binanceMock.On("QueryOrder", "ETHBTC", int64(12)).Times(2).Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             int64(12),
		ExecutedQty:         100.00,
		OrigQty:             100.00,
		Symbol:              "ETHBTC",
		Price:               0.05358,
		Side:                "SELL",
		CummulativeQuoteQty: 100.00 * 0.05358,
	}, nil)

	btcInitialBalance := 1.3455
	balanceServiceMock.On("GetAssetBalance", "BTC", false).Return(5.358+btcInitialBalance, nil)

	binanceMock.On("LimitOrder", "XRPBTC", 375210.00, 0.00001428, "BUY", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             int64(13),
		Symbol:              "XRPBTC",
		ExecutedQty:         0.00,
		OrigQty:             375210.00,
		Price:               0.00001428,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("QueryOrder", "XRPBTC", int64(13)).Times(1).Return(model.BinanceOrder{
		Status:              "PARTIALLY_FILLED",
		OrderId:             int64(13),
		Symbol:              "XRPBTC",
		ExecutedQty:         125210.00,
		OrigQty:             375210.00,
		Price:               0.00001428,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00001428 * 125210.00,
	}, nil)
	binanceMock.On("QueryOrder", "XRPBTC", int64(13)).Times(2).Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             int64(13),
		Symbol:              "XRPBTC",
		ExecutedQty:         375210.00,
		OrigQty:             375210.00,
		Price:               0.00001428,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00001428 * 375210.00,
	}, nil)

	xrpInitialBalance := 4000.00
	balanceServiceMock.On("GetAssetBalance", "XRP", false).Return(375212.00+xrpInitialBalance, nil)

	binanceMock.On("LimitOrder", "XRPETH", 375210.00, 0.0002775, "SELL", "GTC").Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             int64(14),
		Symbol:              "XRPETH",
		ExecutedQty:         0.00,
		OrigQty:             375210.00,
		Price:               0.0002775,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("QueryOrder", "XRPETH", int64(14)).Times(1).Return(model.BinanceOrder{
		Status:              "NEW",
		OrderId:             int64(14),
		Symbol:              "XRPETH",
		ExecutedQty:         0.00,
		OrigQty:             375210.00,
		Price:               0.0002775,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("CancelOrder", "XRPETH", int64(14)).Return(model.BinanceOrder{
		Status:              "CANCELED",
		OrderId:             int64(14),
		Symbol:              "XRPETH",
		ExecutedQty:         0.00,
		OrigQty:             375210.00,
		Price:               0.0002775,
		Side:                "SELL",
		CummulativeQuoteQty: 0.00,
	}, nil)
	binanceMock.On("LimitOrder", "XRPETH", 375210.00, 0.0002784, "SELL", "IOC").Return(model.BinanceOrder{
		Status:              "FILLED",
		OrderId:             int64(14),
		Symbol:              "XRPETH",
		ExecutedQty:         375210.00,
		OrigQty:             375210.00,
		Price:               0.0002775,
		Side:                "SELL",
		CummulativeQuoteQty: 104.46, // 375210.00 * 0.0002784
	}, nil)

	orderRepositoryMock.On("Update", mock.Anything).Once().Return(nil)
	swapRepoMock.On("UpdateSwapAction", mock.Anything).Return(nil)
	balanceServiceMock.On("InvalidateBalanceCache", "ETH").Once()
	ethInitialBalance := 10.99
	balanceServiceMock.On("GetAssetBalance", "ETH", false).Times(2).Return(104.00+ethInitialBalance, nil)

	timeServiceMock := new(TimeServiceMock)
	timeServiceMock.On("WaitSeconds", int64(5)).Times(3)
	timeServiceMock.On("WaitSeconds", int64(7)).Times(3)
	timeServiceMock.On("GetNowDiffMinutes", mock.Anything).Return(50.00)

	executor := service.SwapExecutor{
		SwapRepository:  swapRepoMock,
		OrderRepository: orderRepositoryMock,
		BalanceService:  balanceServiceMock,
		Binance:         binanceMock,
		TimeService:     timeServiceMock,
		Formatter:       &service.Formatter{},
	}

	executor.Execute(order)

	assertion.Equal(104.46, *swapRepoMock.swapAction.EndQuantity)
	assertion.Equal(int64(12), *swapRepoMock.swapAction.SwapOneExternalId)
	assertion.Equal("ETHBTC", swapRepoMock.swapAction.SwapOneSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapOneExternalStatus)
	assertion.Equal(int64(13), *swapRepoMock.swapAction.SwapTwoExternalId)
	assertion.Equal("XRPBTC", swapRepoMock.swapAction.SwapTwoSymbol)
	assertion.Equal("FILLED", *swapRepoMock.swapAction.SwapTwoExternalStatus)
	assertion.Equal(int64(14), *swapRepoMock.swapAction.SwapThreeExternalId)
	assertion.Equal("XRPETH", swapRepoMock.swapAction.SwapThreeSymbol)
	assertion.Equal("FILLED_FORCE", *swapRepoMock.swapAction.SwapThreeExternalStatus)
}
