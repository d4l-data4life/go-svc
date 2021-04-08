package jwt

import "net/http"

func httpClientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
