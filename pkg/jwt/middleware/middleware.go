package middleware

import (
	"context"
	"crypto/rsa"
	"net/http"

	"github.com/gesundheitscloud/go-svc/pkg/jwt"

	"github.com/gofrs/uuid"
)

type logger interface {
	ErrUserAuth(context.Context, error) error
	InfoGeneric(context.Context, string) error
}

// Middleware is the default middleware type used everywhere.
type Middleware func(http.Handler) http.Handler

type verifier struct {
	pk *rsa.PublicKey
	l  logger

	ownerFunc func(*http.Request) uuid.UUID

	scopes []string
}

// Auth creates a new middleware that checks the JWT in flight. The public key is
// used to verify the signature. The options specify the expected token.
func Auth(pk *rsa.PublicKey, l logger, opts ...func(*verifier)) Middleware {
	v := verifier{
		pk: pk, l: l,
	}

	for _, opt := range opts {
		opt(&v)
	}

	return jwt.
		New(v.pk, v.l).
		Verify(
			jwt.WithOwner(v.ownerFunc),
			jwt.WithScopes(v.scopes...),
		)
}

// WithScopes is a function that appends a custom set of scopes.
func WithScopes(scopes ...string) func(v *verifier) {
	return func(v *verifier) {
		v.scopes = append(v.scopes, scopes...)
	}
}
