package middlewares

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/gesundheitscloud/go-svc/pkg/log"
	"github.com/gesundheitscloud/go-svc/pkg/logging"
)

// Trace middleware copies the traceID from the req header to the req context
// if no traceID is present it creates a new one
func Trace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get(log.TraceIDHeaderKey)
		if traceID == "" {
			traceID = GenerateTraceID()
		}
		c := context.WithValue(r.Context(), log.TraceIDContextKey, traceID)
		next.ServeHTTP(w, r.WithContext(c))
	})
}

// GenerateTraceID generates a new traceID with length 32
// on error returns an empty string to fail silently
func GenerateTraceID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

// TraceTransport copies the traceID from the req context to the req header
// after setting the header the default http transport is called
// if no traceID is present it creates a new one
type TraceTransport struct {
}

// RoundTrip implements the http.RoundTripper interface
func (t *TraceTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	traceID, ok := r.Context().Value(log.TraceIDContextKey).(string)
	if !ok {
		traceID = GenerateTraceID()
		r = r.WithContext(context.WithValue(r.Context(), log.TraceIDContextKey, traceID))
		err := errors.New("context is missing trace-id, setting new trace-id")
		logging.LogWarningfCtx(r.Context(), err, "set the trace-id in the request context")
	}
	r.Header.Set(log.TraceIDHeaderKey, traceID)

	res, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	return res, nil
}
