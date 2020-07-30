package instrumented

import (
	"net/http"

	"github.com/gesundheitscloud/go-monitoring/prom"
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
func (ihf *HandlerFactory) NewHandler(handlerName string, extraOpts ...prom.Option) *Handler {
	return newHandler(ihf.subsystemName, handlerName, ihf.defaultInitOptions, ihf.defaultOptions)
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
