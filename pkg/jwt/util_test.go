package jwt

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"time"
)

type testData struct {
	name       string
	request    *http.Request
	middleware func(http.Handler) http.Handler
	checks     []checkFunc
	endHandler func(http.ResponseWriter, *http.Request)
}

////////////////////////////////////////////////////////////////////////////////
// Check Funcs
////////////////////////////////////////////////////////////////////////////////

type checkFunc func(w *httptest.ResponseRecorder) error

func checks(fns ...checkFunc) []checkFunc { return fns }

func hasStatusCode(want int) checkFunc {
	return func(r *httptest.ResponseRecorder) error {
		if have := r.Code; want != have {
			return fmt.Errorf("\nwant: %d\nhave: %d", want, have)
		}

		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// Check Request Funcs
////////////////////////////////////////////////////////////////////////////////

type checkReqFunc func(w *http.Request) error

func checkReqAll(fns ...checkReqFunc) checkReqFunc {
	return func(r *http.Request) error {
		for _, check := range fns {
			if err := check(r); err != nil {
				return err
			}
		}
		return nil
	}
}

// hasInContext returns a function that checks that the context of the request contains
// the expected value associated to the expected key.
func hasInContext(wantKey interface{}, wantValue interface{}) checkReqFunc {
	return func(r *http.Request) error {
		haveValue := r.Context().Value(wantKey)
		if !reflect.DeepEqual(wantValue, haveValue) {
			return fmt.Errorf("unexpected context value for key %v; want: %v, have: %v", wantKey, wantValue, haveValue)
		}

		return nil
	}
}

// hasKeyInContext returns a function that checks that the context of the request contains
// the expected key with a non-nil value.
func hasKeyInContext(wantKey interface{}) checkReqFunc {
	return func(r *http.Request) error {
		haveValue := r.Context().Value(wantKey)
		if haveValue == nil {
			return fmt.Errorf("expected to find key %v; have nil", wantKey)
		}

		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// HTTP Request Builder and OKHandler
////////////////////////////////////////////////////////////////////////////////

type requestBuilder func(*http.Request) error

func buildRequest(url string, fns ...requestBuilder) *http.Request {
	r := httptest.NewRequest("", url, nil)

	for _, fn := range fns {
		_ = fn(r)
	}

	return r
}

func withAuthHeader(key *rsa.PrivateKey, options ...TokenOption) requestBuilder {
	return func(r *http.Request) error {
		options = append(
			// add a default expiration time as the token is not valid without one
			[]TokenOption{WithExpirationTime(time.Now().Add(1 * time.Minute))},
			options...,
		)
		t, err := CreateAccessToken(key, options...)
		if err != nil {
			return err
		}

		r.Header.Add("Authorization", "Bearer "+t.AccessToken)

		return nil
	}
}

func withOwnerPath(owner string) string {
	var builder strings.Builder

	builder.WriteString("/users/")
	builder.WriteString(owner)
	builder.WriteString("/records/456")

	return builder.String()
}

func withOwnerURL(owner string) string {
	var builder strings.Builder

	builder.WriteString("http://test.data4life.care")
	builder.WriteString(withOwnerPath(owner))

	return builder.String()
}

func okHandler(w http.ResponseWriter, r *http.Request) {}

////////////////////////////////////////////////////////////////////////////////
// Test Logger
////////////////////////////////////////////////////////////////////////////////

type testLogger struct{}

func (testLogger) ErrUserAuth(ctx context.Context, err error) error {
	fmt.Println(err)
	return nil
}
func (testLogger) InfoGeneric(ctx context.Context, msg string) error {
	fmt.Println(msg)
	return nil
}
func (testLogger) ErrGeneric(ctx context.Context, err error) error {
	fmt.Println(err)
	return nil
}
