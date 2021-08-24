package transport

import (
	"net/http"
)

type JSONTransport struct {
	rt http.RoundTripper
}

func (t *JSONTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Content-Type", "application/json")
	return t.rt.RoundTrip(req)
}

// JSON sets the content-type header to application/json for outgoing http requests.
func JSON(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}

	return &JSONTransport{
		rt: rt,
	}
}
