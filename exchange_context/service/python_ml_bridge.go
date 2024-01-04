package service

// #cgo pkg-config: python-3.11.6
// #include <Python.h>
import "C"
import (
	"encoding/csv"
	"errors"
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/exchange_context/client"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

type PythonMLBridge struct {
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	Binance            *client.Binance
	Mutex              sync.RWMutex
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

func (p *PythonMLBridge) LearnModel(symbol string) error {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	history := p.Binance.GetKLines(symbol, "1m", 1000)

	trades := p.Binance.TradesAggregate(symbol, 1000)

	log.Printf(
		"Learn [%s] model: klines: %d, trades: %d",
		symbol,
		len(history),
		len(trades),
	)

	datasetPath := fmt.Sprintf("/go/src/app/datasets/dataset_%s.csv", symbol)

	csvFile, err := os.Create(datasetPath)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvWriter := csv.NewWriter(csvFile)

	for _, kLine := range history {
		buyVolume := 0.00
		sellVolume := 0.00

		for _, trade := range trades {
			if trade.Timestamp >= kLine.CloseTime-60000 && trade.Timestamp < kLine.CloseTime {
				if trade.GetOperation() == "BUY" {
					buyVolume += trade.Price * trade.Quantity
				} else {
					sellVolume += trade.Price * trade.Quantity
				}
			}
		}

		row := []string{
			kLine.Open,
			kLine.High,
			kLine.Low,
			kLine.Close,
			kLine.Volume,
			fmt.Sprintf("%f", sellVolume),
			fmt.Sprintf("%f", buyVolume),
		}
		_ = csvWriter.Write(row)
		csvWriter.Flush()
	}

	csvWriter.Flush()
	_ = csvFile.Close()

	resultPath := p.getResultFilePath(symbol)
	modelFilePath := p.getModelFilePath(symbol)

	pyCode := fmt.Sprintf(string([]byte(`
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

func (p *PythonMLBridge) Predict(symbol string) (float64, error) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	kLine := p.ExchangeRepository.GetLastKLine(symbol)

	if kLine == nil {
		return 0.00, errors.New("price is unknown")
	}

	buyVolume, sellVolume := p.ExchangeRepository.GetTradeVolumes(*kLine)

	resultPath := p.getResultFilePath(symbol)
	modelFilePath := p.getModelFilePath(symbol)

	pyCode := fmt.Sprintf(string([]byte(`
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
