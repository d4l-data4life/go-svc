package d4lcontext

import (
	"errors"
	"net/http"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
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

// DefaultTenantID is the tenant id used if is missing in the JWT
var DefaultTenantID = "d4l"

// GetUserID is used by the logger to extract the user id from a request context
func GetUserID(r *http.Request) string {
	if userID, ok := r.Context().Value(UserIDContextKey).(uuid.UUID); ok {
		return userID.String()
	}
	return ""
}

// GetClientID is used by the logger to extract the client id from a request context
func GetClientID(r *http.Request) string {
	if clientID, ok := r.Context().Value(ClientIDContextKey).(string); ok {
		return clientID
	}
	return ""
}

// GetTenantID is used by the logger to extract the tenant id from a request context
func GetTenantID(r *http.Request) string {
	if tenantID, ok := r.Context().Value(TenantIDContextKey).(string); ok && tenantID != "" {
		return tenantID
	}
	return DefaultTenantID
}

// ParseRequesterID returns the requester account id from context (only for protected endpoints)
func ParseRequesterID(w http.ResponseWriter, r *http.Request) (requesterID uuid.UUID, err error) {
	requester := r.Context().Value(UserIDContextKey)
	if requester == nil {
		err := errors.New("missing account id")
		logging.LogErrorfCtx(r.Context(), err, "error parsing Requester UUID")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return uuid.Nil, err
	}

	switch id := requester.(type) {
	case string:
		requesterID, err = uuid.FromString(id)
		if err != nil || requesterID == uuid.Nil {
			err := errors.New("malformed Account ID")
			logging.LogErrorfCtx(r.Context(), err, "error parsing Requester UUID")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return uuid.Nil, err
		}
	case uuid.UUID:
		requesterID = id
	}

	return requesterID, nil
}
