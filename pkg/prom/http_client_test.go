package prom

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestClientInstrumentation(t *testing.T) {
	metricSrv := httptest.NewServer(promhttp.Handler())
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		}))
	instrumenter := NewRoundTripperInstrumenter()
	var testInstrumenter *RoundTripperInstrumenter

	t.Run("NewRoundTripperInstrumenter() idempotent", func(t *testing.T) {
		instrumenter = NewRoundTripperInstrumenter()
	})
	t.Run("NewRoundTripperInstrumenter() idempotent with different buckets", func(t *testing.T) {
		instrumenter = NewRoundTripperInstrumenter(WithLatencyBuckets([]float64{0.1, 0.2}))
	})
	t.Run("NewRoundTripperInstrumenter() idempotent with different subsystem", func(t *testing.T) {
		testInstrumenter = NewRoundTripperInstrumenter(WithSubsystem("test"), WithLatencyBuckets([]float64{0.1, 0.2}))
	})

	t.Run("Basic instrumentation of client", func(t *testing.T) {
		cli := &http.Client{
			Transport: instrumenter.Instrument("test-1", http.DefaultTransport),
		}

		res, err := cli.Get(srv.URL)
		if err != nil {
			t.Fatalf("http client call failed: %v", err)
		}
		res.Body.Close()

		checkMetricsCollection(metricSrv, []string{
			`d4l_phdp_http_out_requests_total{code="200",handler="test-1",method="get"} 1`,
			`d4l_phdp_http_out_request_duration_seconds_count{code="200",handler="test-1",method="get"} 1`,
			`d4l_phdp_http_out_request_duration_seconds_bucket{code="200",handler="test-1",method="get",le="0.25"} 1`,
			`d4l_phdp_http_out_requests{handler="test-1",method="get"} 0`,
		}, t)
	})

	t.Run("Instrumentation on different subsystem with different buckets", func(t *testing.T) {
		cli := &http.Client{
			Transport: testInstrumenter.Instrument("test-2", http.DefaultTransport),
		}

		res, err := cli.Get(srv.URL)
		if err != nil {
			t.Fatalf("http client call failed: %v", err)
		}
		res.Body.Close()

		checkMetricsCollection(metricSrv, []string{
			`d4l_test_http_out_requests_total{code="200",handler="test-2",method="get"} 1`,
			`d4l_test_http_out_request_duration_seconds_count{code="200",handler="test-2",method="get"} 1`,
			`d4l_test_http_out_request_duration_seconds_bucket{code="200",handler="test-2",method="get",le="0.1"} 1`,
			`d4l_test_http_out_requests{handler="test-2",method="get"} 0`,
		}, t)
	})
}
