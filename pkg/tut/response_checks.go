package tut

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

type ResponseCheckFunc func(*http.Response) error

// CheckResponse runs multiple responseChecks sequentially. It stops and returns at the first encountered error.
func CheckResponse(checks ...func(*http.Response) error) func(*http.Response) error {
	return func(res *http.Response) error {
		for _, check := range checks {
			if err := check(res); err != nil {
				return err
			}
		}

		return nil
	}
}

// RespBodyIsValidJSON checks if the body is valid JSON
func RespBodyIsValidJSON(res *http.Response) error {
	bodyReader := MultiReadResponseBody(res)
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return fmt.Errorf("could not read the body")
	}
	if !json.Valid(body) {
		return fmt.Errorf("body is not valid JSON: `%s`", body)
	}
	return nil
}

// RespHasContentType checks that the response contains the given value as content-type header.
// This is a convenience shortcut for checking the value of the "Content-Type" header
func RespHasContentType(wantValue string) ResponseCheckFunc {
	return RespHasHeader(ContentTypeHeader, ValueEquals(wantValue))
}

// RespHasNoCookie checks that a given cookie name is not set in a response
func RespHasNoCookie(name string) ResponseCheckFunc {
	return func(res *http.Response) error {
		for _, cookie := range res.Cookies() {
			if cookie.Name == name {
				return fmt.Errorf("expected cookie %v to not be set; got %v", name, cookie)
			}
		}

		return nil
	}
}

// RespHasSetCookie checks that the response contains the given cookie
// It optionally allows to pass extra checks for the cookie.
func RespHasSetCookie(wantName string, cookieChecks ...CookieCheckFunc) ResponseCheckFunc {
	return func(res *http.Response) error {
		for _, cookie := range res.Cookies() {
			if cookie.Name == wantName {
				for _, check := range cookieChecks {
					if err := check(cookie); err != nil {
						return fmt.Errorf("checks for cookie '%s' failed: %v", wantName, err)
					}
				}
				return nil
			}
		}

		return fmt.Errorf(
			"expected to find cookie '%s', have cookies: '%s'", wantName, res.Cookies(),
		)
	}
}

// RespHasStatusCode checks that the response contains the expected HTTP status code
func RespHasStatusCode(want int) ResponseCheckFunc {
	return func(res *http.Response) error {
		if have := res.StatusCode; have != want {
			return fmt.Errorf("want status code %d, have %d", want, have)
		}
		return nil
	}
}

// RespHasTextBody checks if the response text body passes all passed value checks.
// It ignores whitespaces from the response body.
func RespHasTextBody(checks ...ValueCheckFunc) ResponseCheckFunc {
	return func(res *http.Response) error {
		responseData, err := ioutil.ReadAll(MultiReadResponseBody(res))
		if err != nil {
			return fmt.Errorf("could not read body: %w", err)
		}

		have := strings.TrimSpace(string(responseData))

		for _, check := range checks {
			if err := check(have); err != nil {
				return fmt.Errorf("checks for body '%s' failed: %v", have, err)
			}
		}

		return nil
	}
}

// RespHasJSONBody allows to check that a response body contains the expected JSON value.
// wantValue must be a pointer to the expected body. The body will be tried to be unmarshalled
// in a variable of the same type. The values are compared then using reflect.DeepEqual
func RespHasJSONBody(wantValue interface{}) ResponseCheckFunc {
	rawType := reflect.TypeOf(wantValue)
	if rawType.Kind() != reflect.Ptr {
		panic(fmt.Errorf("RespHasJSONBody expects a pointer as argument, got: %+v", wantValue))
	}

	// get the type behind the pointer
	rawType = rawType.Elem()
	// instantiate a new variable of the type to use as target for json.Decode
	decodePointer := reflect.New(rawType).Interface()
	return func(r *http.Response) error {
		if err := json.NewDecoder(MultiReadResponseBody(r)).Decode(decodePointer); err != nil {
			return fmt.Errorf("error decoding the body: %w", err)
		}

		if !reflect.DeepEqual(wantValue, decodePointer) {
			return fmt.Errorf("expected body: %+v; got %+v", wantValue, decodePointer)
		}

		return nil
	}
}

// RespBodyCanBeDecodedToType checks if the body can be unmarshalled to the given struct.
// If possible, use the stronger check RespHasJSONBody
func RespBodyCanBeDecodedToType(target interface{}) ResponseCheckFunc {
	return func(r *http.Response) error {
		if err := json.NewDecoder(MultiReadResponseBody(r)).Decode(target); err != nil {
			return fmt.Errorf("could not decode response body to the expected type: %w", err)
		}
		return nil
	}
}

// RespBodyHasNoKey checks if the body is JSON and the given key is not present
func RespBodyHasNoKey(key string) ResponseCheckFunc {
	return func(r *http.Response) error {
		var body map[string]json.RawMessage
		if err := json.NewDecoder(MultiReadResponseBody(r)).Decode(&body); err != nil {
			return err
		}
		if _, ok := body[key]; ok {
			return fmt.Errorf("expected JSON body not to have key %q, but got some", key)
		}
		return nil
	}
}

// RespBodyHasKey checks if the body is JSON and the given key is present.
// It allows to optionally perform more checks for the corresponding value.
func RespBodyHasKey(key string, valueChecks ...ValueCheckFunc) ResponseCheckFunc {
	return func(r *http.Response) error {
		var body map[string]interface{}
		if err := json.NewDecoder(MultiReadResponseBody(r)).Decode(&body); err != nil {
			return err
		}
		value, ok := body[key]
		if !ok {
			return fmt.Errorf("expected JSON body has key %q, got none", key)
		}

		for _, check := range valueChecks {
			if err := check(value); err != nil {
				return fmt.Errorf("checks for body key '%s' failed: %v", key, err)
			}
		}

		return nil
	}
}

// MultiReadResponseBody is a helper function for reading the body of
// a *http.Request multiple times.
func MultiReadResponseBody(res *http.Response) *bytes.Reader {
	body, _ := ioutil.ReadAll(res.Body)
	if err := res.Body.Close(); err != nil {
		panic(err)
	}
	newBody := make([]byte, len(body))
	copy(newBody, body)
	res.Body = ioutil.NopCloser(bytes.NewBuffer(newBody))
	return bytes.NewReader(body)
}

// RespHasHeader checks that a header is present on a response
func RespHasHeader(wantHeaderName string, valueChecks ...ValueCheckFunc) ResponseCheckFunc {
	return func(res *http.Response) error {
		val := res.Header.Get(wantHeaderName)
		if val == "" {
			return fmt.Errorf("want response header %s", wantHeaderName)
		}

		for _, check := range valueChecks {
			if err := check(val); err != nil {
				return err
			}
		}
		return nil
	}
}

// RespHasNoHeader checks that a header is not present on a response
func RespHasNoHeader(headerName string) ResponseCheckFunc {
	return func(res *http.Response) error {
		val := res.Header.Get(headerName)
		if val != "" {
			return fmt.Errorf("expected header %v to not be set; got %v", headerName, val)
		}

		return nil
	}
}
