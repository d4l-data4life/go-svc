package middlewares

import (
	"context"
	"crypto/rsa"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gesundheitscloud/go-monitoring/prom"
	"github.com/gesundheitscloud/go-svc/pkg/d4lcontext"
	"github.com/gesundheitscloud/go-svc/pkg/instrumented"
	"github.com/gesundheitscloud/go-svc/pkg/logging"
	uuid "github.com/gofrs/uuid"
)

const (
	// AuthHeaderName is the name of the authheader
	AuthHeaderName string = "Authorization"
)

type claims struct {
	jwt.StandardClaims
	// ClientID is the Client ID claim (gesundheitscloud private claim)
	ClientID string `json:"ghc:cid"`
	// UserID is the claim that encodes the user who originally requested the JWT (gesundheitscloud private claim)
	UserID uuid.UUID `json:"ghc:uid"`
	// TenantID is the claim that encodes the tenant (e.g. d4l, charite)
	TenantID string `json:"ghc:tid"`
}

// Auth is the handler responsible for auth
type Auth struct {
	*instrumented.Handler
	serviceSecret            string
	publicKey                *rsa.PublicKey
	instrumentLatencyBuckets []float64
	instrumentSizeBuckets    []float64
}

// NewAuth initializes the auth middleware
func NewAuth(serviceSecret string, publicKey *rsa.PublicKey, handlerFactory *instrumented.HandlerFactory, opts ...AuthOption) *Auth {
	auth := &Auth{
		serviceSecret:            serviceSecret,
		publicKey:                publicKey,
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

// JWT is a middleware protecting routes with a jwt based auth
func (auth *Auth) JWT(next http.Handler) http.Handler {
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authToken, err := auth.getBearerToken(r)
		if err != nil {
			WriteHTTPErrorCode(w, err, http.StatusUnauthorized)
			return
		}
		tk := &claims{}
		_, err = jwt.ParseWithClaims(authToken, tk, func(token *jwt.Token) (interface{}, error) {
			return auth.publicKey, nil
		})

		if err != nil {
			if ve, ok := err.(*jwt.ValidationError); ok {
				switch {
				case ve.Errors&jwt.ValidationErrorMalformed != 0:
					logging.LogErrorfCtx(r.Context(), err, "malformed jwt")
					WriteHTTPErrorCode(w, errors.New("token is malformed"), http.StatusUnauthorized)
				case ve.Errors&jwt.ValidationErrorExpired != 0:
					WriteHTTPErrorCode(w, errors.New("token is expired"), http.StatusUnauthorized)
				case ve.Errors&jwt.ValidationErrorNotValidYet != 0:
					WriteHTTPErrorCode(w, errors.New("token is not valid yet"), http.StatusUnauthorized)
				default:
					logging.LogErrorfCtx(r.Context(), err, "Error parsing jwt")
					WriteHTTPErrorCode(w, errors.New("error parsing jwt"), http.StatusUnauthorized)
				}
				return
			}
		}

		ctx := context.WithValue(r.Context(), d4lcontext.ClientIDContextKey, tk.ClientID)
		ctx = context.WithValue(ctx, d4lcontext.UserIDContextKey, tk.UserID)
		ctx = context.WithValue(ctx, d4lcontext.TenantIDContextKey, tk.TenantID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
	return auth.Instrumenter().Instrument("auth", handlerFunc)
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

// getBearerToken returns the Bearer AuthToken from the given request
func (auth *Auth) getBearerToken(r *http.Request) (string, error) {
	authHeaderContent := r.Header.Get(AuthHeaderName)
	if authHeaderContent == "" {
		return "", errors.New("missing authentication header")
	}

	authTokenSplit := strings.Split(authHeaderContent, "Bearer ")
	if len(authTokenSplit) != 2 {
		return "", errors.New("malformed authentication header")
	}

	authToken := authTokenSplit[1]
	if authToken == "" {
		return "", errors.New("malformed authentication header")
	}
	return authToken, nil
}
