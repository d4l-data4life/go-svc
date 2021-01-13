package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/d4lcontext"
	"github.com/gesundheitscloud/go-svc/pkg/middlewares"

	"github.com/stretchr/testify/assert"
)

func TestTenantID(t *testing.T) {
	expectedTenantID := "charite"

	req, _ := http.NewRequest(http.MethodGet, "", nil)
	req.Header.Add(middlewares.TenantIDHeaderName, expectedTenantID)
	res := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Context().Value(d4lcontext.TenantIDContextKey).(string)
		assert.Equal(t, expectedTenantID, traceID)
	})

	traceMiddleware := middlewares.Tenant(handler)
	traceMiddleware.ServeHTTP(res, req)
}
