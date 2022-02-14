package middlewares

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Errors
var (
	ErrNoSecretInRequest   = errors.New("no service secret present in request")
	ErrInvalidSecret       = errors.New("provided service secret invalid")
	ErrMalformedAuthHeader = errors.New("malformed authorization header content")
)

type logger interface {
	ErrGeneric(context.Context, error) error
}

// ServiceSecretAuthenticator is the type that provides a service secret authentication middleware
type ServiceSecretAuthenticator struct {
	Secret string
	logger logger
}

func NewServiceSecretAuthenticator(serviceSecret string, l logger) *ServiceSecretAuthenticator {
	return &ServiceSecretAuthenticator{logger: l, Secret: serviceSecret}
}

// Authenticate is the decorator that protects a handler using a service secret.
// Authenticate parses the Authorization header in the request and validates it
// against the configured service secret.
func (sa ServiceSecretAuthenticator) Authenticate() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			authHeaderContent := req.Header.Get("Authorization")
			if authHeaderContent == "" {
				_ = sa.logger.ErrGeneric(req.Context(), fmt.Errorf("extracting auth secret: %w", ErrNoSecretInRequest))
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			headerContentSplit := strings.Split(authHeaderContent, "Bearer ")
			if len(headerContentSplit) != 2 {
				_ = sa.logger.ErrGeneric(req.Context(), fmt.Errorf("parsing auth secret: %w", ErrMalformedAuthHeader))
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			headerSecret := headerContentSplit[1]
			if headerSecret == "" {
				_ = sa.logger.ErrGeneric(req.Context(), fmt.Errorf("parsing auth secret: %w", ErrMalformedAuthHeader))
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if headerSecret != sa.Secret {
				_ = sa.logger.ErrGeneric(req.Context(), fmt.Errorf("%w", ErrInvalidSecret))
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			h.ServeHTTP(w, req)
		})
	}
}
