package prom

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestServerInstrumentation(t *testing.T) {
	metricSrv := httptest.NewServer(promhttp.Handler())
	instrumenter := NewHandlerInstrumenter()
	var testInstrumenter *HandlerInstrumenter

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})

	t.Run("NewHandlerInstrumenter() idempotent", func(t *testing.T) {
		instrumenter = NewHandlerInstrumenter()
	})
	t.Run("NewHandlerInstrumenter() idempotent with different buckets", func(t *testing.T) {
		instrumenter = NewHandlerInstrumenter(WithLatencyBuckets([]float64{0.1, 0.2}), WithSizeBuckets([]float64{1000, 2000, 3000}))
	})
	t.Run("NewHandlerInstrumenter() idempotent with different subsystem", func(t *testing.T) {
		testInstrumenter = NewHandlerInstrumenter(
			WithSubsystem("test"), WithLatencyBuckets([]float64{0.1, 0.2}), WithSizeBuckets([]float64{1000, 2000, 3000}))
	})

	t.Run("Basic instrumentation of handler", func(t *testing.T) {
		srv := httptest.NewServer(instrumenter.Instrument("test-1", handler))
		client := srv.Client()

		res, err := client.Get(srv.URL)
		if err != nil {
			t.Fatalf("http client call failed: %v", err)
		}
		res.Body.Close()

		checkMetricsCollection(metricSrv, []string{
			`d4l_phdp_http_requests_total{code="200",handler="test-1",method="get"} 1`,
			`d4l_phdp_http_request_duration_seconds_count{code="200",handler="test-1",method="get"} 1`,
			`d4l_phdp_http_request_duration_seconds_bucket{code="200",handler="test-1",method="get",le="0.25"} 1`,
			`d4l_phdp_http_requests{handler="test-1",method="get"} 0`,
		}, t)

		srv.Close()
	})

	t.Run("Extended instrumentation of handler", func(t *testing.T) {
		srv := httptest.NewServer(instrumenter.Instrument("test-2", handler, WithReqSize, WithRespSize))
		client := srv.Client()

		res, err := client.Get(srv.URL)
		if err != nil {
			t.Fatalf("http client call failed: %v", err)
		}
		res.Body.Close()

		checkMetricsCollection(metricSrv, []string{
			`d4l_phdp_http_requests_total{code="200",handler="test-2",method="get"} 1`,
			`d4l_phdp_http_request_duration_seconds_count{code="200",handler="test-2",method="get"} 1`,
			`d4l_phdp_http_request_duration_seconds_bucket{code="200",handler="test-2",method="get",le="0.25"} 1`,
			`d4l_phdp_http_requests{handler="test-2",method="get"} 0`,
			`d4l_phdp_http_request_size_bytes_bucket{code="200",handler="test-2",method="get",le="1024"} 1`,
			`d4l_phdp_http_response_size_bytes_bucket{code="200",handler="test-2",method="get",le="1024"} 1`,
		}, t)

		srv.Close()
	})

	t.Run("Instrumentation on different subsystem with different buckets", func(t *testing.T) {
		srv := httptest.NewServer(testInstrumenter.Instrument("test-3", handler, WithReqSize))
		client := srv.Client()

		res, err := client.Get(srv.URL)
		if err != nil {
			t.Fatalf("http client call failed: %v", err)
		}
		res.Body.Close()

		checkMetricsCollection(metricSrv, []string{
			`d4l_test_http_requests_total{code="200",handler="test-3",method="get"} 1`,
			`d4l_test_http_request_duration_seconds_count{code="200",handler="test-3",method="get"} 1`,
			`d4l_test_http_requests{handler="test-3",method="get"} 0`,
			`d4l_test_http_request_duration_seconds_bucket{code="200",handler="test-3",method="get",le="0.1"} 1`,
			`d4l_test_http_request_size_bytes_bucket{code="200",handler="test-3",method="get",le="1000"} 1`,
		}, t)

		srv.Close()
	})

	metricSrv.Close()
}

func checkMetricsCollection(metricSrv *httptest.Server, want []string, t *testing.T) {
	resp, err := metricSrv.Client().Get(metricSrv.URL)
	if err != nil {
		t.Fatalf("error getting metrics: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading body from metrics response: %v", err)
	}
	bodyStr := string(body)

	for _, contain := range want {
		if !strings.Contains(bodyStr, contain) {
			t.Fatalf("%s is not collected: %s", contain, bodyStr)
		}
	}
}
