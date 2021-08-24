package transport

import (
	"fmt"
	"net/http"

	"github.com/gesundheitscloud/go-svc/pkg/log"
)

type TraceIDTransport struct {
	rt http.RoundTripper
}

func (t *TraceIDTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	traceIDVal := req.Context().Value(log.TraceIDContextKey)
	if traceIDVal == nil {
		return t.rt.RoundTrip(req)
	}

	traceID, ok := traceIDVal.(string)
	if !ok {
		return nil, fmt.Errorf("failed casting trace-id to string")
	}

	if traceID != "" {
		req.Header.Add(log.TraceIDHeaderKey, traceID)
	}

	return t.rt.RoundTrip(req)
}

// TraceID parses the trace-id from the request context to the request header.
func TraceID(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}

	return &TraceIDTransport{rt: rt}
}
