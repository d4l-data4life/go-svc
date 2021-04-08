package jwt

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

type contextKey int

const (
	// jwtClaimsContextKey holds the key used to store the JWT Claims in the
	// context.
	jwtClaimsContextKey contextKey = iota
)

// NewContext embeds the claims in the returned context, that is a
// child of ctx.
func NewContext(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, jwtClaimsContextKey, claims)
}

// fromContext extracts the claims from the given context. The
// returned boolean is false if no claims are found.
func fromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(jwtClaimsContextKey).(*Claims)
	return claims, ok
}

// GetSubjectID returns the subject's ID from the JWT that has been parsed and stored as context
func GetSubjectID(ctx context.Context) (uuid.UUID, error) {
	claims, ok := fromContext(ctx)
	if !ok {
		return uuid.Nil, ErrNoClaims
	}

	subjectID := claims.Subject.ID
	if subjectID == uuid.Nil {
		return subjectID, errors.New("subjectID not set in claims")
	}

	return subjectID, nil
}

// GetUserID returns the user's ID from the JWT that has been parsed and stored as context
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	claims, ok := fromContext(ctx)
	if !ok {
		return uuid.Nil, ErrNoClaims
	}

	return claims.UserID, nil
}

// GetAppID returns the app's ID from the JWT that has been parsed and stored as context
func GetAppID(ctx context.Context) (uuid.UUID, error) {
	claims, ok := fromContext(ctx)
	if !ok {
		return uuid.Nil, ErrNoClaims
	}

	return claims.AppID, nil
}

// GetClientID returns the client ID from the JWT that has been parsed and stored as context
func GetClientID(ctx context.Context) (string, error) {
	claims, ok := fromContext(ctx)
	if !ok {
		return "", ErrNoClaims
	}

	return claims.ClientID, nil
}

// GetTenantID returns the tenant ID from the JWT that has been parsed and stored as context
func GetTenantID(ctx context.Context) (string, error) {
	claims, ok := fromContext(ctx)
	if !ok {
		return "", ErrNoClaims
	}

	return claims.TenantID, nil
}

// GetTagsInScope returns the list of tagged scopes from the JWT that has been parsed and stored as context
func GetTagsInScope(ctx context.Context) ([]Tag, error) {
	claims, ok := fromContext(ctx)
	if !ok {
		return nil, ErrNoClaims
	}

	return claims.Scope.Tags()
}
