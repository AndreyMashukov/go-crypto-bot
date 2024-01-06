package service

// #cgo pkg-config: python-3.11.6
// #include <Python.h>
import "C"
import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

type PythonMLBridge struct {
	DataSetBuilder     *DataSetBuilder
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	Mutex              sync.RWMutex
	RDB                *redis.Client
	Ctx                *context.Context
	CurrentBot         *ExchangeModel.Bot
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
	p.Mutex = sync.RWMutex{}
}

func (p *PythonMLBridge) Finalize() {
	C.Py_Finalize()
}

func (p *PythonMLBridge) setLearning(value bool) {
	p.Learning = value
}

func (p *PythonMLBridge) getPythonCode(symbol string, datasetPath string) string {
	resultPath := p.getResultFilePath(symbol)
	modelFilePath := p.getModelFilePath(symbol)

	return fmt.Sprintf(string([]byte(`
price_dataset = pd.read_csv(
    filepath_or_buffer='%s',
    names=["open", "high", "low", "close", "volume", "sell_vol", "buy_vol"]
)
X = pd.DataFrame(np.c_[price_dataset['volume'], price_dataset['buy_vol'], price_dataset['sell_vol'], price_dataset['open'], price_dataset['low'],price_dataset['high']], columns = ['volume','buy_vol', 'sell_vol', 'open', 'low', 'high'])
Y = price_dataset['close']

X_train, X_test, Y_train, Y_test = train_test_split(X, Y, test_size = 0.2, random_state=5)

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

y_train_predict = lr2.predict(X_train.loc[[10]])
print(y_train_predict)
result_path = '%s'
with open(result_path, 'w') as out:
    print(y_train_predict[0], file=out)
`)), datasetPath, modelFilePath, resultPath)
}

func (p *PythonMLBridge) getPythonAltCoinCode(symbol string, datasetPath string) string {
	resultPath := p.getResultFilePath(symbol)
	modelFilePath := p.getModelFilePath(symbol)

	return fmt.Sprintf(string([]byte(`
price_dataset = pd.read_csv(
    filepath_or_buffer='%s',
    names=["open", "high", "low", "close", "volume", "sell_vol", "buy_vol", "btc_price"]
)
X = pd.DataFrame(np.c_[price_dataset['volume'], price_dataset['buy_vol'], price_dataset['sell_vol'], price_dataset['open'], price_dataset['low'],price_dataset['high'],price_dataset['btc_price']], columns = ['volume','buy_vol', 'sell_vol', 'open', 'low', 'high', 'btc_price'])
Y = price_dataset['close']

X_train, X_test, Y_train, Y_test = train_test_split(X, Y, test_size = 0.2, random_state=5)

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

y_train_predict = lr2.predict(X_train.loc[[10]])
print(y_train_predict)
result_path = '%s'
with open(result_path, 'w') as out:
    print(y_train_predict[0], file=out)
`)), datasetPath, modelFilePath, resultPath)
}

func (p *PythonMLBridge) LearnModel(symbol string) error {
	p.setLearning(true)
	defer p.setLearning(false)

	datasetPath, err := p.DataSetBuilder.PrepareDataset(symbol)
	if err != nil {
		return err
	}

	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	defer os.Remove(datasetPath)

	resultPath := p.getResultFilePath(symbol)

	var pyCode string
	if "BTCUSDT" == symbol {
		pyCode = p.getPythonCode(symbol, datasetPath)
	} else {
		pyCode = p.getPythonAltCoinCode(symbol, datasetPath)
	}

	pyCodeC := C.CString(pyCode)
	defer C.free(unsafe.Pointer(pyCodeC))
	C.PyRun_SimpleString(pyCodeC)

	fileContent, err := os.ReadFile(resultPath)
	if err != nil {
		log.Fatal(err)
	}

	text := string(fileContent)
	log.Printf("[%s] Result is: %s", symbol, text)

	return nil
}

func (p *PythonMLBridge) GetPythonPredictCode(kLine ExchangeModel.KLine) string {
	buyVolume, sellVolume := p.ExchangeRepository.GetTradeVolumes(kLine)
	resultPath := p.getResultFilePath(kLine.Symbol)
	modelFilePath := p.getModelFilePath(kLine.Symbol)

	return fmt.Sprintf(string([]byte(`
test = pd.DataFrame(np.c_[%f, %f, %f, %f, %f, %f], columns = ['volume','buy_vol', 'sell_vol', 'open', 'low', 'high'])

lr2 = joblib.load('%s')
a = lr2.predict(test)
result_path = '%s'
with open(result_path, 'w') as out:
    print(a[0], file=out)
`)),
		kLine.Volume,
		buyVolume,
		sellVolume,
		kLine.Open,
		kLine.Low,
		kLine.High,
		modelFilePath,
		resultPath,
	)
}

func (p *PythonMLBridge) GetPythonPredictAltCoinCode(kLine ExchangeModel.KLine, btcPrice float64) string {
	buyVolume, sellVolume := p.ExchangeRepository.GetTradeVolumes(kLine)
	resultPath := p.getResultFilePath(kLine.Symbol)
	modelFilePath := p.getModelFilePath(kLine.Symbol)

	return fmt.Sprintf(string([]byte(`
test = pd.DataFrame(np.c_[%f, %f, %f, %f, %f, %f, %f], columns = ['volume','buy_vol', 'sell_vol', 'open', 'low', 'high', 'btc_price'])

lr2 = joblib.load('%s')
a = lr2.predict(test)
result_path = '%s'
with open(result_path, 'w') as out:
    print(a[0], file=out)
`)),
		kLine.Volume,
		buyVolume,
		sellVolume,
		kLine.Open,
		kLine.Low,
		kLine.High,
		btcPrice,
		modelFilePath,
		resultPath,
	)
}

func (p *PythonMLBridge) Predict(symbol string) (float64, error) {
	if p.Learning {
		return 0.00, errors.New("learning in the process")
	}

	modelFilePath := p.getModelFilePath(symbol)
	_, err := os.Stat(modelFilePath)
	if err != nil {
		return 0.00, err
	}

	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	kLine := p.ExchangeRepository.GetLastKLine(symbol)

	if kLine == nil {
		return 0.00, errors.New("price is unknown")
	}

	resultPath := p.getResultFilePath(symbol)

	var pyCode string
	if "BTCUSDT" == symbol {
		pyCode = p.GetPythonPredictCode(*kLine)
	} else {
		btcKline := p.ExchangeRepository.GetLastKLine("BTCUSDT")
		if btcKline == nil {
			return 0.00, errors.New("BTC price is unknown")
		}

		pyCode = p.GetPythonPredictAltCoinCode(*kLine, btcKline.Close)
	}

	pyCodeC := C.CString(pyCode)
	defer C.free(unsafe.Pointer(pyCodeC))
	C.PyRun_SimpleString(pyCodeC)

	fileContent, err := os.ReadFile(resultPath)
	if err != nil {
		log.Fatal(err)
	}

	text := string(fileContent)
	result, _ := strconv.ParseFloat(strings.TrimSpace(text), 64)

	return result, nil
}
