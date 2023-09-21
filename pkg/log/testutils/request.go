/*
This file contains helper tools for creating a request for test purposes.
*/
package testutils

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Request creates a new request
// and applies the given request options to it.
func Request(opts ...func(*http.Request)) *http.Request {
	req, err := http.NewRequest("", "", nil)
	if err != nil {
		panic(err)
	}

	for _, opt := range opts {
		opt(req)
	}

	return req
}

// WithContext returns a function that modifies the given request
// by setting the given context on the request object.
func WithContext(ctx context.Context) func(*http.Request) {
	return func(req *http.Request) {
		*req = *req.WithContext(ctx)
	}
}

// WithContextValue returns a function that modifies the given request
// by setting the given context on the request object.
func WithContextValue(key, value interface{}) func(*http.Request) {
	return func(req *http.Request) {
		*req = *(req.WithContext(context.WithValue(req.Context(), key, value)))
	}
}

// WithJSONBody returns a function that modifies the given request
// by setting the marshaling the given interface{} to the request body.
func WithJSONBody(v interface{}) func(*http.Request) {
	return func(req *http.Request) {
		jsonValue, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}

		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Body = io.NopCloser(strings.NewReader(string(jsonValue)))
	}
}

// WithTargetURL returns a function that modifies the given request
// by setting the given target URL.
// If parsing failed, it panics.
func WithTargetURL(target string) func(*http.Request) {
	return func(req *http.Request) {
		url, err := url.Parse(target)
		if err != nil {
			panic(err)
		}
		req.URL = url
	}
}

// WithMethod returns a function that modifies the given request
// by setting its HTTP method.
func WithMethod(m string) func(*http.Request) {
	return func(req *http.Request) {
		req.Method = m
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
		req.Body = io.NopCloser(strings.NewReader(data.Encode()))
	}
}

// WithCookies returns a function that modifies the given request
// by setting the given cookie
func WithCookies(cookies ...*http.Cookie) func(*http.Request) {
	return func(req *http.Request) {
		for _, c := range cookies {
			req.AddCookie(c)
		}
	}
}

// WithHeader returns a function that modifies the given request
// by setting the given header.
// nolint-unparam - will be useful in the future and therefore not fixed
func WithHeader(key, value string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

// WithValue returns a function that modifies the url values by adding a key-value pair
func WithValue(key, value string) func(*url.Values) {
	return func(data *url.Values) {
		data.Add(key, value)
	}
}

// WithQuery adds query parameters to the url.
func WithQuery(values url.Values) func(*http.Request) {
	return func(req *http.Request) {
		req.URL.RawQuery = values.Encode()
	}
}

// WithAppendToHeader adds header key / values without overwriting values.
func WithAppendToHeader(key, value string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Add(key, value)
	}
}
