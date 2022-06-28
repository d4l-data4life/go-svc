package tut

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type ValueCheckFunc func(value interface{}) error

// ValueMatchesRegex allows to check that a string value matches a given regexp.
// It will panic if the regexp cannot be compiled.
func ValueMatchesRegex(wantRegex string) ValueCheckFunc {
	return func(value interface{}) error {
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %v", value)
		}
		if !regexp.MustCompile(wantRegex).MatchString(strVal) {
			return fmt.Errorf("want value matching: '%v', have: '%v'", wantRegex, strVal)
		}

		return nil
	}
}

// ValueEquals allows to check the strict equality of a value with an expected value.
// The equality is checked using reflect.DeepEqual
func ValueEquals(wantValue interface{}) ValueCheckFunc {
	return func(value interface{}) error {
		if !reflect.DeepEqual(wantValue, value) {
			return fmt.Errorf("want value: '%v', have: '%v'", wantValue, value)
		}

		return nil
	}
}

// ValueNotContains checks if a string value does not contain an expected value
func ValueNotContains(wantValue string) ValueCheckFunc {
	return func(value interface{}) error {
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %v", value)
		}

		if strings.Contains(strVal, wantValue) {
			return fmt.Errorf("expected %s to not contain %s", strVal, wantValue)
		}

		return nil
	}
}

// ValuePassesCheck allows to checks that a value passes
// some custom checks.
func ValuePassesCheck(check func(interface{}) error) ValueCheckFunc {
	return func(value interface{}) error {
		if err := check(value); err != nil {
			return fmt.Errorf("value didn't pass the check: have %v, err: '%v'", value, err)
		}
		return nil
	}
}

// ValueHasPrefix checks if a string value starts with the given prefix
func ValueHasPrefix(prefix string) ValueCheckFunc {
	return func(value interface{}) error {
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %v", value)
		}

		if !strings.HasPrefix(strVal, prefix) {
			return fmt.Errorf("expected %s to have prefix %s", strVal, prefix)
		}

		return nil
	}
}

// URLHasQueryParameter checks if a URL string includes the given query parameter.
// It optionally allows to perform extra checks on the value of the query parameter.
// It uses (url.Values).Get to get the value associated with the key, so only the first
// matching value is checked. For multiple occurrences of the same key use URLHasQueryParameterMultipleValues
func URLHasQueryParameter(parameter string, valueChecks ...ValueCheckFunc) ValueCheckFunc {
	return func(value interface{}) error {
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %v", value)
		}

		u, err := url.Parse(strVal)
		if err != nil {
			return fmt.Errorf("got error when parsing URL '%v': %w", strVal, err)
		}
		q := u.Query()
		paramVal := q.Get(parameter)

		if paramVal == "" {
			return fmt.Errorf("expected to find query parameter '%s', found '%v'", parameter, paramVal)
		}

		for _, check := range valueChecks {
			if err := check(paramVal); err != nil {
				return fmt.Errorf("checks for query parameter '%s' failed: %v", parameter, err)
			}
		}
		return nil
	}
}

// URLHasQueryParameterMultipleValues checks if a URL string includes the given query parameter.
// It optionally allows to perform extra checks on the values of the query parameter.
// It uses the query map directly, so this can be used to check keys associated with multiple values.
// The compared result is always of type slice.
func URLHasQueryParameterMultipleValues(wantKey string, valueChecks ...ValueCheckFunc) ValueCheckFunc {
	return func(value interface{}) error {
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %v", value)
		}

		u, err := url.Parse(strVal)
		if err != nil {
			return fmt.Errorf("got error when parsing URL '%v': %w", strVal, err)
		}

		haveValues := u.Query()[wantKey]
		if len(haveValues) == 0 {
			return fmt.Errorf("expected query param key '%s' to have values; got none", wantKey)
		}

		for _, check := range valueChecks {
			if err := check(haveValues); err != nil {
				return fmt.Errorf("checks for query parameter '%s' failed: %v", wantKey, err)
			}
		}
		return nil
	}
}

// URLHasNoQueryParameter checks if the URL value does not include a given parameter
func URLHasNoQueryParameter(key string) ValueCheckFunc {
	return func(value interface{}) error {
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %v", value)
		}

		u, err := url.Parse(strVal)
		if err != nil {
			return fmt.Errorf("got error when parsing redirect location: %w", err)
		}
		q := u.Query()
		presentValue := q.Get(key)
		if presentValue != "" {
			return fmt.Errorf("expected parameter \"%s\" to not be set, got \"%s\"", key, presentValue)
		}
		return nil
	}
}

// ValueIsParsableToInt checks that a string value represents a positive integer.
// It optionally allows to perform more checks on the parsed int value
func ValueIsParsableToInt(valueChecks ...func(int) error) ValueCheckFunc {
	return func(v interface{}) error {
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("expected value to be a string, got %v", v)
		}

		i, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("expected value to be parseable to integer, got %v", err)
		}

		for _, check := range valueChecks {
			if err := check(i); err != nil {
				return fmt.Errorf("int value %v didn't pass the checks: %v", i, err)
			}
		}
		return nil
	}
}

// IsPositiveInt checks that an integer value is strictly positive
func IsPositiveInt(i int) error {
	if i <= 0 {
		return fmt.Errorf("expected positive int, got %v", i)
	}

	return nil
}
