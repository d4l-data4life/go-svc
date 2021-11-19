package log

import (
	"net/http"
	"time"
)

// Transport is deprecated, please use go-svc/pkg/transport
type Transport struct {
	next   http.RoundTripper
	logger *Logger
}

// RoundTrip is deprecated, please use go-svc/pkg/transport
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Context() != nil {
		traceID, _ := req.Context().Value(TraceIDContextKey).(string)
		if traceID != "" {
			req.Header.Set(TraceIDHeaderKey, traceID)
		}
	}

	reqTime := time.Now()
	_ = t.logger.HttpOutReq(req, nil)

	resp, err := t.next.RoundTrip(req)

	_ = t.logger.HttpOutResponse(req, resp, reqTime, nil)

	return resp, err
}

// LoggedTransport is deprecated, please use go-svc/pkg/transport
func (l *Logger) LoggedTransport(next http.RoundTripper) *Transport {
	if next == nil {
		next = http.DefaultTransport
	}
	return &Transport{
		next:   next,
		logger: l,
	}
}
