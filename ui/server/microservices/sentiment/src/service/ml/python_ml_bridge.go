package ml

// #cgo pkg-config: python-3.11.6
// #include <Python.h>
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sentiment/src/model"
	"sentiment/src/service"
	"strings"
	"sync"
	"unsafe"
)

type PythonMLBridge struct {
	Mutex        *sync.RWMutex
	CoinDetector *service.CoinDetector
}

func (p *PythonMLBridge) getResultFilePath() string {
	return fmt.Sprintf("/go/src/app/results/result.txt")
}

func (p *PythonMLBridge) Initialize() {
	C.Py_Initialize()

	fmt.Println("-----------------------------------")
	fmt.Println("Python version via py C api:")
	fmt.Println(C.GoString(C.Py_GetVersion()))
	fmt.Println("-----------------------------------")

	pyCode := `
from transformers import pipeline
pipe = pipeline(
    "sentiment-analysis",
    model="/go/src/app/model",
)
`
	pyCodeC := C.CString(pyCode)
	defer C.free(unsafe.Pointer(pyCodeC))
	C.PyRun_SimpleString(pyCodeC)
	p.Mutex = &sync.RWMutex{}
}

func (p *PythonMLBridge) Finalize() {
	C.Py_Finalize()
}

func (p *PythonMLBridge) GetPythonPredictCode(sentiment string) string {
	resultPath := p.getResultFilePath()

	return fmt.Sprintf(string([]byte(`
result = pipe('%s')
print(result)

result_path = '%s'
with open(result_path, 'w') as out:
    print(result, file=out)
`)),
		sentiment,
		resultPath,
	)
}

func (p *PythonMLBridge) Predict(sentiment string) (*model.Prediction, error) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	resultPath := p.getResultFilePath()
	pyCode := p.GetPythonPredictCode(sentiment)

	pyCodeC := C.CString(pyCode)
	defer C.free(unsafe.Pointer(pyCodeC))
	C.PyRun_SimpleString(pyCodeC)

	fileContent, err := os.ReadFile(resultPath)
	if err != nil {
		return nil, err
	}

	var result []model.Prediction
	log.Println(strings.TrimSpace(string(fileContent)))
	jsonString := strings.ReplaceAll(strings.TrimSpace(string(fileContent)), "'", "\"")
	err = json.Unmarshal([]byte(jsonString), &result)

	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, errors.New("Result is empty.")
	}

	prediction := result[0]
	prediction.Coins = p.CoinDetector.DetectCoins(sentiment)

	return &prediction, nil
}
