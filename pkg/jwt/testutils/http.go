package testutils

import (
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/jwt"
)

type RequestBuilder func(*http.Request) error

func BuildRequest(fns ...RequestBuilder) *http.Request {
	r := httptest.NewRequest("", "/some/url", nil)

	for _, fn := range fns {
		_ = fn(r)
	}

	return r
}

func WithHeader(header map[string]string) RequestBuilder {
	return func(r *http.Request) error {
		for k, v := range header {
			r.Header.Add(k, v)
		}

		return nil
	}
}

func WithAuthHeader(key *rsa.PrivateKey, options ...jwt.TokenOption) RequestBuilder {
	return func(r *http.Request) error {
		options = append(
			// add a default expiration time as the token is not valid without one
			[]jwt.TokenOption{jwt.WithExpirationTime(time.Now().Add(1 * time.Minute))},
			options...,
		)
		t, err := jwt.CreateAccessToken(key, options...)
		if err != nil {
			return err
		}

		r.Header.Add("Authorization", "Bearer "+t.AccessToken)

		return nil
	}
}

func WithTargetURL(target string) func(*http.Request) error {
	return func(req *http.Request) error {
		url, err := url.Parse(target)
		if err != nil {
			panic(err)
		}
		req.URL = url
		return nil
	}
}

func OkHandler(w http.ResponseWriter, r *http.Request) {}
