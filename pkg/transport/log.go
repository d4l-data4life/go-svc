package transport

import (
	"net/http"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/log"
)

type LogTransport struct {
	rt     http.RoundTripper
	logger *log.Logger
	obf    map[string][]log.HTTPObfuscator
}

func (t *LogTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqTime := time.Now()
	_ = t.logger.HttpOutReq(req, t.obf)

	res, err := t.rt.RoundTrip(req)
	_ = t.logger.HttpOutResponse(req, res, reqTime, t.obf)

	return res, err
}

// Log adds entries for outgoing and incoming http request to the log stream.
func Log(logger *log.Logger, options ...func(*LogTransport)) TransportFunc {
	return func(rt http.RoundTripper) http.RoundTripper {
		if rt == nil {
			rt = http.DefaultTransport
		}

		lt := &LogTransport{
			rt:     rt,
			logger: logger,
			obf:    make(map[string][]log.HTTPObfuscator),
		}

		for _, apply := range options {
			apply(lt)
		}

		return lt
	}
}

func WithObfuscators(o ...log.HTTPObfuscator) func(*LogTransport) {
	return func(l *LogTransport) {
		for _, obf := range o {
			key := log.ObfuscatorKey(obf.GetEventType(), obf.GetReqMethod())
			l.obf[key] = append(l.obf[key], obf)
		}
	}
}
