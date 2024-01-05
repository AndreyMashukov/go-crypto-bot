package service

// #cgo pkg-config: python-3.11.6
// #include <Python.h>
import "C"
import (
	"archive/zip"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	ExchangeModel "gitlab.com/open-soft/go-crypto-bot/exchange_context/model"
	ExchangeRepository "gitlab.com/open-soft/go-crypto-bot/exchange_context/repository"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

type KlineCSV struct {
	OpenTime                 string
	Open                     string
	High                     string
	Low                      string
	Close                    string
	Volume                   string
	CloseTime                string
	QuoteAssetVolume         string
	NumberOfTrades           string
	TakerBuyBaseAssetVolume  string
	TakerBuyQuoteAssetVolume string
	Ignore                   string
}

func (k *KlineCSV) GetOpenTime() int64 {
	time, _ := strconv.ParseInt(k.OpenTime, 10, 64)

	return time
}

func (k *KlineCSV) GetCloseTime() int64 {
	time, _ := strconv.ParseInt(k.CloseTime, 10, 64)

	return time
}

func (k *KlineCSV) UnmarshalCSV(csv string) (err error) {
	panic(csv)

	return err
}

func ReadCSV(filePath string) [][]string {
	defer os.Remove(filePath)
	in, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer in.Close()

	csvReader := csv.NewReader(in)
	records, err := csvReader.ReadAll()
	if err != nil {
		panic(err)
	}

	return records
}

type TradeCSV struct {
	TradeId      string
	Price        string
	Qty          string
	Time         string
	IsBuyerMaker string
}

func (c *TradeCSV) GetOperation() string {
	if c.IsBuyerMaker == "True" {
		return "SELL"
	}

	return "BUY"
}

func (c *TradeCSV) GetVolume() float64 {
	quantity, _ := strconv.ParseFloat(c.Qty, 64)
	price, _ := strconv.ParseFloat(c.Price, 64)

	return price * quantity
}

func (c *TradeCSV) GetTime() int64 {
	time, _ := strconv.ParseInt(c.Time, 10, 64)

	return time
}

func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func Unzip(path string) (string, error) {
	archive, err := zip.OpenReader(path)
	defer archive.Close()
	defer os.Remove(path)

	if err != nil {
		return "", err
	}

	for _, f := range archive.File {
		filePath := f.Name
		fmt.Println("unzipping file ", filePath)

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", err
		}

		fileInArchive, err := f.Open()
		if err != nil {
			return "", err
		}

		_, err = io.Copy(dstFile, fileInArchive)
		if err != nil {
			return "", err
		}

		dstFile.Close()
		fileInArchive.Close()
		return filePath, nil
	}

	return "", errors.New("not found")
}

func PrepareDataset(symbol string) (string, error) {
	datasetPath := fmt.Sprintf("/go/src/app/datasets/dataset_%s.csv", symbol)
	csvFile, err := os.Create(datasetPath)

	if err != nil {
		return "", err
	}

	csvWriter := csv.NewWriter(csvFile)

	kLines := make([]KlineCSV, 0)
	tradesPath := fmt.Sprintf("%s-trades.csv.zip", symbol)

	dateString := time.Now().UTC().Add(time.Duration(-36) * time.Hour).Format("2006-01-02")

	err = DownloadFile(tradesPath, fmt.Sprintf("https://data.binance.vision/data/spot/daily/trades/%s/%s-trades-%s.zip",
		symbol,
		symbol,
		dateString,
	))

	if err != nil {
		return "", err
	}

	unzipped, err := Unzip(tradesPath)
	if err != nil {
		return "", err
	}
	trades := ReadCSV(unzipped)

	kLinesPath := fmt.Sprintf("%s-1m.csv.zip", symbol)
	err = DownloadFile(kLinesPath, fmt.Sprintf("https://data.binance.vision/data/spot/daily/klines/%s/1m/%s-1m-%s.zip",
		symbol,
		symbol,
		dateString,
	))
	if err != nil {
		return "", err
	}

	tradeIndex := 0

	unzipped, err = Unzip(kLinesPath)
	if err != nil {
		return "", err
	}
	for _, record := range ReadCSV(unzipped) {
		kline := KlineCSV{
			OpenTime:         record[0],
			Open:             record[1],
			High:             record[2],
			Low:              record[3],
			Close:            record[4],
			Volume:           record[5],
			CloseTime:        record[6],
			QuoteAssetVolume: record[7],
			NumberOfTrades:   record[8],
		}
		kLines = append(kLines, kline)

		sellVolume := 0.00
		buyVolume := 0.00

		seen := false
		tradeAmount := 0

		for index, trade := range trades {
			tradeCSV := TradeCSV{
				TradeId:      trade[0],
				Price:        trade[1],
				Qty:          trade[2],
				Time:         trade[4],
				IsBuyerMaker: trade[5],
			}

			if tradeCSV.GetTime() > kline.GetOpenTime() && tradeCSV.GetTime() <= kline.GetCloseTime() {
				if tradeCSV.GetOperation() == "BUY" {
					buyVolume += tradeCSV.GetVolume()
				} else {
					sellVolume += tradeCSV.GetVolume()
				}
				tradeAmount++
				seen = true
				continue
			}

			if seen {
				tradeIndex = index
				trades = trades[tradeIndex:]
				break
			}
		}

		row := []string{
			kline.Open,
			kline.High,
			kline.Low,
			kline.Close,
			kline.Volume,
			fmt.Sprintf("%f", sellVolume),
			fmt.Sprintf("%f", buyVolume),
		}
		_ = csvWriter.Write(row)
		csvWriter.Flush()

		// trade Id	price	qty	quoteQty	time	isBuyerMaker	isBestMatch
		log.Printf(
			"[%s] Sell Volume: %f, Buy volume: %f, Close = %s",
			symbol,
			sellVolume,
			buyVolume,
			kline.Close,
		)
	}

	csvWriter.Flush()

	csvFile.Close()

	return datasetPath, nil
}

type PythonMLBridge struct {
	ExchangeRepository *ExchangeRepository.ExchangeRepository
	Mutex              sync.RWMutex
	RDB                *redis.Client
	Ctx                *context.Context
	CurrentBot         *ExchangeModel.Bot
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

	datasetPath, err := PrepareDataset(symbol)
	if err != nil {
		return err
	}

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
	var predictedPrice float64

	predictedPriceCacheKey := fmt.Sprintf("predicted-price-%s-%d", symbol, p.CurrentBot.Id)
	predictedPriceCached := p.RDB.Get(*p.Ctx, predictedPriceCacheKey).Val()

	if len(predictedPriceCached) > 0 {
		_ = json.Unmarshal([]byte(predictedPriceCached), &predictedPrice)
		return predictedPrice, nil
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

	buyVolume, sellVolume := p.ExchangeRepository.GetTradeVolumes(*kLine)

	resultPath := p.getResultFilePath(symbol)

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

	encoded, _ := json.Marshal(result)
	p.RDB.Set(*p.Ctx, predictedPriceCacheKey, string(encoded), time.Second*2)

	return result, nil
}
