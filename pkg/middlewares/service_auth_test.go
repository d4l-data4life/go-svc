package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type DummyLogger struct {
}

func (dl DummyLogger) ErrGeneric(ctx context.Context, err error) error {
	return nil
}

func TestAuthenticator_Authenticate(t *testing.T) {
	testServiceSecret := "test-secret"
	tests := []struct {
		name              string
		AuthHeaderContent string
		expectedStatus    int
	}{
		{"valid auth header", fmt.Sprintf("Bearer %s", testServiceSecret), http.StatusOK},
		{"missing bearer", testServiceSecret, http.StatusUnauthorized},
		{"missing secret", "Bearer ", http.StatusUnauthorized},
		{"invalid secret", fmt.Sprintf("Bearer %s", "something else"), http.StatusUnauthorized},
		{"invalid format", fmt.Sprintf("bearer%s", "something else"), http.StatusUnauthorized},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			nextRequest, _ := http.NewRequest(http.MethodGet, "", nil)
			nextRequest.Header.Add("Authorization", tc.AuthHeaderContent)
			nextResponse := httptest.NewRecorder()

			authenticator := NewServiceSecretAuthenticator(testServiceSecret, DummyLogger{})

			dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			protectedHandler := authenticator.Authenticate()(dummyHandler)
			protectedHandler.ServeHTTP(nextResponse, nextRequest)
			assert.Equal(t, tc.expectedStatus, nextResponse.Code)
		})
	}
}
