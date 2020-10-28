package middlewares

import (
	"net/http"

	"github.com/gesundheitscloud/go-monitoring/prom"
	"github.com/gesundheitscloud/go-svc/pkg/d4lcontext"
	"github.com/gesundheitscloud/go-svc/pkg/instrumented"
	"golang.org/x/net/xsrftoken"
)

// XSRF is the handler responsible for xsrf validation
type XSRF struct {
	*instrumented.Handler
	xsrfSecret               string
	xsrfHeader               string
	instrumentLatencyBuckets []float64
	instrumentSizeBuckets    []float64
}

// NewXSRF initializes a new handler
func NewXSRF(xsrfSecret string, xsrfHeader string, handlerFactory *instrumented.HandlerFactory, opts ...XSRFOption) *XSRF {
	xsrf := &XSRF{
		xsrfSecret:               xsrfSecret,
		xsrfHeader:               xsrfHeader,
		instrumentLatencyBuckets: instrumented.LatencyBuckets,
		instrumentSizeBuckets:    instrumented.SizeBuckets,
	}
	for _, opt := range opts {
		opt(xsrf)
	}

	xsrf.Handler = handlerFactory.NewHandler("xsrf",
		prom.WithLatencyBuckets(xsrf.instrumentLatencyBuckets),
		prom.WithSizeBuckets(xsrf.instrumentSizeBuckets),
	)
	return xsrf

}

// XSRFOption is to be implemented by functional options
type XSRFOption func(*XSRF)

// XSRFWithLatencyBuckets changes the default latency buckets
func XSRFWithLatencyBuckets(latencyBuckets []float64) XSRFOption {
	return func(x *XSRF) {
		x.instrumentLatencyBuckets = latencyBuckets
	}
}

// XSRFWithSizeBuckets changes the default size buckets
func XSRFWithSizeBuckets(sizeBuckets []float64) XSRFOption {
	return func(x *XSRF) {
		x.instrumentSizeBuckets = sizeBuckets
	}
}

// XSRF returns the XSRFMiddleware
func (xsrf *XSRF) XSRF(next http.Handler) http.Handler {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ignore XSRF for GET, HEAD, OPTIONS methods
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Test for XSRF token in request header
		token := r.Header.Get(xsrf.xsrfHeader)
		if token == "" {
			http.Error(w, "missing XSRF token", http.StatusForbidden)
			return
		}

		// Get account id from the request
		accountID, err := d4lcontext.ParseRequesterID(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate XSRF token
		if !xsrftoken.Valid(token, xsrf.xsrfSecret, accountID.String(), "") {
			http.Error(w, "invalid XSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})

	return xsrf.Instrumenter().Instrument("auth", handlerFunc)
}
