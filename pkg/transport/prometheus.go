package transport

import (
	"net/http"

	"github.com/gesundheitscloud/go-svc/pkg/prom"
)

// Prometheus adds monitoring to http requests.
func Prometheus(name string) TransportFunc {
	return func(rt http.RoundTripper) http.RoundTripper {
		if rt == nil {
			rt = http.DefaultTransport
		}

		return prom.NewRoundTripperInstrumenter().Instrument(name, rt)
	}
}
