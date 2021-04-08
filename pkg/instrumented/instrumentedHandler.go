package instrumented

import (
	"errors"
	"net/http"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
	"github.com/gesundheitscloud/go-svc/pkg/prom"
)

var (
	//LatencyBuckets define buckets for histogram of HTTP request/reply latency metric - in seconds
	LatencyBuckets = []float64{.0001, .001, .01, .1, .25, .5, 1, 2.5, 5, 10}
	//SizeBuckets define buckets for histogram of HTTP request/reply size metric - in bytes
	SizeBuckets = []float64{16, 32, 64, 128, 256, 512, 1024, 5120, 20480, 102400, 512000, 1000000, 10000000}
	//DefaultInstrumentOptions hold options (API-path-specific) for HTTP instrumenter - record request size and response size
	DefaultInstrumentOptions = []prom.Option{prom.WithReqSize, prom.WithRespSize}
	//DefaultInstrumentInitOptions hold initialization options (API-handler-specific) for HTTP instrumenter - definitions of histogram buckets
	DefaultInstrumentInitOptions = []prom.InitOption{
		prom.WithLatencyBuckets(LatencyBuckets),
		prom.WithSizeBuckets(SizeBuckets),
	}
)

// HandlerFactory stores default settings and produces  handlers
type HandlerFactory struct {
	subsystemName      string
	defaultInitOptions []prom.InitOption
	defaultOptions     []prom.Option
}

// NewHandlerFactory returns new instrumentation object for HTTP handler
func NewHandlerFactory(subsystemName string, defaultInitOpts []prom.InitOption, defaultOpts []prom.Option) *HandlerFactory {
	return &HandlerFactory{
		subsystemName:      subsystemName,
		defaultInitOptions: defaultInitOpts,
		defaultOptions:     defaultOpts,
	}
}

// NewHandler produces an Handler with default values set
func (ihf *HandlerFactory) NewHandler(handlerName string, extraOpts ...interface{}) *Handler {
	initOptions := ihf.defaultInitOptions
	options := ihf.defaultOptions

	for _, o := range extraOpts {
		switch opt := o.(type) {
		case prom.InitOption:
			initOptions = append(initOptions, opt)
		case prom.Option:
			options = append(options, opt)
		default:
			logging.LogWarningf(errors.New("unknown instrumentation option type"), "change option to a known option")
		}

	}
	return newHandler(ihf.subsystemName, handlerName, initOptions, options)
}

// Handler holds instrumentation object for HTTP handlers
type Handler struct {
	instrumenter *prom.HandlerInstrumenter
	handlerName  string
	options      []prom.Option
}

// newHandler returns new instrumentation object for HTTP handler
func newHandler(subsystemName, name string, defaultInitOpts []prom.InitOption, defaultOpts []prom.Option) *Handler {
	nameInitOption := prom.WithSubsystem(subsystemName)
	initOptions := []prom.InitOption{nameInitOption}
	initOptions = append(initOptions, defaultInitOpts...)
	ih := &Handler{
		handlerName:  name,
		instrumenter: prom.NewHandlerInstrumenter(initOptions...),
		options:      []prom.Option{},
	}
	return ih.WithOptions(defaultOpts...)
}

// WithOptions returns Handler with Options set
func (ih *Handler) WithOptions(options ...prom.Option) *Handler {
	ih.options = options
	return ih
}

// Instrumenter returns HandlerInstrumenter
func (ih *Handler) Instrumenter() *prom.HandlerInstrumenter {
	return ih.instrumenter
}

// InstrumentChi applies the monitoring to the handler and return it together with path, so that it is easy to consume by chi router
// If `options` is set, default options are removed and replaced with provided parameters
func (ih *Handler) InstrumentChi(path string, fn func(w http.ResponseWriter, r *http.Request), options ...prom.Option) (string, func(w http.ResponseWriter, r *http.Request)) {
	if len(options) > 0 { //rewrite options if requested
		ih.WithOptions(options...)
	}
	return path, ih.Instrumenter().Instrument(ih.handlerName+path, http.HandlerFunc(fn), ih.options...).ServeHTTP
}
