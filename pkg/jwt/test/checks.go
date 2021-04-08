package test

import (
	"fmt"
	"net/http/httptest"
)

type CheckFunc func(w *httptest.ResponseRecorder) error

func Checks(fns ...CheckFunc) []CheckFunc { return fns }

func HasStatusCode(want int) CheckFunc {
	return func(r *httptest.ResponseRecorder) error {
		if have := r.Code; want != have {
			return fmt.Errorf("\nwant: %d\nhave: %d", want, have)
		}

		return nil
	}
}
