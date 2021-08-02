package d4lcontext

import (
	"context"
	"net/http"

	uuid "github.com/gofrs/uuid"
)

type contextKey string

const (
	// UserIDContextKey is the key to store the user ID in the context
	UserIDContextKey contextKey = "user-id"

	// ClientIDContextKey is the key to store the client ID in the context
	ClientIDContextKey contextKey = "client-id"

	// TenantIDContextKey is the key to store the tenant ID in the context
	TenantIDContextKey contextKey = "tenant-id"
)

// GetUserID extracts the user id from a request context
func GetUserID(r *http.Request) string {
	return GetUserIDFromCtx(r.Context())
}

// GetClientID extracts the client id from a request context
func GetClientID(r *http.Request) string {
	return GetClientIDFromCtx(r.Context())
}

// GetTenantID extracts the tenant id from a request context.
func GetTenantID(r *http.Request) string {
	return GetTenantIDFromCtx(r.Context())
}

// GetUserID extracts the user id from a request context
func GetUserIDFromCtx(ctx context.Context) string {
	rawUserID := ctx.Value(UserIDContextKey)
	if userID, ok := rawUserID.(uuid.UUID); ok {
		return userID.String()
	}
	if userID, ok := rawUserID.(string); ok {
		return userID
	}
	return ""
}

// GetClientID extracts the client id from a request context
func GetClientIDFromCtx(ctx context.Context) string {
	if clientID, ok := ctx.Value(ClientIDContextKey).(string); ok {
		return clientID
	}
	return ""
}

// GetTenantID extracts the tenant id from a request context.
func GetTenantIDFromCtx(ctx context.Context) string {
	if tenantID, ok := ctx.Value(TenantIDContextKey).(string); ok {
		return tenantID
	}
	return ""
}

// WithUserID adds the user id to the request context
func WithUserID(r *http.Request, userID string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), UserIDContextKey, userID))
}

// WithClientID adds the client id to the request context
func WithClientID(r *http.Request, clientID string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ClientIDContextKey, clientID))
}

// WithTenantID adds the tenant id to the request context.
// It defaults to 'd4l' if the value is not found in the context.
func WithTenantID(r *http.Request, tenantID string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), TenantIDContextKey, tenantID))
}
