package jwt_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
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

// hasInContextExtract returns a function that allows to extract some value from the context of the request
// and compare it with the expected value.
func hasInContextExtract(extract func(context.Context) (interface{}, error), wantValue interface{}) checkReqFunc {
	return func(r *http.Request) error {
		haveValue, err := extract(r.Context())
		if err != nil {
			return fmt.Errorf("can't extract from context: %w", err)
		}
		if !reflect.DeepEqual(wantValue, haveValue) {
			return fmt.Errorf("unexpected context value found in context; want: %v, have: %v", wantValue, haveValue)
		}

		return nil
	}
}
