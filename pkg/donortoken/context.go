package donortoken

import "context"

type contextKey struct{}

// WithClaims returns a copy of ctx carrying the verified donor claims (used by the adapters
// and by tests of handlers behind them).
func WithClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, contextKey{}, claims)
}

// FromContext returns the donor claims put into the context by HTTPMiddleware,
// UnaryServerInterceptor or WithClaims.
func FromContext(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(contextKey{}).(Claims)
	return claims, ok
}
