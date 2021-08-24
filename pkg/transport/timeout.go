package transport

import (
	"context"
	"net/http"
	"time"
)

type TimeoutTransport struct {
	timeout time.Duration
	rt      http.RoundTripper
}

func (t *TimeoutTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	req = req.WithContext(ctx)

	return t.rt.RoundTrip(req)
}

// Timeout rewraps the outgoing http request in a new context, which expires after passed duration.
func Timeout(t time.Duration) TransportFunc {
	return func(rt http.RoundTripper) http.RoundTripper {
		if rt == nil {
			rt = http.DefaultTransport
		}

		return &TimeoutTransport{
			timeout: t,
			rt:      rt,
		}
	}
}
