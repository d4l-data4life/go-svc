package transport

import (
	"fmt"
	"net/http"

	"github.com/gesundheitscloud/go-svc/pkg/d4lcontext"
)

type AccessTokenTransport struct {
	rt http.RoundTripper
}

func (t *AccessTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	accessTokenVal := req.Context().Value(d4lcontext.AccessTokenContextKey)
	if accessTokenVal == nil {
		return t.rt.RoundTrip(req)
	}

	accessToken, ok := accessTokenVal.(string)
	if !ok {
		return nil, fmt.Errorf("failed casting access token to string")
	}

	if accessToken != "" {
		req.Header.Add("Authorization", accessToken)
	}

	return t.rt.RoundTrip(req)
}

// AccessToken parses the access token from the request context to the request header.
func AccessToken(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}

	return &AccessTokenTransport{rt: rt}
}
