package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/d4l-data4life/go-svc/pkg/log"
	"github.com/d4l-data4life/go-svc/pkg/middlewares"

	"github.com/stretchr/testify/assert"
)

func TestTenantID(t *testing.T) {
	expectedTenantID := "charite"

	req, _ := http.NewRequest(http.MethodGet, "", nil)
	req.Header.Add(middlewares.TenantIDHeaderName, expectedTenantID)
	res := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Context().Value(log.TenantIDContextKey).(string)
		assert.Equal(t, expectedTenantID, traceID)
	})

	traceMiddleware := middlewares.Tenant(handler)
	traceMiddleware.ServeHTTP(res, req)
}
