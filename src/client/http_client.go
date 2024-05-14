package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type HttpClient struct {
}

func (h *HttpClient) Post(url string, message []byte, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(message))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 400 {
		return nil, errors.New(fmt.Sprintf("Request [%s] failed with error code: %d", url, res.StatusCode))
	}

	responseBody, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

func (h *HttpClient) Get(url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 400 {
		return nil, errors.New(fmt.Sprintf("Request [%s] failed with error code: %d", url, res.StatusCode))
	}

	responseBody, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return nil, err
	}

	return responseBody, nil
}
