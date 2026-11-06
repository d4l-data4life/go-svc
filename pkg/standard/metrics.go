package standard

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/d4l-data4life/go-svc/pkg/logging"
)

// ListenAndServeMetrics serves the Prometheus metrics at /metrics on port until runCtx is canceled.
// The returned channel closes when the server has stopped.
//
// Use a port that no Kubernetes service exposes, and let Prometheus scrape the pod through the
// prometheus.io/port annotation. Metrics on the application port are published with it by the ingress.
func ListenAndServeMetrics(runCtx context.Context, port string) <-chan struct{} {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())

	listenAddress := net.JoinHostPort("", port)
	logging.LogInfof("metrics listening on %s", listenAddress)
	return serve(runCtx, &http.Server{Addr: listenAddress, Handler: mux, ReadHeaderTimeout: 3 * time.Second})
}
