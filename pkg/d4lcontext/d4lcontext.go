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

	// AccessTokenContextKey is the key to store the access token in the context
	AccessTokenContextKey contextKey = "access-token"
)

// GetUserID extracts the user id from a request context
func GetUserID(r *http.Request) uuid.UUID {
	return GetUserIDFromCtx(r.Context())
}

// GetUserIDString extracts the user id as string from a request context
func GetUserIDString(r *http.Request) string {
	return GetUserIDFromCtx(r.Context()).String()
}

// GetClientID extracts the client id from a request context
func GetClientID(r *http.Request) string {
	return GetClientIDFromCtx(r.Context())
}

// GetTenantID extracts the tenant id from a request context.
func GetTenantID(r *http.Request) string {
	return GetTenantIDFromCtx(r.Context())
}

// GetAccessToken extracts the access token from a request context.
func GetAccessToken(r *http.Request) string {
	return GetAccessTokenFromCtx(r.Context())
}

// GetUserID extracts the user id from a request context
func GetUserIDFromCtx(ctx context.Context) uuid.UUID {
	rawUserID := ctx.Value(UserIDContextKey)
	if userID, ok := rawUserID.(uuid.UUID); ok {
		return userID
	}
	return uuid.Nil
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

// GetAccessTokenFromCtx extracts the access token from the context.
func GetAccessTokenFromCtx(ctx context.Context) string {
	if accessToken, ok := ctx.Value(AccessTokenContextKey).(string); ok {
		return accessToken
	}
	return ""
}

// WithUserID adds the user id to the request context
func WithUserID(r *http.Request, userID uuid.UUID) *http.Request {
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

// WithAccessToken adds the access token to the request context.
func WithAccessToken(r *http.Request, accessToken string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), AccessTokenContextKey, accessToken))
}
