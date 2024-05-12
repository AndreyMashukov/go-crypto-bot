package ml

// #cgo pkg-config: python-3.11.6
// #include <Python.h>
import "C"
import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gitlab.com/open-soft/go-crypto-bot/src/model"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"gitlab.com/open-soft/go-crypto-bot/src/utils"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

type PythonMLBridge struct {
	DataSetBuilder     *DataSetBuilder
	ExchangeRepository *repository.ExchangeRepository
	SwapRepository     *repository.SwapRepository
	TimeService        *utils.TimeHelper
	Mutex              *sync.RWMutex
	LearnLock          *sync.RWMutex
	RDB                *redis.Client
	Ctx                *context.Context
	CurrentBot         *model.Bot
	Learning           bool
}

func (p *PythonMLBridge) getModelFilePath(symbol string) string {
	return fmt.Sprintf("/go/src/app/models/lin_model_%s.pkl", symbol)
}

func (p *PythonMLBridge) getResultFilePath(symbol string) string {
	return fmt.Sprintf("/go/src/app/results/result_%s.txt", symbol)
}

func (p *PythonMLBridge) Initialize() {
	C.Py_Initialize()

	fmt.Println("-----------------------------------")
	fmt.Println("Python version via py C api:")
	fmt.Println(C.GoString(C.Py_GetVersion()))
	fmt.Println("-----------------------------------")

	pyCode := `
import pandas as pd
import pickle
import numpy as np
import joblib
import sklearn
from sklearn.model_selection import train_test_split
from sklearn.linear_model import LinearRegression
from sklearn.metrics import mean_squared_error
`
	pyCodeC := C.CString(pyCode)
	defer C.free(unsafe.Pointer(pyCodeC))
	C.PyRun_SimpleString(pyCodeC)
	p.Mutex = &sync.RWMutex{}
	p.LearnLock = &sync.RWMutex{}
}

func (p *PythonMLBridge) Finalize() {
	C.Py_Finalize()
}

func (p *PythonMLBridge) IsLearning() bool {
	p.LearnLock.Lock()
	isLearning := p.Learning
	p.LearnLock.Unlock()
	return isLearning
}

func (p *PythonMLBridge) setLearning(value bool) {
	p.LearnLock.Lock()
	p.Learning = value
	p.LearnLock.Unlock()
}

func (p *PythonMLBridge) getPythonCode(symbol string, datasetPath string) string {
	resultPath := p.getResultFilePath(symbol)
	modelFilePath := p.getModelFilePath(symbol)

	return fmt.Sprintf(string([]byte(`
price_dataset = pd.read_csv(
    filepath_or_buffer='%s',
    names=[
        "order_book_buy_first_qty",
        "order_book_sell_first_qty",
        "order_book_buy_qty_sum",
        "order_book_sell_qty_sum",
        "order_book_buy_volume_sum",
        "order_book_sell_volume_sum",
        "secondary_price",
        "close"
    ]
)
X = pd.DataFrame(
    np.c_[
        price_dataset['order_book_buy_first_qty'], 
        price_dataset['order_book_sell_first_qty'], 
        price_dataset['order_book_buy_qty_sum'], 
        price_dataset['order_book_sell_qty_sum'], 
        price_dataset['order_book_buy_volume_sum'], 
        price_dataset['order_book_sell_volume_sum'], 
        price_dataset['secondary_price'], 
    ], columns = [
        "order_book_buy_first_qty",
        "order_book_sell_first_qty",
        "order_book_buy_qty_sum",
        "order_book_sell_qty_sum",
        "order_book_buy_volume_sum",
        "order_book_sell_volume_sum",
        "secondary_price",
    ])
Y = price_dataset['close']

X_train, X_test, Y_train, Y_test = train_test_split(X, Y, test_size = 0.3, random_state=5)

lin_model = LinearRegression()
lin_model.fit(X_train, Y_train)
y_train_predict = lin_model.predict(X_train)
rmse = (np.sqrt(mean_squared_error(Y_train, y_train_predict)))
r2 = sklearn.metrics.r2_score(Y_train, y_train_predict)

print("The model performance for training set")
print("--------------------------------------")
print('RMSE is {}'.format(rmse))
print('R2 score is {}'.format(r2))
print("\n")

# model evaluation for testing set
y_test_predict = lin_model.predict(X_test)
rmse = (np.sqrt(mean_squared_error(Y_test, y_test_predict)))
r2 = sklearn.metrics.r2_score(Y_test, y_test_predict)

print("The model performance for testing set")
print("--------------------------------------")
print('RMSE is {}'.format(rmse))
print('R2 score is {}'.format(r2))
model_file_path = '%s'
with open(model_file_path, 'wb') as f:
    pickle.dump(lin_model, f)

lr2 = joblib.load(model_file_path)

y_train_predict = lr2.predict(X_train.loc[[X_train.index[0]]])
print(y_train_predict)

result_path = '%s'
with open(result_path, 'w') as out:
    print('{:.10f}'.format(y_train_predict[0]), file=out)
`)), datasetPath, modelFilePath, resultPath)
}

