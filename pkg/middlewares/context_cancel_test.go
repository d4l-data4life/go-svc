package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockHandler struct {
	returnCode int
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(h.returnCode)
}

func TestHandleContextCancel(t *testing.T) {
	for _, tc := range [...]struct {
		name           string
		handler        http.Handler
		expectedStatus int
		cancelContext  bool
	}{
		{
			name: "should not override the status code for normal context",
			handler: &mockHandler{
				returnCode: http.StatusOK,
			},
			cancelContext:  false,
			expectedStatus: http.StatusOK,
		},
		{
			name: "should override the status if the context was cancelled",
			handler: &mockHandler{
				returnCode: http.StatusOK,
			},
			cancelContext:  true,
			expectedStatus: 499,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			req, _ := http.NewRequest(http.MethodPost, "", nil)
			req = req.WithContext(ctx)

			res := httptest.NewRecorder()

			if tc.cancelContext {
				cancel()
			}

			HandleContextCancel(tc.handler).ServeHTTP(res, req)

			result := res.Result()
			defer result.Body.Close()

			if result.StatusCode != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, result.StatusCode)
			}
		})
	}
}
