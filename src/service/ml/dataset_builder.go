package ml

import (
	"encoding/csv"
	"errors"
	"fmt"
	"gitlab.com/open-soft/go-crypto-bot/src/repository"
	"log"
	"os"
)

type DataSetBuilder struct {
	StatRepository *repository.StatRepository
}

func (d *DataSetBuilder) PrepareDataset(symbol string) (string, error) {
	datasetPath := fmt.Sprintf("/go/src/app/datasets/dataset_%s.csv", symbol)
	_ = os.Remove(datasetPath)
	csvFile, err := os.Create(datasetPath)

	if err != nil {
		return "", err
	}

	csvWriter := csv.NewWriter(csvFile)
	log.Printf("[%s] Fetching ML dataset...", symbol)
	dataset := d.StatRepository.GetMLDataset(symbol, d.GetSecondarySymbol(symbol))
	log.Printf("[%s] ML dataset length is %d", symbol, len(dataset))

	if len(dataset) < 50 {
		return "", errors.New("not enough dataset length")
	}

	for _, record := range dataset {
		row := []string{
			fmt.Sprintf("%.10f", record.OrderBookBuyFirstQty),
			fmt.Sprintf("%.10f", record.OrderBookSellFirstQty),
			fmt.Sprintf("%.10f", record.OrderBookBuyQtySum),
			fmt.Sprintf("%.10f", record.OrderBookSellQtySum),
			fmt.Sprintf("%.10f", record.OrderBookBuyVolumeSum),
			fmt.Sprintf("%.10f", record.OrderBookSellVolumeSum),
			fmt.Sprintf("%.10f", record.SecondaryPrice),
			fmt.Sprintf("%.10f", record.PrimaryPrice),
		}

		_ = csvWriter.Write(row)
		csvWriter.Flush()
	}

	csvWriter.Flush()
	_ = csvFile.Close()

	return datasetPath, nil
}

func (d *DataSetBuilder) GetSecondarySymbol(symbol string) string {
	secondary := "BTCUSDT"

	if symbol == "BTCUSDT" {
		secondary = "ETHUSDT"
	}

	return secondary
}
