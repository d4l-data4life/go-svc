package middlewares

import (
	"context"
	"net/http"

	"github.com/d4l-data4life/go-svc/pkg/log"
)

const (
	// TenantIDHeaderName is the name of the tenant id header
	TenantIDHeaderName string = "X-Tenant-ID"
)

// Tenant middleware copies the tenantID from the req header to the req context
func Tenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get(TenantIDHeaderName)
		ctx := context.WithValue(r.Context(), log.TenantIDContextKey, tenantID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
