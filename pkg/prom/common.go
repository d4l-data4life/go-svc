package prom

import "github.com/prometheus/client_golang/prometheus"

// Options for instrumenting the handler
type Options struct {
	reqSize  bool
	respSize bool
}

// Option to be implemented by functional options
type Option func(*Options)

// WithReqSize instruments the handler with a request size metric
var WithReqSize Option = func(o *Options) { //nolint-gochecknoglobals would require a lot of refactoring in other services
	o.reqSize = true
}

// WithRespSize instruments the handler with a response size metric
var WithRespSize Option = func(o *Options) { //nolint-gochecknoglobals would require a lot of refactoring in other services
	o.respSize = true
}

// InitOptions is used to initialize the instrumentation
type InitOptions struct {
	subsystem      string
	sizeBuckets    []float64
	latencyBuckets []float64
}

// InitOption to be implemented by functional options
type InitOption func(*InitOptions)

// WithSubsystem changes the default subsystem (i.e. "phdp")
func WithSubsystem(subsystem string) InitOption {
	return func(o *InitOptions) {
		o.subsystem = subsystem
	}
}

// WithSizeBuckets changes the default size buckets
func WithSizeBuckets(sizeBuckets []float64) InitOption {
	return func(o *InitOptions) {
		o.sizeBuckets = sizeBuckets
	}
}

// WithLatencyBuckets changes the default latency buckets
func WithLatencyBuckets(latencyBuckets []float64) InitOption {
	return func(o *InitOptions) {
		o.latencyBuckets = latencyBuckets
	}
}

const namespace = "d4l"
const defaultSubsystem = "phdp"

func defaultSizeBuckets() []float64 {
	return []float64{1024, 5120, 20480, 102400, 512000, 1048576, 10485760, 52428800} // 1KB, 5KB, 20KB, 100KB, 500KB, 1MB, 10MB, 50MB
}
func defaultLatencyBuckets() []float64 {
	return []float64{.25, .5, 1, 2.5, 5, 10} // seconds
}

func registerHistogramMetric(subsystem string, name string, help string, buckets []float64) *prometheus.HistogramVec {
	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      name,
			Help:      help,
			Buckets:   buckets,
		},
		[]string{"code", "method", "handler"},
	)
	if err := prometheus.Register(histogram); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			histogram = are.ExistingCollector.(*prometheus.HistogramVec)
		} else {
			panic(err)
		}
	}
	return histogram
}

func registerCountMetric(subsystem string, name string, help string) *prometheus.CounterVec {
	count := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      name,
			Help:      help,
		},
		[]string{"code", "method", "handler"},
	)
	if err := prometheus.Register(count); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			count = are.ExistingCollector.(*prometheus.CounterVec)
		} else {
			panic(err)
		}
	}
	return count
}

func registerGaugeMetric(subsystem string, name string, help string) *prometheus.GaugeVec {
	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      name,
			Help:      help,
		},
		[]string{"method", "handler"},
	)
	if err := prometheus.Register(gauge); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			gauge = are.ExistingCollector.(*prometheus.GaugeVec)
		} else {
			panic(err)
		}
	}
	return gauge
}
