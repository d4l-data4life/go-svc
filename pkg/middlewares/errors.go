package middlewares

import (
	"net/http"
)

// WriteHTTPErrorCode writes given HTTP Code to the HTTP response and provides explanation - for errors
func WriteHTTPErrorCode(w http.ResponseWriter, err error, code int) {
	http.Error(w, err.Error(), code)
}
