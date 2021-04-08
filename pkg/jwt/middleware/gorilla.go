package middleware

import (
	"crypto/rsa"
	"net/http"

	"github.com/gesundheitscloud/go-svc/pkg/jwt"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
)

func uuidOrNil(id string) uuid.UUID {
	uid, err := uuid.FromString(id)
	if err != nil {
		return uuid.Nil
	}
	return uid
}

// WithGorillaMux makes the verifier viable with github.com/gorilla/mux.
func WithGorillaMux(ownerKey string) func(*verifier) {
	return func(v *verifier) {
		v.ownerFunc = func(r *http.Request) uuid.UUID {
			vars := mux.Vars(r)
			if value, ok := vars[ownerKey]; ok {
				return uuidOrNil(value)
			}

			return uuid.Nil
		}
	}
}

// GorillaAuth is a wrapper for the authenticator struct in the go-jwt package. Abstracts the gorilla specific logic away.
type GorillaAuth struct {
	authenticator *jwt.Authenticator
	ownerFunc     func(*http.Request) uuid.UUID
}

// NewGorillaAuth creates a GorillaAuth that can verify handlers by scopes.
func NewGorillaAuth(
	pk *rsa.PublicKey,
	l logger,
	urlParam string,
) *GorillaAuth {
	authenticator := jwt.New(pk, l)
	ownerFunc := ownerFunc(urlParam)

	return &GorillaAuth{authenticator, ownerFunc}
}

func ownerFunc(urlParam string) func(r *http.Request) uuid.UUID {
	return func(r *http.Request) uuid.UUID {
		vars := mux.Vars(r)
		if value, ok := vars[urlParam]; ok {
			return uuidOrNil(value)
		}

		return uuid.Nil
	}
}

// WithScopes returns a middleware that verifies the owner and the expected scopes.
func (a *GorillaAuth) WithScopes(s ...string) func(http.Handler) http.Handler {
	return a.authenticator.Verify(
		jwt.WithOwner(a.ownerFunc),
		jwt.WithScopes(s...),
	)
}

// WithScopes returns a middleware that verifies the expected scopes only (not the owner!).
func (a *GorillaAuth) WithScopesNoOwner(s ...string) func(http.Handler) http.Handler {
	return a.authenticator.Verify(
		jwt.WithScopes(s...),
	)
}
