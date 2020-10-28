package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gesundheitscloud/go-svc/pkg/d4lcontext"
	"github.com/gesundheitscloud/go-svc/pkg/instrumented"
)

func TestXSRFTokenGeneration(t *testing.T) {
	xsrfSecret := "xsrf-secret"
	xsrfHeader := "X-Csrf-Token"
	handlerFactory := instrumented.NewHandlerFactory("d4l", instrumented.DefaultInstrumentInitOptions, instrumented.DefaultInstrumentOptions)

	request, _ := http.NewRequest("method", "url", nil)
	response := httptest.NewRecorder()

	e := NewXSRFHandler(xsrfSecret, xsrfHeader, handlerFactory)
	ctx := request.Context()
	ctx = context.WithValue(ctx, d4lcontext.UserIDContextKey, "123e4567-e89b-12d3-a456-426655440000")
	request = request.WithContext(ctx)
	e.XSRF(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

	assert.NotEmpty(t, response.Header().Get(e.HeaderName))
}
