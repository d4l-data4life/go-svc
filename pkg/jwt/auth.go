package jwt

import (
	"crypto/rsa"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	jwtReq "github.com/dgrijalva/jwt-go/request"
	"github.com/go-chi/chi"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

const ErrMsgVerifier = "verification failed"

var (
	ErrNoClaims     = errors.New("missing claims")
	ErrMissingScope = errors.New("necessary scope not in jwt")
	ErrInvalidToken = errors.New("token is invalid")
)

// Authenticator contains the public key necessary to verify the signature.
type Authenticator struct {
	publicKey *rsa.PublicKey
	logger    logger
}

// New creates an Authenticator that creates an auth Middleware.
func New(pk *rsa.PublicKey, l logger) *Authenticator {
	return &Authenticator{pk, l}
}

type rule func(*http.Request) error

// Verify checks if JWT satisfies the given rules.
func (auth *Authenticator) Verify(rules ...rule) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawToken, err := jwtReq.OAuth2Extractor.ExtractToken(r)
			if err != nil {
				_ = auth.logger.ErrUserAuth(r.Context(), errors.Wrap(err, ErrMsgVerifier))
				httpClientError(w, http.StatusUnauthorized)
				return
			}

			parsedToken, err := jwt.ParseWithClaims(rawToken, &Claims{}, func(_ *jwt.Token) (interface{}, error) {
				return auth.publicKey, nil
			})
			if err != nil {
				_ = auth.logger.ErrUserAuth(r.Context(), errors.Wrap(err, ErrMsgVerifier))
				httpClientError(w, http.StatusUnauthorized)
				return
			}

			if !parsedToken.Valid {
				_ = auth.logger.ErrUserAuth(r.Context(), errors.Wrap(ErrInvalidToken, ErrMsgVerifier))
				httpClientError(w, http.StatusBadRequest)
				return
			}

			claims, ok := parsedToken.Claims.(*Claims)
			if !ok {
				_ = auth.logger.ErrUserAuth(r.Context(), errors.Wrap(ErrNoClaims, ErrMsgVerifier))
				httpClientError(w, http.StatusBadRequest)
				return
			}

			r = r.WithContext(NewContext(r.Context(), claims))

			for _, rule := range rules {
				if err := rule(r); err != nil {
					_ = auth.logger.ErrUserAuth(r.Context(), errors.Wrap(err, ErrMsgVerifier))
					httpClientError(w, http.StatusUnauthorized)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// WithOwner verifies that the given function returns the UUID of the JWT's subject ID.
func WithOwner(owner func(r *http.Request) uuid.UUID) rule {
	return func(r *http.Request) error {
		claims, ok := fromContext(r.Context())
		if !ok {
			return ErrNoClaims
		}

		haveID := owner(r)
		if haveID != claims.Subject.ID || haveID == uuid.Nil {
			return ErrSubjectNotOwner
		}

		return nil
	}
}

// WithGorillaOwner provides a Gorilla/Mux specific solution for parsing the owner from the path.
// This logic started to be replicated all around the services and was the initial reason for
// adding the middleware package.
func WithGorillaOwner(ownerKey string) rule {
	return func(r *http.Request) error {
		claims, ok := fromContext(r.Context())
		if !ok {
			return ErrNoClaims
		}

		vars := mux.Vars(r)
		value, ok := vars[ownerKey]
		if !ok {
			return ErrSubjectNotOwner
		}

		haveID, err := uuid.FromString(value)
		if err != nil || haveID != claims.Subject.ID || haveID == uuid.Nil {
			return ErrSubjectNotOwner
		}

		return nil
	}
}

// WithChiOwner provides a Chi/Mux specific solution for parsing the owner from the path.
// It returns a function that verifies that the owner from the path matches the subject ID from the JWT.
func WithChiOwner(ownerKey string) rule {
	return func(r *http.Request) error {
		claims, ok := fromContext(r.Context())
		if !ok {
			return ErrNoClaims
		}

		value := chi.URLParam(r, ownerKey)
		if value == "" {
			return ErrSubjectNotOwner
		}

		haveID, err := uuid.FromString(value)
		if err != nil || haveID != claims.Subject.ID || haveID == uuid.Nil {
			return ErrSubjectNotOwner
		}

		return nil
	}
}

// WithAnyScope verifies that at lest one of the given scopes is in the JWT.
func WithAnyScope(scopes ...string) rule {
	return func(r *http.Request) error {
		claims, ok := fromContext(r.Context())
		if !ok {
			return ErrNoClaims
		}

		for _, scope := range scopes {
			if claims.Scope.Contains(scope) {
				return nil
			}
		}

		return ErrMissingScope
	}
}

// WithAllScopes verifies that all the given scopes are in the JWT.
func WithAllScopes(scopes ...string) rule {
	return func(r *http.Request) error {
		claims, ok := fromContext(r.Context())
		if !ok {
			return ErrNoClaims
		}

		for _, scope := range scopes {
			if !claims.Scope.Contains(scope) {
				return ErrMissingScope
			}
		}

		return nil
	}
}

// WithScopes verifies that the given scopes are in the JWT.
// DEPRECATED!
func WithScopes(scopes ...string) rule {
	return WithAllScopes(scopes...)
}
