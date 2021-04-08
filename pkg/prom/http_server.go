package prom

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RegisterHTTPCountMetric registers and returns the default counter for HTTP requests
// to be used for prometheus monitoring
func RegisterHTTPCountMetric() *prometheus.CounterVec {
	return registerHTTPCountMetric(defaultSubsystem)
}

// RegisterHTTPDurationMetric registers and returns the default histogram for HTTP request durations
// to be used for prometheus monitoring
func RegisterHTTPDurationMetric() *prometheus.HistogramVec {
	return registerHTTPDurationMetric(defaultSubsystem, defaultLatencyBuckets())
}

func registerHTTPCountMetric(subsystem string) *prometheus.CounterVec {
	return registerCountMetric(subsystem, "http_requests_total",
		"How many HTTP requests processed, partitioned by path, status code and HTTP method.")
}

func registerHTTPDurationMetric(subsystem string, latencyBuckets []float64) *prometheus.HistogramVec {
	return registerHistogramMetric(subsystem, "http_request_duration_seconds",
		"A histogram of latencies for requests.",
		latencyBuckets)
}

func registerHTTPGaugeMetric(subsystem string) *prometheus.GaugeVec {
	return registerGaugeMetric(subsystem, "http_requests",
		"Amount of concurrently processed HTTP requests")
}

func registerHTTPRequestSizeMetric(subsystem string, sizeBuckets []float64) *prometheus.HistogramVec {
	return registerHistogramMetric(subsystem, "http_request_size_bytes",
		"A histogram of request sizes for HTTP requests.",
		sizeBuckets)
}

func registerHTTPResponseSizeMetric(subsystem string, sizeBuckets []float64) *prometheus.HistogramVec {
	return registerHistogramMetric(subsystem, "http_response_size_bytes",
		"A histogram of response sizes for HTTP requests.",
		sizeBuckets)
}

// HandlerInstrumenter that keeps pointers to the initialized metrics
type HandlerInstrumenter struct {
	gauge    *prometheus.GaugeVec
	count    *prometheus.CounterVec
	duration *prometheus.HistogramVec
	reqSize  *prometheus.HistogramVec
	respSize *prometheus.HistogramVec
}

// Instrument wraps the given handler and records gauge, count and duration of all HTTP requests
func (t *HandlerInstrumenter) Instrument(handlerName string, next http.Handler, options ...Option) http.Handler {
	o := &Options{}
	for _, option := range options {
		option(o)
	}
	if o.respSize {
		next = promhttp.InstrumentHandlerResponseSize(t.respSize.MustCurryWith(prometheus.Labels{"handler": handlerName}), next)
	}
	if o.reqSize {
		next = promhttp.InstrumentHandlerRequestSize(t.reqSize.MustCurryWith(prometheus.Labels{"handler": handlerName}), next)
	}

	return instrumentHandlerInFlight(t.gauge.MustCurryWith(prometheus.Labels{"handler": handlerName}),
		promhttp.InstrumentHandlerCounter(t.count.MustCurryWith(prometheus.Labels{"handler": handlerName}),
			promhttp.InstrumentHandlerDuration(t.duration.MustCurryWith(prometheus.Labels{"handler": handlerName}),
				next)))
}

type Middleware struct {
	*HandlerInstrumenter
	handlerName string
}

// Instrument is the middleware instrumentation function for the handlerName given when creating the Middleware
func (m *Middleware) Instrument(options ...Option) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return m.HandlerInstrumenter.Instrument(m.handlerName, next, options...)
	}
}

func (t *HandlerInstrumenter) Middleware(handlerName string) *Middleware {
	return &Middleware{t, handlerName}
}

// NewHandlerInstrumenter returns a new Instrumenter with PHDP's default metrics for incoming HTTP requests
func NewHandlerInstrumenter(options ...InitOption) *HandlerInstrumenter {
	o := &InitOptions{
		subsystem:      defaultSubsystem,
		sizeBuckets:    defaultSizeBuckets(),
		latencyBuckets: defaultLatencyBuckets(),
	}
	for _, option := range options {
		option(o)
	}

	return &HandlerInstrumenter{
		gauge:    registerHTTPGaugeMetric(o.subsystem),
		count:    registerHTTPCountMetric(o.subsystem),
		duration: registerHTTPDurationMetric(o.subsystem, o.latencyBuckets),
		reqSize:  registerHTTPRequestSizeMetric(o.subsystem, o.sizeBuckets),
		respSize: registerHTTPResponseSizeMetric(o.subsystem, o.sizeBuckets),
	}
}

// InstrumentHandlerInFlight instruments the given handler with a Prometheus GaugeVec
// in order to count the concurrent requests
// GaugeVec must have at least one label (the HTTP method)
func instrumentHandlerInFlight(g *prometheus.GaugeVec, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sanitizedM := strings.ToLower(r.Method)
		g.WithLabelValues(sanitizedM).Inc()
		defer g.WithLabelValues(sanitizedM).Dec()
		next.ServeHTTP(w, r)
	})
}
