package tests

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type BinanceMock struct {
	mock.Mock
}

func (m *BinanceMock) GetDepth(symbol string) (interface{}, error) {
	args := m.Called(symbol)
	return args.Get(0), args.Error(1)
}

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestBinanceMock(t *testing.T) {
	assert := assert.New(t)
	binanceMock := new(BinanceMock)
	binanceMock.On("GetDepth", "ABCDEF").Return(nil, errors.New("test_error"))

	depth, err := binanceMock.GetDepth("ABCDEF")
	assert.Nil(depth, "Depth is not nil!")
	assert.Equal(err.Error(), "test_error")
}
