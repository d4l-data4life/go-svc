package prom

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func registerHTTPOutCountMetric(subsystem string) *prometheus.CounterVec {
	return registerCountMetric(subsystem, "http_out_requests_total",
		"The amount of outgoing HTTP requests, partitioned by status code, method and target")
}

func registerHTTPOutDurationMetric(subsystem string, latencyBuckets []float64) *prometheus.HistogramVec {
	return registerHistogramMetric(subsystem, "http_out_request_duration_seconds",
		"A histogram of latencies for outgoing requests.",
		latencyBuckets)
}

func registerHTTPOutGaugeMetric(subsystem string) *prometheus.GaugeVec {
	return registerGaugeMetric(subsystem, "http_out_requests",
		"Amount of concurrently processed outgoing HTTP requests")
}

// RoundTripperInstrumenter that keeps pointers to the before registered metrics
type RoundTripperInstrumenter struct {
	gauge    *prometheus.GaugeVec
	count    *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

// Instrument wraps the given handler and records gauge, count and duration of all HTTP requests
func (t *RoundTripperInstrumenter) Instrument(handlerName string, next http.RoundTripper) http.RoundTripper {
	return instrumentRoundTripperInFlight(t.gauge.MustCurryWith(prometheus.Labels{"handler": handlerName}),
		promhttp.InstrumentRoundTripperCounter(t.count.MustCurryWith(prometheus.Labels{"handler": handlerName}),
			promhttp.InstrumentRoundTripperDuration(t.duration.MustCurryWith(prometheus.Labels{"handler": handlerName}),
				next)))
}

// NewRoundTripperInstrumenter returns a new Instrumenter with PHDP's default metrics for incoming HTTP requests
func NewRoundTripperInstrumenter(options ...InitOption) *RoundTripperInstrumenter {
	o := &InitOptions{
		subsystem:      defaultSubsystem,
		sizeBuckets:    defaultSizeBuckets(),
		latencyBuckets: defaultLatencyBuckets(),
	}
	for _, option := range options {
		option(o)
	}

	return &RoundTripperInstrumenter{
		gauge:    registerHTTPOutGaugeMetric(o.subsystem),
		count:    registerHTTPOutCountMetric(o.subsystem),
		duration: registerHTTPOutDurationMetric(o.subsystem, o.latencyBuckets),
	}
}

func instrumentRoundTripperInFlight(g *prometheus.GaugeVec, next http.RoundTripper) promhttp.RoundTripperFunc {
	return promhttp.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		sanitizedM := strings.ToLower(r.Method)
		g.WithLabelValues(sanitizedM).Inc()
		defer g.WithLabelValues(sanitizedM).Dec()
		return next.RoundTrip(r)
	})
}
