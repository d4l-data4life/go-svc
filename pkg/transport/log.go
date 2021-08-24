package transport

import (
	"net/http"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/log"
)

type LogTransport struct {
	rt     http.RoundTripper
	logger *log.Logger
}

func (t *LogTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqTime := time.Now()
	_ = t.logger.HttpOutReq(req)

	res, err := t.rt.RoundTrip(req)
	_ = t.logger.HttpOutResponse(req, res, reqTime)

	return res, err
}

// Log adds entries for outgoing and incoming http request to the log stream.
func Log(logger *log.Logger) TransportFunc {
	return func(rt http.RoundTripper) http.RoundTripper {
		if rt == nil {
			rt = http.DefaultTransport
		}

		return &LogTransport{
			rt:     rt,
			logger: logger,
		}
	}
}
