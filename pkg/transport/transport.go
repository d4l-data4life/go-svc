package transport

import "net/http"

// nolint: revive
type TransportFunc func(http.RoundTripper) http.RoundTripper
