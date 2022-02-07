package testutils

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/jwt"
)

type RequestBuilder func(*http.Request)

func BuildRequest(fns ...RequestBuilder) *http.Request {
	r := httptest.NewRequest("", "/some/url", nil)

	for _, fn := range fns {
		fn(r)
	}

	return r
}

func WithHeader(header map[string]string) RequestBuilder {
	return func(r *http.Request) {
		for k, v := range header {
			r.Header.Add(k, v)
		}
	}
}

func WithMethod(m string) RequestBuilder {
	return func(r *http.Request) {
		r.Method = m
	}
}

func WithAuthHeader(key *rsa.PrivateKey, options ...jwt.TokenOption) RequestBuilder {
	return func(r *http.Request) {
		options = append(
			// add a default expiration time as the token is not valid without one
			[]jwt.TokenOption{jwt.WithExpirationTime(time.Now().Add(1 * time.Minute))},
			options...,
		)
		t, err := jwt.CreateAccessToken(key, options...)
		if err != nil {
			panic(fmt.Errorf("can't create access token: %w", err))
		}

		r.Header.Add("Authorization", "Bearer "+t.AccessToken)
	}
}

// WithFormAccessToken creates a valid JWT and adds it to the request form body
// in the `access_token` field
func WithFormAccessToken(key *rsa.PrivateKey, options ...jwt.TokenOption) RequestBuilder {
	return func(r *http.Request) {
		options = append(
			// add a default expiration time as the token is not valid without one
			[]jwt.TokenOption{jwt.WithExpirationTime(time.Now().Add(1 * time.Minute))},
			options...,
		)
		t, err := jwt.CreateAccessToken(key, options...)
		if err != nil {
			panic(err)
		}

		if r.Form == nil {
			r.Form = url.Values{}
		}
		r.Form.Add("access_token", t.AccessToken)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
}

// WithCookieAccessToken creates a valid JWT and adds it to the request cookie with
// jwt.AccessCookieName as cookie name
func WithCookieAccessToken(key *rsa.PrivateKey, options ...jwt.TokenOption) RequestBuilder {
	return func(r *http.Request) {
		options = append(
			// add a default expiration time as the token is not valid without one
			[]jwt.TokenOption{jwt.WithExpirationTime(time.Now().Add(1 * time.Minute))},
			options...,
		)
		t, err := jwt.CreateAccessToken(key, options...)
		if err != nil {
			panic(err)
		}

		ac := http.Cookie{
			Name:  jwt.AccessCookieName,
			Value: t.AccessToken,
		}

		r.AddCookie(&ac)
	}
}

// WithCookie adds a cookie to a request cookie name
func WithCookie(c *http.Cookie) RequestBuilder {
	return func(r *http.Request) {
		r.AddCookie(c)
	}
}

func WithTargetURL(target string) func(*http.Request) {
	return func(req *http.Request) {
		url, err := url.Parse(target)
		if err != nil {
			panic(err)
		}
		req.URL = url
	}
}

// WithForm returns a function that modifies the given request
// by setting the given key/value pairs as the request form.
// It also sets the 'Content-Type' header to 'application/x-www-form-urlencoded'
func WithForm(vals ...func(*url.Values)) func(*http.Request) {
	return func(req *http.Request) {
		data := make(url.Values)
		for _, valFunc := range vals {
			valFunc(&data)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Body = ioutil.NopCloser(strings.NewReader(data.Encode()))
	}
}

// WithValue returns a function that modifies the url values by adding a key-value pair.
// This can be used to add values to a form request body
func WithValue(key, value string) func(*url.Values) {
	return func(data *url.Values) {
		data.Add(key, value)
	}
}

func OkHandler(w http.ResponseWriter, r *http.Request) {}
