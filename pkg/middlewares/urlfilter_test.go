package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrlValidator(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{"Success", "/api/v1/consentDocuments?key=d4l.registry\u0026version=1", http.StatusOK},
		{"Invalid 1", "/api/v1/fhir/stu3/foo/?url=www.example.org&version=%001", http.StatusBadRequest},
		{"Invalid 2", "/api/v1/consentDocuments?key=%00\u0026version=1", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, tt.url, nil)
			res := httptest.NewRecorder()

			// target handler after query validation
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			filterMiddleware := UrlValidator(handler)
			filterMiddleware.ServeHTTP(res, req)
			assert.Equal(t, tt.expectedStatus, res.Code)
		})
	}
}
