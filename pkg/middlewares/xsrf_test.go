package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/d4lcontext"
	"github.com/gesundheitscloud/go-svc/pkg/instrumented"
	uuid "github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/xsrftoken"
)

func TestXSRF(t *testing.T) {
	xsrfSecret := "xsrf-secret"
	xsrfHeader := "X-Csrf-Token"

	handlerFactory := instrumented.NewHandlerFactory("d4l", instrumented.DefaultInstrumentInitOptions, instrumented.DefaultInstrumentOptions)
	xsrfMiddleware := NewXSRF(xsrfSecret, xsrfHeader, handlerFactory, XSRFWithLatencyBuckets([]float64{4, 8, 16}), XSRFWithSizeBuckets([]float64{4, 8, 16}))
	xsrfToken := xsrftoken.Generate(xsrfSecret, uuid.FromStringOrNil("123e4567-e89b-12d3-a456-426655440000").String(), "")

	tests := []struct {
		name           string
		xsrftoken      string
		expectedStatus int
	}{
		{"backwards compatibility", xsrfToken, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, "/route", nil)
			ctx := context.WithValue(req.Context(), d4lcontext.UserIDContextKey, "123e4567-e89b-12d3-a456-426655440000")
			req = req.WithContext(ctx)
			req.Header.Add(xsrfHeader, tt.xsrftoken)

			res := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			xsrfMiddleware.XSRF(handler).ServeHTTP(res, req)
			assert.Equal(t, tt.expectedStatus, res.Code)
		})
	}

}