func (p *PythonMLBridge) LearnModel(symbol string) error {
	p.setLearning(true)
	defer p.setLearning(false)

	datasetPath, err := p.DataSetBuilder.PrepareDataset(symbol)
	if err != nil {
		log.Printf("[%s] ML dataset error: %s", symbol, err.Error())
		return err
	}

	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	defer os.Remove(datasetPath)

	resultPath := p.getResultFilePath(symbol)
	pyCode := p.getPythonCode(symbol, datasetPath)

	pyCodeC := C.CString(pyCode)
	defer C.free(unsafe.Pointer(pyCodeC))
	C.PyRun_SimpleString(pyCodeC)

	fileContent, err := os.ReadFile(resultPath)

	if err != nil {
		return err
	}

	text := string(fileContent)
	log.Printf("[%s] Result is: %s", symbol, text)

	return nil
}

func (p *PythonMLBridge) GetPythonPredictCode(params model.TradePricePredictParams) string {
	resultPath := p.getResultFilePath(params.Symbol)
	modelFilePath := p.getModelFilePath(params.Symbol)

	return fmt.Sprintf(string([]byte(`
test = pd.DataFrame(np.c_[
		%.10f, 
		%.10f, 
		%.10f, 
		%.10f, 
		%.10f, 
		%.10f, 
		%.10f, 
	], columns = [
        "order_book_buy_first_qty",
        "order_book_sell_first_qty",
        "order_book_buy_qty_sum",
        "order_book_sell_qty_sum",
        "order_book_buy_volume_sum",
        "order_book_sell_volume_sum",
        "secondary_price",
	]
)

lr2 = joblib.load('%s')
a = lr2.predict(test)
result_path = '%s'
with open(result_path, 'w') as out:
    print('{:.10f}'.format(a[0]), file=out)
`)),
		params.OrderBookBuyFirstQty,
		params.OrderBookSellFirstQty,
		params.OrderBookBuyQtySum,
		params.OrderBookSellQtySum,
		params.OrderBookBuyVolumeSum,
		params.OrderBookSellVolumeSum,
		params.SecondaryPrice,
		modelFilePath,
		resultPath,
	)
}

func (p *PythonMLBridge) Predict(symbol string) (float64, error) {
	if p.IsLearning() {
		return 0.00, errors.New("learning in the process")
	}

	modelFilePath := p.getModelFilePath(symbol)
	_, err := os.Stat(modelFilePath)
	if err != nil {
		return 0.00, err
	}

	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	kLine := p.ExchangeRepository.GetCurrentKline(symbol)

	if kLine == nil {
		return 0.00, errors.New("price is unknown")
	}

	secondaryKline := p.ExchangeRepository.GetCurrentKline(p.DataSetBuilder.GetSecondarySymbol(symbol))

	if secondaryKline == nil {
		return 0.00, errors.New("secondary price is unknown")
	}

	depth := p.ExchangeRepository.GetDepth(symbol, 500)

	if depth.IsEmpty() {
		return 0.00, errors.New("order depth is empty")
	}

	resultPath := p.getResultFilePath(symbol)
	pyCode := p.GetPythonPredictCode(model.TradePricePredictParams{
		Symbol:                 symbol,
		OrderBookBuyFirstQty:   depth.GetFirstBuyQty(),
		OrderBookSellFirstQty:  depth.GetFirstSellQty(),
		OrderBookBuyQtySum:     depth.GetQtySumBid(),
		OrderBookSellQtySum:    depth.GetQtySumAsk(),
		OrderBookBuyVolumeSum:  depth.GetBidVolume(),
		OrderBookSellVolumeSum: depth.GetAskVolume(),
		SecondaryPrice:         secondaryKline.Close,
	})

	pyCodeC := C.CString(pyCode)
	defer C.free(unsafe.Pointer(pyCodeC))
	C.PyRun_SimpleString(pyCodeC)

	fileContent, err := os.ReadFile(resultPath)
	if err != nil {
		log.Fatal(err)
	}

	text := string(fileContent)
	result, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	if err != nil {
		return 0.00, errors.New("result reading error")
	}

	return result, nil
}

func (p *PythonMLBridge) StartAutoLearn() {
	symbols := make([]string, 0)
	for _, tradeLimit := range p.ExchangeRepository.GetTradeLimits() {
		symbols = append(symbols, tradeLimit.Symbol)
	}
	if !slices.Contains(symbols, "BTCUSDT") {
		symbols = append(symbols, "BTCUSDT")
	}
	if !slices.Contains(symbols, "ETHUSDT") {
		symbols = append(symbols, "ETHUSDT")
	}

	wg := sync.WaitGroup{}
	for _, symbol := range symbols {
		wg.Add(1)
		go func(s string) {
			for {
				// todo: write to database and read from database
				err := p.LearnModel(s)
				wg.Done()
				if err != nil {
					log.Printf("[%s] %s", s, err.Error())
					p.TimeService.WaitSeconds(60)
					wg.Add(1) // just to handle negative counter
					continue
				}
				p.TimeService.WaitSeconds(3600)
				wg.Add(1) // just to handle negative counter
			}
		}(symbol)
	}

	wg.Wait()
	log.Printf("ML autolearn enabled, all models processed")
}
