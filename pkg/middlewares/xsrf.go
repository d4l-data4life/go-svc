package middlewares

import (
	"net/http"

	"github.com/gesundheitscloud/go-svc/pkg/instrumented"
)

// XSRF is the handler responsible for xsrf validation
type XSRF struct{}

// XSRFOption is to be implemented by functional options
type XSRFOption func(*XSRF)

// XSRFWithLatencyBuckets changes the default latency buckets
func XSRFWithLatencyBuckets(latencyBuckets []float64) XSRFOption {
	return func(x *XSRF) {}
}

// XSRFWithSizeBuckets changes the default size buckets
func XSRFWithSizeBuckets(sizeBuckets []float64) XSRFOption {
	return func(x *XSRF) {}
}

// NewXSRF initializes a new handler
// Deprecated: Superfluous, as we already attach authentication information to request headers (JWT token, service secrets).
// Extra XSRF protection would only be required if we used authentication information attached automatically, e.g. cookies.
// Please remove the middleware from your service, but keep the XSRF handler for backwards-compatibility.
func NewXSRF(xsrfSecret string, xsrfHeader string, handlerFactory *instrumented.HandlerFactory, opts ...interface{}) *XSRF {
	return &XSRF{}
}

// XSRF returns the XSRFMiddleware
func (xsrf *XSRF) XSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
