package metrics_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/AndreyMashukov/go-crypto-bot/server/shared/metrics"
)

// TestServeExposesMetricsAndHonoursContext spins the admin HTTP server
// on a high port, hits /metrics + /healthz, asserts both succeed, then
// cancels the context and confirms Serve returns promptly.
func TestServeExposesMetricsAndHonoursContext(t *testing.T) {
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()
	done := make(chan error, 1)
	go func() { done <- metrics.Serve(rootCtx, ":19101") }()

	httpClient := &http.Client{Timeout: 500 * time.Millisecond}

	healthCtx, healthCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer healthCancel()
	for {
		select {
		case <-healthCtx.Done():
			t.Fatalf("metrics server did not come up: %v", healthCtx.Err())
		default:
		}
		req, _ := http.NewRequestWithContext(healthCtx, http.MethodGet, "http://127.0.0.1:19101/healthz", http.NoBody)
		resp, err := httpClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	metrics.TickPublishFailure.WithLabelValues("BTCUSDT_serve_smoke").Inc()

	scrapeCtx, scrapeCancel := context.WithTimeout(context.Background(), time.Second)
	defer scrapeCancel()
	req, _ := http.NewRequestWithContext(scrapeCtx, http.MethodGet, "http://127.0.0.1:19101/metrics", http.NoBody)
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if !strings.Contains(string(body), "tick_publish_failure_total") {
		t.Errorf("/metrics output missing tick_publish_failure_total\nbody:\n%s", string(body))
	}

	rootCancel()
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("Serve returned unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not return after cancel within 2s")
	}
}
