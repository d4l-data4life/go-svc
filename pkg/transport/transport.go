package transport

import "net/http"

type TransportFunc func(http.RoundTripper) http.RoundTripper
