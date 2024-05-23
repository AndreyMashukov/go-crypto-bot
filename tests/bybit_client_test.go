package tests

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/open-soft/go-crypto-bot/src/client"
	"testing"
)

func TestShouldRecoverExchangeOrderIfNotFetched(t *testing.T) {
	assertion := assert.New(t)

	httpClientMock := new(HttpClientMock)

	bybitClient := client.ByBit{
		HttpClient: httpClientMock,
		DSN:        "https://fake.url",
	}

	limitOrderResponse := []byte("{\"retMsg\": \"OK\", \"result\": {\"orderId\": \"111999\"}}")
	httpClientMock.On("Post", "https://fake.url/v5/order/create", mock.Anything, mock.Anything).Return(limitOrderResponse, nil)
	httpClientMock.On("Get", "https://fake.url/v5/order/realtime?category=spot&limit=1&orderId=111999&symbol=BTCUSDT&openOnly=0", mock.Anything).Return([]byte(""), errors.New("WTF!!!"))

	exchangeOrder, err := bybitClient.LimitOrder("BTCUSDT", 0.004, 50000.00, "BUY", "GTC")
	assertion.Nil(err)
	assertion.Equal("BTCUSDT", exchangeOrder.Symbol)
	assertion.Equal("111999", exchangeOrder.OrderId)
	assertion.Equal("NEW", exchangeOrder.Status)
	assertion.Equal("LIMIT", exchangeOrder.Type)
	assertion.Equal(float64(0.004), exchangeOrder.OrigQty)
	assertion.Equal(float64(0.00), exchangeOrder.ExecutedQty)
	assertion.Equal(float64(0.004)*float64(50000.00), exchangeOrder.CummulativeQuoteQty)
	assertion.Equal(int64(0), exchangeOrder.WorkingTime)
	assertion.Equal(int64(0), exchangeOrder.TransactTime)
	assertion.Equal(float64(50000.00), exchangeOrder.Price)
}
