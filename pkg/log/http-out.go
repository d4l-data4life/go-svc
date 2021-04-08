package log

import (
	"net/http"
	"time"
)

type Transport struct {
	next   http.RoundTripper
	logger *Logger
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Context() != nil {
		traceID, _ := req.Context().Value(TraceIDContextKey).(string)
		if traceID != "" {
			req.Header.Set(TraceIDHeaderKey, traceID)
		}
	}

	reqTime := time.Now()
	_ = t.logger.HttpOutReq(req)

	resp, err := t.next.RoundTrip(req)

	_ = t.logger.HttpOutResponse(req, resp, reqTime)

	return resp, err
}

func (l *Logger) LoggedTransport(next http.RoundTripper) *Transport {
	if next == nil {
		next = http.DefaultTransport
	}
	return &Transport{
		next:   next,
		logger: l,
	}
}
