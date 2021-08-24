package transport

import "net/http"

type ChainTransport struct {
	rt http.RoundTripper
}

func (t *ChainTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.rt.RoundTrip(req)
}

// Chain executes passed transports as chain in linear order starting from the first transport.
func Chain(transports ...TransportFunc) TransportFunc {
	return func(rt http.RoundTripper) http.RoundTripper {
		if rt == nil {
			rt = http.DefaultTransport
		}

		for i := len(transports) - 1; i >= 0; i-- {
			rt = transports[i](rt)
		}

		return &ChainTransport{
			rt: rt,
		}

	}
}
