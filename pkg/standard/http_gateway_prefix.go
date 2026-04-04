package standard

import (
	"net/http"
	"net/url"
	"strings"
)

// WithPathPrefixStrip returns a handler that strips prefix from request URLs when they match
// the given path prefix (exact match or prefix + "/"). If prefix is empty or doesn't match,
// the original request is passed through unchanged.
func WithPathPrefixStrip(prefix string, h http.Handler) http.Handler {
	if prefix == "" {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path != prefix && !strings.HasPrefix(path, prefix+"/") {
			h.ServeHTTP(w, r)
			return
		}

		newPath := strings.TrimPrefix(path, prefix)
		if newPath == "" {
			newPath = "/"
		}

		u2 := new(url.URL)
		*u2 = *r.URL
		u2.Path = newPath
		u2.RawPath = ""

		r2 := r.Clone(r.Context())
		r2.URL = u2
		h.ServeHTTP(w, r2)
	})
}
