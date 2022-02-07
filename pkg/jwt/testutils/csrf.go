package testutils

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/justinas/nosurf"
)

// CSRFValues returns a cookie that is used as a store for the service to evaluate that the
// CSRF header given by the client is valid. The string returned is the CSRF header value to
// be set alongside the cookie during requests to protected endpoints.
func CSRFValues() (*http.Cookie, string) {
	handler := nosurf.New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(nosurf.HeaderName, nosurf.Token(r))
	}))

	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		panic(err)
	}

	handler.ServeHTTP(w, r)

	res := w.Result()
	if len(res.Cookies()) != 1 {
		panic(fmt.Sprintf("expected one cookie, received: %v", res.Cookies()))
	}
	res.Body.Close()

	cookie := res.Cookies()[0]
	if cookie.Name != nosurf.CookieName {
		panic(fmt.Sprintf("expected cookie name %s, but have %s", nosurf.CookieName, cookie.Name))
	}

	token := res.Header.Get(nosurf.HeaderName)
	if token == "" {
		panic(fmt.Sprintf("expected header %s, but have headers: %v", nosurf.HeaderName, res.Header))
	}

	return cookie, token
}
