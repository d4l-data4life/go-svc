package middlewares

import (
	"errors"
	"net/http"

	"github.com/d4l-data4life/go-svc/pkg/logging"
)

var (
	ErrInvalidQuery = errors.New("query contains invalid parameters or values")
)

// stringContainsCTLByte reports whether s contains any ASCII control character.
func stringContainsCTLByte(s string) bool {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b < ' ' || b == 0x7f {
			return true
		}
	}
	return false
}

func stringsContainCTLByte(strings []string) bool {
	for _, s := range strings {
		if stringContainsCTLByte(s) {
			return true
		}
	}
	return false
}

// URLValidator middleware validates the query params to contain no control characters
func URLValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryValues := r.URL.Query()

		// validate query values
		for param, values := range queryValues {
			if stringContainsCTLByte(param) || stringsContainCTLByte(values) {
				logging.LogErrorfCtx(r.Context(), ErrInvalidQuery, "")
				WriteHTTPErrorCode(w, ErrInvalidQuery, http.StatusBadRequest)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
