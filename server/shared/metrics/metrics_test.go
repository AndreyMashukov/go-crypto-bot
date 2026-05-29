package metrics_test

import (
	"testing"

	dto "github.com/prometheus/client_model/go"

	"github.com/AndreyMashukov/go-crypto-bot/server/shared/metrics"
)

func counterValue(t *testing.T, c interface {
	Write(*dto.Metric) error
},
) float64 {
	t.Helper()
	m := &dto.Metric{}
	if err := c.Write(m); err != nil {
		t.Fatalf("Write metric: %v", err)
	}
	if m.Counter == nil {
		t.Fatal("metric has no counter payload")
	}
	return m.Counter.GetValue()
}

func TestTickPublishFailureLabelsAndIncrements(t *testing.T) {
	c := metrics.TickPublishFailure.WithLabelValues("BTCUSDT_test_publish")
	start := counterValue(t, c)
	c.Inc()
	c.Inc()
	got := counterValue(t, c)
	if got-start != 2 {
		t.Errorf("TickPublishFailure delta = %f, want 2", got-start)
	}
}

func TestTickDropLabelsAndIncrements(t *testing.T) {
	c := metrics.TickDrop.WithLabelValues("ETHUSDT_test_drop")
	start := counterValue(t, c)
	c.Inc()
	got := counterValue(t, c)
	if got-start != 1 {
		t.Errorf("TickDrop delta = %f, want 1", got-start)
	}
}

func TestPnLRealizedAccumulates(t *testing.T) {
	c := metrics.PnLRealized.WithLabelValues("XRPUSDT_test_pnl")
	start := counterValue(t, c)
	c.Add(12.5)
	c.Add(7.5)
	got := counterValue(t, c)
	if got-start != 20.0 {
		t.Errorf("PnLRealized delta = %f, want 20", got-start)
	}
}
