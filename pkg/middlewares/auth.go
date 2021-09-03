package middlewares

import (
	"context"
	"crypto/rsa"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"github.com/gesundheitscloud/go-svc/pkg/d4lcontext"
	"github.com/gesundheitscloud/go-svc/pkg/dynamic"
	"github.com/gesundheitscloud/go-svc/pkg/instrumented"
	"github.com/gesundheitscloud/go-svc/pkg/logging"
	"github.com/gesundheitscloud/go-svc/pkg/prom"
	uuid "github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
)

const (
	// AuthHeaderName is the name of the authheader
	AuthHeaderName string = "Authorization"
)

var (
	ErrorAuthMisconfiguredJWTProvider  = errors.New("authenticator misconfigured: JWT public key provider")
	ErrorAuthMisconfiguredJWTPublicKey = errors.New("authenticator misconfigured: JWT public key - file missing")
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
	publicKeyProvider        JWTPublicKeysProvider
	instrumentLatencyBuckets []float64
	instrumentSizeBuckets    []float64
}

// NewAuth initializes the auth middleware
// DEPRECATED in favour of: NewAuthentication
func NewAuth(serviceSecret string, publicKey *rsa.PublicKey, handlerFactory *instrumented.HandlerFactory, opts ...AuthOption) *Auth {
	return NewAuthentication(serviceSecret, AuthWithRSAPublicKey(publicKey), handlerFactory, opts...)
}

type JWTPublicKeysProvider interface {
	JWTPublicKeys() ([]dynamic.JWTPublicKey, error)
}

// NewAuthentication initializes the auth middleware using JWT pub keys from ViperConfig
func NewAuthentication(serviceSecret string, keys AuthOptionJWTKeys, handlerFactory *instrumented.HandlerFactory, opts ...AuthOption) *Auth {
	auth := &Auth{
		serviceSecret:            serviceSecret,
		publicKey:                nil,
		publicKeyProvider:        nil,
		instrumentLatencyBuckets: instrumented.LatencyBuckets,
		instrumentSizeBuckets:    instrumented.SizeBuckets,
	}

	// setup keys first - this is obligatory, but passing keys=nil would cause panic
	if keys != nil {
		keys(auth)
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

// AuthOptionJWTKeys is to be implemented by functional options
type AuthOptionJWTKeys func(*Auth)

// AuthWithRSAPublicKey - DEPRECATED - sets the public key from file
// Use AuthWithPublicKeyProvider instead
func AuthWithRSAPublicKey(pk *rsa.PublicKey) AuthOptionJWTKeys {
	return func(a *Auth) {
		a.publicKey = pk
	}
}

// AuthWithPublicKeyProvider sets the JWT public key provider interface
func AuthWithPublicKeyProvider(prov JWTPublicKeysProvider) AuthOptionJWTKeys {
	return func(a *Auth) {
		a.publicKeyProvider = prov
	}
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
		authToken, terr := auth.getBearerToken(r)
		if terr != nil {
			WriteHTTPErrorCode(w, terr, http.StatusUnauthorized)
			return
		}
		var tk *claims
		var status int
		var err error
		switch {
		case auth.publicKeyProvider != nil:
			tk, status, err = auth.jwtProvider(authToken, w, r) // preferred, required for "easier secrets rotation"
		default:
			tk, status, err = auth.jwtFile(authToken, w, r) // deprecated, should serve only as fallback for the transition-period
		}
		if err != nil {
			WriteHTTPErrorCode(w, err, status)
			return
		}
		ctx := context.WithValue(r.Context(), d4lcontext.ClientIDContextKey, tk.ClientID)
		ctx = context.WithValue(ctx, d4lcontext.UserIDContextKey, tk.UserID)
		ctx = context.WithValue(ctx, d4lcontext.TenantIDContextKey, tk.TenantID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
	return auth.Instrumenter().Instrument("auth", handlerFunc)
}

// jwtProvider verifies the request using public keys provided by the key-provider - at least one public key must match
func (auth *Auth) jwtProvider(authToken string, w http.ResponseWriter, r *http.Request) (tk *claims, status int, err error) {
	tk = &claims{}
	if auth.publicKeyProvider == nil {
		logging.LogErrorfCtx(r.Context(), ErrorAuthMisconfiguredJWTProvider, "JWT public key provider is nil")
		return tk, http.StatusInternalServerError, ErrorAuthMisconfiguredJWTProvider
	}

	// we check multiple public keys
	pubKeys, err := auth.publicKeyProvider.JWTPublicKeys()
	if err != nil {
		logging.LogErrorfCtx(r.Context(), err, "unable to use public keys")
		return tk, http.StatusInternalServerError, err
	}
	var jwtParseErr error
	for _, key := range pubKeys {
		logging.LogDebugfCtx(r.Context(), "verifying using JWT public key '%s'", key.Name)
		_, jwtParseErr = jwt.ParseWithClaims(authToken, tk, func(token *jwt.Token) (interface{}, error) {
			return key.Key, nil
		})
		if jwtParseErr == nil {
			logging.LogDebugfCtx(r.Context(), "JWT public key '%s': match", key.Name)
			return tk, http.StatusContinue, nil
		} else {
			logging.LogDebugfCtx(r.Context(), "JWT public key '%s': no match - %s", key.Name, jwtParseErr.Error())
		}
	}
	// 0 public keys match - return last error
	return tk, http.StatusUnauthorized, jwtParseErr
}

// jwtFile verifies the request using single public key provided from file
func (auth *Auth) jwtFile(authToken string, w http.ResponseWriter, r *http.Request) (tk *claims, status int, err error) {
	tk = &claims{}
	if auth.publicKey == nil {
		logging.LogErrorfCtx(r.Context(), ErrorAuthMisconfiguredJWTPublicKey, "JWT rsa.PublicKey is nil")
		return tk, http.StatusInternalServerError, ErrorAuthMisconfiguredJWTPublicKey
	}
	_, err = jwt.ParseWithClaims(authToken, tk, func(token *jwt.Token) (interface{}, error) {
		return auth.publicKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			switch {
			case ve.Errors&jwt.ValidationErrorMalformed != 0:
				logging.LogErrorfCtx(r.Context(), err, "malformed jwt")
				return tk, http.StatusUnauthorized, errors.New("token is malformed")
			case ve.Errors&jwt.ValidationErrorExpired != 0:
				return tk, http.StatusUnauthorized, errors.New("token is expired")
			case ve.Errors&jwt.ValidationErrorNotValidYet != 0:
				return tk, http.StatusUnauthorized, errors.New("token is not valid yet")
			default:
				logging.LogErrorfCtx(r.Context(), err, "Error parsing jwt")
				return tk, http.StatusUnauthorized, errors.New("error parsing jwt")
			}
		}
	}
	return tk, http.StatusContinue, nil
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
