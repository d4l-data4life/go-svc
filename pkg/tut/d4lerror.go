package tut

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gesundheitscloud/go-svc/pkg/d4lerror"
)

type ErrorResponseCheckFunc func(d4lerror.ErrorItem) error

// BodyContainsErrorV2 checks if the body is a well-formatted errors v2 response body
// and that the expected error is present in the response. It also tests that there
// is only one error returned.
func BodyContainsErrorV2(checks ...ErrorResponseCheckFunc) ResponseCheckFunc {
	return func(res *http.Response) error {
		var body d4lerror.ErrorResponse
		if err := json.NewDecoder(MultiReadResponseBody(res)).Decode(&body); err != nil {
			return fmt.Errorf("could not parse body: %w", err)
		}

		if len(body.Errors) != 1 {
			return fmt.Errorf("expected one error, have %v", body.Errors)
		}

		for _, check := range checks {
			if err := check(body.Errors[0]); err != nil {
				return err
			}
		}

		return nil
	}
}

// HasErrorCode checks that the error response item contains the given code
func HasErrorCode(want d4lerror.ErrorCode) ErrorResponseCheckFunc {
	return func(e d4lerror.ErrorItem) error {
		if e.Code != want {
			return fmt.Errorf("expected to find error code %v, have %v", want, e.Code)
		}
		return nil
	}
}

// HasErrorMessage checks that the error response item contains the given message
func HasErrorMessage(want string) ErrorResponseCheckFunc {
	return func(e d4lerror.ErrorItem) error {
		if e.Message != want {
			return fmt.Errorf("expected to find error message '%v', have '%v'", want, e.Message)
		}
		return nil
	}
}

// HasDetails checks if the decoded value of Details matches the expected values.
// Note that since Details is defined as an interface{} in the struct, it will be decoded
// as a map[string]interface{}. So the expected value can't be anything else.
// Also the values of the map need to match the default JSON types when unmarshalling an interface
// see https://golang.org/pkg/encoding/json/#Unmarshal
func HasDetails(want interface{}) ErrorResponseCheckFunc {
	return func(e d4lerror.ErrorItem) error {
		if !reflect.DeepEqual(e.Details, want) {
			return fmt.Errorf("expected to find details %v, have %v", want, e.Details)
		}
		return nil
	}
}

// HasDetailsKey allows to checks that the details returned with an error v2 contains an expected key.
// It also allows to perform more checks for the value associated with the expected key.
func HasDetailsKey(wantKey string, valueChecks ...ValueCheckFunc) ErrorResponseCheckFunc {
	return func(e d4lerror.ErrorItem) error {
		details, ok := e.Details.(map[string]interface{})
		if !ok {
			return errors.New("details are not of the expected type")
		}

		for _, check := range valueChecks {
			if err := check(details[wantKey]); err != nil {
				return fmt.Errorf("details value for key '%s' doesn't pass the checks: %v", wantKey, err)
			}
		}
		return nil
	}
}

// HasTraceID checks that the error response item contains the given trace ID
func HasTraceID(want string) ErrorResponseCheckFunc {
	return func(e d4lerror.ErrorItem) error {
		if e.TraceID != want {
			return fmt.Errorf("expected to find trace ID %v, have %v", want, e.TraceID)
		}
		return nil
	}
}
