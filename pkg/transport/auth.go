package transport

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrBearerSecretMissing happens when password is missing.
var ErrBearerSecretMissing = errors.New("no bearer secret provided")

type BearerAuthTransport struct {
	secret string

	rt http.RoundTripper
}

func (t *BearerAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.secret == "" {
		return nil, ErrBearerSecretMissing
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.secret))

	return t.rt.RoundTrip(req)
}

// BearerAuth adds the passed secret as bearer authorization header to outgoing http requests.
func BearerAuth(secret string) TransportFunc {
	return func(rt http.RoundTripper) http.RoundTripper {
		if rt == nil {
			rt = http.DefaultTransport
		}

		return &BearerAuthTransport{
			secret: secret,
			rt:     rt,
		}
	}
}
