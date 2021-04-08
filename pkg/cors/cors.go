package cors // import "github.com/gesundheitscloud/go-cors"

import (
	"net/http"
	"strconv"
	"strings"
)

// WithDomains is an option for Wrap, that accepts a slice of domains that will
// be whitelisted for CORS.
func WithDomains(domains []string) func(*corsHandler) {
	return func(h *corsHandler) {
		h.allowedDomains = domains
	}
}

// WithDomainList is an option for Wrap, that accepts a comma-separated list of
// domains that will be whitelisted for CORS.
func WithDomainList(domainList string) func(*corsHandler) {
	domainList = strings.Replace(domainList, " ", "", -1)
	domains := strings.Split(domainList, ",")
	return WithDomains(domains)
}

// WithHeaderList is an option for Wrap, that accepts a comma-separated list of
// headers that will be allowed for CORS.
func WithHeaderList(headerList string) func(*corsHandler) {
	return func(h *corsHandler) {
		h.allowedHeaders = headerList
	}
}

// WithMethodList is an option for Wrap that accepts a comma-separated list of
// methods that will be allowed for CORS.
func WithMethodList(methodList string) func(*corsHandler) {
	return func(h *corsHandler) {
		h.allowedMethods = methodList
	}
}

type corsHandler struct {
	next http.Handler

	allowedDomains     []string
	allowedMethods     string
	allowedHeaders     string
	allowedCredentials bool
}

// Options are the options for CORS
type Options struct {
	AllowedDomains     []string
	AllowedMethods     string
	AllowedHeaders     string
	AllowedCredentials bool
}

// ServeHTTP makes corsHandler implement http.Handler. The method is designed
// to only add the CORS headers if the request passes the Origin check. The
// value of the response's "Access-Control-Allow-Origin" header will only
// reflect the request's Origin, or "*".
func (h corsHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	origin := req.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}

	for _, allowedDomain := range h.allowedDomains {
		if allowedDomain == "*" || origin == allowedDomain {
			rw.Header().Set("Access-Control-Allow-Origin", origin)
			rw.Header().Set("Access-Control-Allow-Methods", h.allowedMethods)
			rw.Header().Set("Access-Control-Allow-Headers", h.allowedHeaders)
			rw.Header().Set("Access-Control-Allow-Credentials", strconv.FormatBool(h.allowedCredentials))
			break
		}
	}

	if req.Method == http.MethodOptions {
		return
	}

	h.next.ServeHTTP(rw, req)
}

// Wrap returns the given handler, decorated with a middleware that adds CORS
// headers if the request's Origin header matches one of the allowed origins.
//
// Default values:
//
//   Access-Control-Allow-Origin: *
//   Access-Control-Allow-Methods: OPTIONS, HEAD, POST, GET, PUT, DELETE
//   Access-Control-Allow-Headers: Authorization
func Wrap(handler http.Handler, options ...func(*corsHandler)) http.Handler {
	decorated := corsHandler{
		next:           handler,
		allowedDomains: []string{"*"},
		allowedMethods: "OPTIONS, HEAD, POST, GET, PUT, DELETE",
		allowedHeaders: "Authorization",
	}

	for _, apply := range options {
		apply(&decorated)
	}

	return decorated
}

// New creates a new CORS handler
// Default values:
//
//  Access-Control-Allow-Origin: *
//  Access-Control-Allow-Methods: OPTIONS, HEAD, POST, GET, PUT, DELETE
//  Access-Control-Allow-Headers: Authorization
func New(options Options) corsHandler {
	if len(options.AllowedDomains) == 0 {
		options.AllowedDomains = []string{"*"}
	}
	if options.AllowedMethods == "" {
		options.AllowedMethods = "OPTIONS, HEAD, POST, GET, PUT, DELETE"
	}
	if options.AllowedHeaders == "" {
		options.AllowedHeaders = "Authorization"
	}
	return corsHandler{
		allowedDomains:     options.AllowedDomains,
		allowedMethods:     options.AllowedMethods,
		allowedHeaders:     options.AllowedHeaders,
		allowedCredentials: options.AllowedCredentials,
	}
}

// Adds mux middleware for corshandler.
func (h corsHandler) MiddleWare(handler http.Handler) http.Handler {
	h.next = handler
	return h
}

func HTTPMiddleware(options ...func(*corsHandler)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return Wrap(next, options...)
	}
}
