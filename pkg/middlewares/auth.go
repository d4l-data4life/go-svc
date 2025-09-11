package middlewares

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"github.com/d4l-data4life/go-svc/pkg/instrumented"
	"github.com/d4l-data4life/go-svc/pkg/logging"
	"github.com/d4l-data4life/go-svc/pkg/prom"
)

const (
	// AuthHeaderName is the name of the authheader
	AuthHeaderName string = "Authorization"
)

// Auth is the handler responsible for auth
type Auth struct {
	*instrumented.Handler
	serviceSecret            string
	instrumentLatencyBuckets []float64
	instrumentSizeBuckets    []float64
}

// NewAuthentication initializes the auth middleware using JWT pub keys from ViperConfig
func NewAuthentication(serviceSecret string, handlerFactory *instrumented.HandlerFactory, opts ...AuthOption) *Auth {
	auth := &Auth{
		serviceSecret:            serviceSecret,
		instrumentLatencyBuckets: instrumented.LatencyBuckets,
		instrumentSizeBuckets:    instrumented.SizeBuckets,
	}

	for _, opt := range opts {
		opt(auth)
	}

	auth.Handler = handlerFactory.NewHandler("auth",
		prom.WithLatencyBuckets(auth.instrumentLatencyBuckets),
		prom.WithSizeBuckets(auth.instrumentSizeBuckets),
	)
	return auth
}

// AuthOption is to be implemented by functional options
type AuthOption func(*Auth)

// AuthWithLatencyBuckets changes the default latency buckets
func AuthWithLatencyBuckets(latencyBuckets []float64) AuthOption {
	return func(a *Auth) {
		a.instrumentLatencyBuckets = latencyBuckets
	}
}

// AuthWithSizeBuckets changes the default size buckets
func AuthWithSizeBuckets(sizeBuckets []float64) AuthOption {
	return func(a *Auth) {
		a.instrumentSizeBuckets = sizeBuckets
	}
}

// ServiceSecret is a middleware protecting routes with a service secret
func (auth *Auth) ServiceSecret(next http.Handler) http.Handler {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authToken, err := auth.getAuthSecret(r)
		if err != nil {
			WriteHTTPErrorCode(w, err, http.StatusUnauthorized)
			return
		}
		if subtle.ConstantTimeCompare([]byte(authToken), []byte(auth.serviceSecret)) == 0 {
			// If it is service based authentication authToken should be the appSecret
			http.Error(w, "invalid Authorization header", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
	return auth.Instrumenter().Instrument("auth", handlerFunc)
}

// getAuthSecret returns the contents of the authorization header
func (auth *Auth) getAuthSecret(r *http.Request) (string, error) {
	authHeaderContent := r.Header.Get(AuthHeaderName)
	if strings.HasPrefix(authHeaderContent, "Bearer ") {
		deprecatedErr := errors.New("deprecated 'bearer' key in the authentication")
		logging.LogWarningfCtx(r.Context(), deprecatedErr, "remove the 'bearer' key from the header for service-secert authentication")
		authHeaderContent = strings.TrimPrefix(authHeaderContent, "Bearer ")
	}

	if authHeaderContent == "" {
		err := errors.New("missing authentication header")
		logging.LogErrorfCtx(r.Context(), err, "error in secret-based authorization")
		return "", err
	}
	return authHeaderContent, nil
}
