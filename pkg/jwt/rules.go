package jwt

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
)

type rule func(*http.Request, *Claims) error

// WithOwner verifies that the given function returns the UUID of the JWT's subject ID.
func WithOwner(owner func(r *http.Request) uuid.UUID) rule {
	return func(r *http.Request, claims *Claims) error {
		haveID := owner(r)
		if haveID != claims.Subject.ID || haveID == uuid.Nil {
			return ErrSubjectNotOwner
		}

		return nil
	}
}

// WithGorillaOwner provides a Gorilla/Mux specific solution for parsing the owner from the path
// and checking that it mathes the subject ID from the JWT claims
func WithGorillaOwner(ownerKey string) rule {
	return func(r *http.Request, claims *Claims) error {
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
	return func(r *http.Request, claims *Claims) error {
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
	return func(r *http.Request, claims *Claims) error {
		for _, scope := range scopes {
			if claims.Scope.Contains(scope) {
				return nil
			}
		}

		return fmt.Errorf("%w: expected ANY scope of %v, got %v",
			ErrMissingScope, scopes, claims.Scope.Tokens)
	}
}

// WithAllScopes verifies that all the given scopes are in the JWT.
func WithAllScopes(scopes ...string) rule {
	return func(r *http.Request, claims *Claims) error {
		for _, scope := range scopes {
			if !claims.Scope.Contains(scope) {
				return fmt.Errorf("%w: expected ALL scopes %v, got %v",
					ErrMissingScope, scopes, claims.Scope.Tokens)
			}
		}

		return nil
	}
}
