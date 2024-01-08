package service

import (
	"archive/zip"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
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

func (d *DataSetBuilder) ReadCSV(filePath string, remove bool) [][]string {
	if remove {
		defer os.Remove(filePath)
	}

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

type DataSetBuilder struct {
	ExcludeDependedDataset []string
	EthDependent           []string
	BtcDependent           []string
}

func (d *DataSetBuilder) DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("status code: %d", resp.StatusCode))
	}

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

func (d *DataSetBuilder) Unzip(path string, suffix string) (string, error) {
	archive, err := zip.OpenReader(path)

	if err != nil {
		log.Printf("[%s] error: %s", path, err.Error())
		return "", err
	}

	defer archive.Close()

	for _, f := range archive.File {
		filePath := f.Name
		dstFilePath := fmt.Sprintf("%s%s", filePath, suffix)
		//fmt.Println("unzipping file ", filePath)
		_ = os.Remove(dstFilePath)

		dstFile, err := os.OpenFile(dstFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
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
		return dstFilePath, nil
	}

	return "", errors.New("not found")
}

func (d *DataSetBuilder) GetSources(symbol string, dateString string) (string, string, string, error) {
	tradesPath := fmt.Sprintf("/go/src/app/datasets/%s-trades.csv.zip", symbol)
	kLinesPath := fmt.Sprintf("/go/src/app/datasets/%s-1m.csv.zip", symbol)
	_ = os.Remove(tradesPath)
	_ = os.Remove(kLinesPath)

	var err error = nil

	if dateString == "" {
		for i := 12; i <= 60; i = i + 12 {
			dateString = time.Now().UTC().Add(time.Duration(i*-1) * time.Hour).Format("2006-01-02")

			err = d.DownloadFile(tradesPath, fmt.Sprintf("https://data.binance.vision/data/spot/daily/trades/%s/%s-trades-%s.zip",
				symbol,
				symbol,
				dateString,
			))

			if err != nil {
				log.Printf("[%s] Dataset %s downloading error [%s]: %s", symbol, dateString, tradesPath, err.Error())
				continue
			}

			err = d.DownloadFile(kLinesPath, fmt.Sprintf("https://data.binance.vision/data/spot/daily/klines/%s/1m/%s-1m-%s.zip",
				symbol,
				symbol,
				dateString))
			if err != nil {
				log.Printf("[%s] Klines dataset %s downloading error [%s]: %s", symbol, dateString, kLinesPath, err.Error())
				continue
			}

			if !slices.Contains(d.ExcludeDependedDataset, symbol) {
				altSymbol := d.GetDependentOn(symbol)
				altBtcKLinesPath := fmt.Sprintf("/go/src/app/datasets/%s-1m-%s.csv.zip", altSymbol, symbol)
				_ = os.Remove(altBtcKLinesPath)

				err = d.DownloadFile(altBtcKLinesPath, fmt.Sprintf("https://data.binance.vision/data/spot/daily/klines/%s/1m/%s-1m-%s.zip",
					altSymbol,
					altSymbol,
					dateString))

				if err != nil {
					log.Printf("[%s] Klines dataset %s downloading error [%s]: %s", symbol, dateString, altBtcKLinesPath, err.Error())
					continue
				}
			}

			archive, existErr := zip.OpenReader(tradesPath)
			if existErr == nil {
				_ = archive.Close()
				log.Printf("[%s] downloaded: %s", symbol, tradesPath)
				break
			}
			_ = os.Remove(tradesPath)

			log.Printf(
				"[%s] Dataset downloading error [%s]: %s",
				symbol,
				dateString,
				existErr.Error(),
			)
		}
	} else {
		err = d.DownloadFile(tradesPath, fmt.Sprintf("https://data.binance.vision/data/spot/daily/trades/%s/%s-trades-%s.zip",
			symbol,
			symbol,
			dateString))

		if err != nil {
			log.Printf("[%s] Trades dataset %s download error: %s", symbol, dateString, err.Error())

			return "", "", dateString, err
		}

		err = d.DownloadFile(kLinesPath, fmt.Sprintf("https://data.binance.vision/data/spot/daily/klines/%s/1m/%s-1m-%s.zip",
			symbol,
			symbol,
			dateString))
		if err != nil {
			log.Printf("[%s] Klines dataset %s download error: %s", symbol, dateString, err.Error())

			return "", "", dateString, err
		}
	}

	if err != nil {
		return "", "", dateString, err
	}

	archive, existErr := zip.OpenReader(tradesPath)
	if existErr != nil {
		return "", "", dateString, existErr
	}
	_ = archive.Close()

	unzippedTrades, err := d.Unzip(tradesPath, "")
	defer os.Remove(tradesPath)
	if err != nil {
		return "", "", dateString, err
	}

	unzippedKLines, err := d.Unzip(kLinesPath, "")
	defer os.Remove(kLinesPath)

	if err != nil {
		return "", "", dateString, err
	}

	return unzippedTrades, unzippedKLines, dateString, nil
}

func (d *DataSetBuilder) WriteToCsv(symbol string, dateString string, csvWriter *csv.Writer, unzippedTrades string, unzippedKLines string) {
	dependentPriceMap := make(map[string]string)
	priceInDependent := make(map[string]string)

	if "BTCUSDT" != symbol {
		dependentSymbol := ""
		if slices.Contains(d.BtcDependent, strings.ReplaceAll(symbol, "USDT", "")) {
			dependentSymbol = "BTC"
		}
		if slices.Contains(d.EthDependent, strings.ReplaceAll(symbol, "USDT", "")) {
			dependentSymbol = "ETH"
		}

		if dependentSymbol == "" {
			log.Panic(fmt.Sprintf("unsupported symbol: %s", symbol))
		}

		dependency := fmt.Sprintf("%sUSDT", dependentSymbol)
		btcKLinesPath := fmt.Sprintf("/go/src/app/datasets/%s-1m-%s.csv.zip", dependency, symbol)
		_ = os.Remove(btcKLinesPath)

		err := d.DownloadFile(btcKLinesPath, fmt.Sprintf("https://data.binance.vision/data/spot/daily/klines/%s/1m/%s-1m-%s.zip",
			dependency,
			dependency,
			dateString,
		))

		if err == nil {
			unzippedBtcKLines, err := d.Unzip(btcKLinesPath, fmt.Sprintf(".alt-%s.csv", symbol))
			defer os.Remove(btcKLinesPath)

			if err == nil {
				for _, btcKline := range d.ReadCSV(unzippedBtcKLines, true) {
					dependentPriceMap[btcKline[6]] = btcKline[4]
				}
			}
		}

		altSymbol := d.GetDependentOn(symbol)

		altBtcKLinesPath := fmt.Sprintf("/go/src/app/datasets/%s-1m-%s.csv.zip", altSymbol, symbol)
		_ = os.Remove(altBtcKLinesPath)

		err = d.DownloadFile(altBtcKLinesPath, fmt.Sprintf("https://data.binance.vision/data/spot/daily/klines/%s/1m/%s-1m-%s.zip",
			altSymbol,
			altSymbol,
			dateString,
		))

		if err == nil {
			unzippedAltBtcKLines, err := d.Unzip(altBtcKLinesPath, fmt.Sprintf(".alt-btc-%s.csv", symbol))
			defer os.Remove(altBtcKLinesPath)

			if err == nil {
				for _, altBtcKline := range d.ReadCSV(unzippedAltBtcKLines, true) {
					priceInDependent[altBtcKline[6]] = altBtcKline[4]
				}
			}
		}
	}

	kLines := make([]KlineCSV, 0)
	tradeIndex := 0
	trades := d.ReadCSV(unzippedTrades, true)
	for _, record := range d.ReadCSV(unzippedKLines, true) {
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

		if "BTCUSDT" != symbol {
			row = append(row, dependentPriceMap[kline.CloseTime])

			value, ok := priceInDependent[kline.CloseTime]
			if slices.Contains(d.ExcludeDependedDataset, symbol) {
				value = "0.00"
			} else {
				if !ok {
					continue
				}
			}

			row = append(row, value)
		}

		_ = csvWriter.Write(row)
		csvWriter.Flush()

		//log.Printf(
		//	"[%s] Sell Volume: %f, Buy volume: %f, Close = %s",
		//	symbol,
		//	sellVolume,
		//	buyVolume,
		//	kline.Close,
		//)
	}
}

func (d *DataSetBuilder) GetHistoryDataset(symbol string) (string, error) {
	datasetPathHistory := fmt.Sprintf("/go/src/app/datasets/dataset_%s_history.csv", symbol)

	_, err := os.Stat(datasetPathHistory)
	if err == nil {
		return datasetPathHistory, nil
	}

	importantDates := []string{
		"2022-11-07",
		"2022-11-08",
		"2022-11-09",
		"2022-11-10",
		"2022-11-11",
		"2022-11-12",
		"2022-11-13",
		"2023-02-19",
		"2023-02-20",
		"2023-02-21",
		"2023-02-22",
		"2023-02-23",
		"2023-02-24",
		"2023-02-25",
		"2023-08-15",
		"2023-08-16",
		"2023-08-17",
		"2023-08-18",
		"2023-08-19",
		"2023-11-04",
		"2023-11-05",
		"2023-11-06",
		"2023-11-07",
		"2023-11-08",
		"2023-11-09",
		"2023-11-10",
		"2023-11-11",
		"2023-12-26",
		"2023-12-27",
		"2023-12-28",
		"2024-01-03",
	}

	csvFileHistory, err := os.Create(datasetPathHistory)

	if err != nil {
		return "", err
	}

	csvWriterHistory := csv.NewWriter(csvFileHistory)

	for _, dateString := range importantDates {
		unzippedTradesHistory, unzippedKLinesHistory, _, err := d.GetSources(symbol, dateString)
		if err == nil {
			d.WriteToCsv(symbol, dateString, csvWriterHistory, unzippedTradesHistory, unzippedKLinesHistory)
		} else {
			log.Printf("[%s] history dataset failed: %s", symbol, err.Error())
		}
	}

	csvWriterHistory.Flush()
	_ = csvFileHistory.Close()

	return datasetPathHistory, nil
}

func (d *DataSetBuilder) PrepareDataset(symbol string) (string, error) {
	datasetPath := fmt.Sprintf("/go/src/app/datasets/dataset_%s.csv", symbol)
	_ = os.Remove(datasetPath)
	csvFile, err := os.Create(datasetPath)

	if err != nil {
		return "", err
	}

	csvWriter := csv.NewWriter(csvFile)

	unzippedTrades, unzippedKLines, dateString, err := d.GetSources(symbol, "")

	if err != nil {
		return "", err
	}

	//datasetPathHistory, err := d.GetHistoryDataset(symbol)
	//if err == nil {
	//	rows := d.ReadCSV(datasetPathHistory, false)
	//	for _, row := range rows {
	//		_ = csvWriter.Write(row)
	//		csvWriter.Flush()
	//	}
	//}

	d.WriteToCsv(symbol, dateString, csvWriter, unzippedTrades, unzippedKLines)

	csvWriter.Flush()

	_ = csvFile.Close()

	return datasetPath, nil
}

func (d *DataSetBuilder) GetDependentOn(symbol string) string {
	asset := strings.ReplaceAll(symbol, "USDT", "")
	dependOn := d.GetCryptoQuote(symbol)

	return fmt.Sprintf("%s%s", asset, dependOn)
}

func (d *DataSetBuilder) GetCryptoQuote(symbol string) string {
	asset := strings.ReplaceAll(symbol, "USDT", "")
	dependOn := ""

	if slices.Contains(d.BtcDependent, asset) {
		dependOn = "BTC"
	}

	if slices.Contains(d.EthDependent, asset) {
		dependOn = "ETH"
	}

	if dependOn == "" {
		log.Panic(fmt.Sprintf("unsupported symbol %s", symbol))
	}

	return dependOn
}
