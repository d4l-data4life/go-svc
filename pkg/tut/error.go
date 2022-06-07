package tut

import (
	"errors"
	"fmt"
)

type ErrorCheckFunc func(e error) error

// HasNoError checks that the given error is nil
func HasNoError() ErrorCheckFunc {
	return func(e error) error {
		if e != nil {
			return fmt.Errorf("expected no error, got '%v'", e)
		}
		return nil
	}
}

// HasError returns a function that can check that an error matches the expectations.
// The expected error can be nil.
func HasError(expected error) ErrorCheckFunc {
	return func(got error) error {
		if !errors.Is(got, expected) {
			return fmt.Errorf("expected error: `%v`, have: `%v`", expected, got)
		}
		return nil
	}
}

// HasAnyError allows to check that an error was returned.
func HasAnyError(got error) error {
	if got == nil {
		return fmt.Errorf("expected some error; have: `%v`", got)
	}
	return nil
}
