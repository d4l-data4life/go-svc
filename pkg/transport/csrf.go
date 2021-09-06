package transport

import "net/http"

type CSRFTransport struct {
	rt    http.RoundTripper
	token string
}

func (t *CSRFTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.token != "" {
		req.Header.Add("x-csrf-token", t.token)
	}

	res, err := t.rt.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	t.token = res.Header.Get("x-csrf-token")

	return res, nil
}

func CSRF(token string) TransportFunc {
	return func(rt http.RoundTripper) http.RoundTripper {
		if rt == nil {
			rt = http.DefaultTransport
		}

		return &CSRFTransport{
			rt:    rt,
			token: token,
		}
	}
}
