package tut

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/go-chi/chi"
)

type RequestCheckFunc func(*http.Request) error

// CheckRequest allows to aggregate multiple request check functions
func CheckRequest(checks ...func(*http.Request) error) RequestCheckFunc {
	return func(res *http.Request) error {
		for _, check := range checks {
			if err := check(res); err != nil {
				return err
			}
		}

		return nil
	}
}

// ReqHasHTTPMethod returns a function that checks that a request has the expected HTTP method
func ReqHasHTTPMethod(wantMethod string) RequestCheckFunc {
	return func(r *http.Request) error {
		if have := r.Method; wantMethod != have {
			return fmt.Errorf("want %s, have %s", wantMethod, have)
		}

		return nil
	}
}

// ReqHasHeader returns a function that checks that a request has the expected header key
// and header value.
func ReqHasHeader(wantKey string, valueChecks ...ValueCheckFunc) RequestCheckFunc {
	return func(r *http.Request) error {
		have := r.Header.Get(wantKey)
		if have == "" {
			return fmt.Errorf("expected header key '%s' to be set", wantKey)
		}

		for _, check := range valueChecks {
			if err := check(have); err != nil {
				return err
			}
		}
		return nil
	}
}

// ReqURLChecks returns a function that checks the URL of a request.
func ReqURLChecks(checks ...ValueCheckFunc) RequestCheckFunc {
	return func(r *http.Request) error {
		for _, check := range checks {
			if err := check(r.URL.String()); err != nil {
				return fmt.Errorf("URL '%s' doesn't pass the checks: %v", r.URL, err)
			}
		}

		return nil
	}
}

func ReqHasURLParam(wantKey string, valueChecks ...ValueCheckFunc) RequestCheckFunc {
	return func(r *http.Request) error {
		param := chi.URLParam(r, wantKey)
		if param == "" {
			return fmt.Errorf("expected URL param key '%s' to be set", wantKey)
		}

		for _, check := range valueChecks {
			if err := check(param); err != nil {
				return fmt.Errorf("URL param '%s' doesn't pass the checks: %v", wantKey, err)
			}
		}

		return nil
	}
}

// ReqHasJSONBody allows to check that a request body contains the expected JSON value.
// wantValue must be a pointer to the expected body. The body will be tried to be unmarshalled
// in a variable of the same type. The values are compared then using reflect.DeepEqual
func ReqHasJSONBody(wantValue interface{}) RequestCheckFunc {
	rawType := reflect.TypeOf(wantValue)
	if rawType.Kind() != reflect.Ptr {
		panic(fmt.Errorf("ReqHasJSONBody expects a pointer as argument, got: %+v", wantValue))
	}

	// get the type behind the pointer
	rawType = rawType.Elem()
	// instantiate a new variable of the type to use as target for json.Decode
	decodePointer := reflect.New(rawType).Interface()
	return func(r *http.Request) error {
		if err := json.NewDecoder(r.Body).Decode(decodePointer); err != nil {
			return fmt.Errorf("error decoding the body: %w", err)
		}

		if !reflect.DeepEqual(wantValue, decodePointer) {
			return fmt.Errorf("expected body: %+v; got %+v", wantValue, decodePointer)
		}

		return nil
	}
}

// ReqHasInContext returns a function that checks that the context of the request contains
// the expected value associated to the expected key.
func ReqHasInContext(wantKey interface{}, wantValue interface{}) RequestCheckFunc {
	return func(r *http.Request) error {
		haveValue := r.Context().Value(wantKey)
		if !reflect.DeepEqual(wantValue, haveValue) {
			return fmt.Errorf("unexpected context value for key %v; want: %v, have: %v", wantKey, wantValue, haveValue)
		}

		return nil
	}
}

// ReqHasInContextExtract returns a function that allows to extract some value from the context of the request
// and compare it with the expected value.
func ReqHasInContextExtract(extract func(context.Context) (interface{}, error), wantValue interface{}) RequestCheckFunc {
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
