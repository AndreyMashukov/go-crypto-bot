package metrics

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// DefaultAddr is the admin port the writer exposes /metrics on. Lives
// off the trading API so dashboards / scrapers reach the bot's
// observability surface without going through the user-facing routes.
const DefaultAddr = ":9101"

// Serve runs a tiny HTTP server with the Prometheus default-registry
// handler on /metrics and a /healthz probe. Blocks until ctx cancels
// or the underlying server errors. A nil ctx falls back to
// context.Background.
func Serve(ctx context.Context, addr string) error {
	if addr == "" {
		addr = DefaultAddr
	}
	if ctx == nil {
		ctx = context.Background()
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			errCh <- nil
			return
		}
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
