package testutils

import (
	"crypto/rsa"
	"net/http"
	"net/url"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/jwt"
)

type RequestBuilder func(*http.Request)

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

func OkHandler(w http.ResponseWriter, r *http.Request) {}
