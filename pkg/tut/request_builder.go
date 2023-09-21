package tut

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gesundheitscloud/go-svc/pkg/jwt"

	"github.com/gofrs/uuid"
)

// Request creates a new request
// and applies the given request options to it.
func Request(opts ...func(*http.Request)) *http.Request {
	req, err := http.NewRequest("", "", bytes.NewReader([]byte{}))
	if err != nil {
		panic(err)
	}

	for _, opt := range opts {
		opt(req)
	}

	return req
}

// ReqWithContext returns a function that modifies the given request
// by setting the given context on the request object.
func ReqWithContext(ctx context.Context) func(*http.Request) {
	return func(req *http.Request) {
		*req = *req.WithContext(ctx)
	}
}

// ReqWithContextValue returns a function that modifies the given request
// by setting the given context on the request object.
func ReqWithContextValue(key, value interface{}) func(*http.Request) {
	return func(req *http.Request) {
		*req = *(req.WithContext(context.WithValue(req.Context(), key, value)))
	}
}

// ReqWithClaimsInContext returns a function that adds the given claims to the context
// so that they can be found by the jwt context lib.
func ReqWithClaimsInContext(claims *jwt.Claims) func(*http.Request) {
	return func(req *http.Request) {
		*req = *(req.WithContext(jwt.NewContext(req.Context(), claims)))
	}
}

// ReqWithJSONBody returns a function that modifies the given request
// by setting the marshaling the given interface{} to the request body.
func ReqWithJSONBody(v interface{}) func(*http.Request) {
	return func(req *http.Request) {
		jsonValue, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}

		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Body = io.NopCloser(strings.NewReader(string(jsonValue)))
	}
}

// ReqWithJSONBodyString returns a function that modifies the given request
// by setting the given string to the request body.
func ReqWithJSONBodyString(json string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Body = io.NopCloser(strings.NewReader(json))
	}
}

// ReqWithTextBody returns a function that modifies the given request
// by setting request body to the given string.
func ReqWithTextBody(body string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set("Content-Type", "text/plain")
		req.Body = io.NopCloser(strings.NewReader(body))
	}
}

// ReqWithByteArrayBody returns a function that modifies the given request
// by setting request body to the given byte array.
// Sets content type to "application/octet-stream".
func ReqWithByteArrayBody(body []byte) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Body = io.NopCloser(bytes.NewReader(body))
	}
}

// ReqWithTargetURL returns a function that modifies the given request
// by setting the given target URL.
// If parsing failed, it panics.
func ReqWithTargetURL(target string) func(*http.Request) {
	return func(req *http.Request) {
		url, err := url.Parse(target)
		if err != nil {
			panic(err)
		}
		req.URL = url
	}
}

// ReqWithMethod returns a function that modifies the given request
// by setting its HTTP method.
func ReqWithMethod(m string) func(*http.Request) {
	return func(req *http.Request) {
		req.Method = m
	}
}

// ReqWithFormValue returns a function that modifies the given request
// by setting the given key/value pairs as the request form.
// It also sets the 'Content-Type' header to 'application/x-www-form-urlencoded'
func ReqWithFormValue(key, value string) func(*http.Request) {
	return func(req *http.Request) {
		if req.Form == nil {
			req.Form = url.Values{}
		}
		req.Form.Add(key, value)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
}

// ReqWithRequestParams returns a function that modifies the given request
// by setting the given key/value pairs as request parameters.
func ReqWithRequestParams(uri string, vals ...func(*url.Values)) func(*http.Request) {
	return func(req *http.Request) {
		data := make(url.Values)
		for _, valFunc := range vals {
			valFunc(&data)
		}

		rawURL := uri + "?" + data.Encode()
		url, err := url.Parse(rawURL)
		if err != nil {
			panic(err)
		}

		req.URL = url
	}
}

// ReqWithFormBody returns a function that modifies the given request
// by setting the given key/value pairs as the request form body.
// It also sets the 'Content-Type' header to 'application/x-www-form-urlencoded'
func ReqWithFormBody(vals ...func(*url.Values)) func(*http.Request) {
	return func(req *http.Request) {
		data := make(url.Values)
		for _, valFunc := range vals {
			valFunc(&data)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Body = io.NopCloser(strings.NewReader(data.Encode()))
	}
}

// ReqWithCookies returns a function that modifies the given request
// by setting the given cookie
func ReqWithCookies(cookies ...*http.Cookie) func(*http.Request) {
	return func(req *http.Request) {
		for _, c := range cookies {
			req.AddCookie(c)
		}
	}
}

// ReqWithHeader returns a function that modifies the given request
// by setting the given header.
//
//nolint:unparam - will be useful in the future and therefore not fixed
func ReqWithHeader(key, value string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

// ReqWithValue returns a function that modifies the url values by adding a key-value pair
func ReqWithValue(key, value string) func(*url.Values) {
	return func(data *url.Values) {
		data.Add(key, value)
	}
}

// ReqWithQuery adds query parameters to the url.
func ReqWithQuery(values url.Values) func(*http.Request) {
	return func(req *http.Request) {
		req.URL.RawQuery = values.Encode()
	}
}

// ReqWithAppendToHeader adds header key / values without overwriting values.
func ReqWithAppendToHeader(key, value string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Add(key, value)
	}
}

// ReqWithAccessCookie allows to add an auth cookie to a request. This is a convenience shortcut for
// creating the auth cookie (using GenerateAccessToken) and adding it to the request using WithCookies
func ReqWithAccessCookie(userID uuid.UUID, privateKey *rsa.PrivateKey,
	claimsOptions ...jwt.TokenOption) func(*http.Request) {
	token := GenerateAccessToken(userID, privateKey, claimsOptions...)

	return func(req *http.Request) {
		req.AddCookie(&http.Cookie{
			Name:  jwt.AccessCookieName,
			Value: token,
		})
	}
}

func ReqWithAuthHeader(userID uuid.UUID,
	privateKey *rsa.PrivateKey, claimsOptions ...jwt.TokenOption) func(*http.Request) {
	return ReqWithHeader("Authorization", MakeAuthHeader(userID, privateKey, claimsOptions...))
}
