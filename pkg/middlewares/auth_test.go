package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/d4l-data4life/go-svc/pkg/instrumented"

	"github.com/stretchr/testify/assert"
)

func TestServiceSecret(t *testing.T) {
	validAuthHeader := "service-secret"

	handlerFactory := instrumented.NewHandlerFactory(
		"d4l",
		instrumented.DefaultInstrumentInitOptions,
		instrumented.DefaultInstrumentOptions,
	)
	auth := NewAuthentication(validAuthHeader, handlerFactory)
	tests := []struct {
		name              string
		AuthHeaderContent string
		expectedStatus    int
	}{
		{"valid Service Auth", validAuthHeader, http.StatusOK},
		{"invalid Service Auth", "random", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/route", nil)
			req.Header.Add(AuthHeaderName, tt.AuthHeaderContent)
			res := httptest.NewRecorder()

			// target handler after auth check
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			authMiddleware := auth.ServiceSecret(handler)
			authMiddleware.ServeHTTP(res, req)
			assert.Equal(t, tt.expectedStatus, res.Code)
		})
	}
}

func TestGetAuthSecret(t *testing.T) {
	tests := []struct {
		name              string
		authHeaderContent string
		expectedSecret    string
	}{
		{"Service secret without prefix", "secret", "secret"},
		{"Service secret with prefix", "Bearer anothersecret", "anothersecret"},
	}

	auth := &Auth{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/route", nil)
			req.Header.Add(AuthHeaderName, tt.authHeaderContent)

			authToken, _ := auth.getAuthSecret(req)
			assert.Equal(t, tt.expectedSecret, authToken)
		})
	}
}
