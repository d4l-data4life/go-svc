package jwt

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	jwtReq "github.com/golang-jwt/jwt/v4/request"
	"github.com/justinas/nosurf"
)

const (
	AccessCookieName        string = "phdp-access-token"
	AccessTokenArgumentName string = "access_token"
)

// cookieExtractor extracts an access token from a cookie.
// Implements the tokenExtractor interface.
type cookieExtractor struct{}

func newCookieExtractor() cookieExtractor {
	return cookieExtractor{}
}

// ExtractToken implements the tokenExtractor interface.
// It attempts to extract the token from the access cookie.
// If the cookie is present and the method is not a CSRF-safe method,
// it also checks if the CSRF protection criteria are met.
// If yes, the found token is considered valid.
// If not, the token is ignored.
func (e cookieExtractor) ExtractToken(req *http.Request) (string, error) {
	tc, err := req.Cookie(AccessCookieName)
	if err != nil || tc.Value == "" {
		return "", fmt.Errorf("trying to extract token from cookie: no access cookie found")
	}

	// in case an access token was found, it will be considered only if the CSRF
	// protection is also present.
	if err = checkCSRF(req); err != nil {
		return "", fmt.Errorf("trying to extract token from cookie: access cookie found but with CSRF error: %w", err)
	}

	return tc.Value, nil
}

// checkCSRF checks if the request passes the double-submit-cookie CSRF protection.
// For CSRF-safe methods ("GET", "HEAD", "OPTIONS", "TRACE") this check will always pass.
func checkCSRF(req *http.Request) error {
	// unfortunately nosurf doesn't expose an API for directly validating a request.
	// Therefore we need to use the middleware they expose and get the error from
	// the failure handler.
	var csrfErr error
	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	failureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		csrfErr = nosurf.Reason(r)
	})
	csrfHandler := nosurf.New(successHandler)
	csrfHandler.SetFailureHandler(failureHandler)

	// call the middleware with a mock response writer
	csrfHandler.ServeHTTP(httptest.NewRecorder(), req)

	if csrfErr != nil {
		return fmt.Errorf("CSRF validation failed: %w", csrfErr)
	}

	return nil
}

// headerExtractor extracts an access token from the authorization header.
// Implements the tokenExtractor interface.
type headerExtractor struct {
	extractor jwtReq.Extractor
}

func newHeaderExtractor() headerExtractor {
	return headerExtractor{
		extractor: jwtReq.AuthorizationHeaderExtractor,
	}
}

// ExtractToken implements the tokenExtractor interface.
func (e headerExtractor) ExtractToken(req *http.Request) (string, error) {
	// delegate the actual extract to the golang-jwt/jwt/v4/request package.
	// this wrapper just offers better logging of errors.
	t, err := e.extractor.ExtractToken(req)
	if err != nil {
		return "", fmt.Errorf("trying to extract token from header: %w", err)
	}

	return t, nil
}

// argumentExtractor extracts a token from request arguments. This
// includes a POSTed form or GET URL arguments.
// Implements the tokenExtractor interface.
type argumentExtractor struct {
	extractor jwtReq.Extractor
}

func newArgumentExtractor() argumentExtractor {
	return argumentExtractor{
		extractor: jwtReq.ArgumentExtractor{AccessTokenArgumentName},
	}
}

// ExtractToken implements the tokenExtractor interface.
func (e argumentExtractor) ExtractToken(req *http.Request) (string, error) {
	// delegate the actual extract to the golang-jwt/jwt/v4/request package
	// but wrap it with better error logs
	t, err := e.extractor.ExtractToken(req)
	if err != nil {
		return "", fmt.Errorf("trying to extract token from argument '%v': %w", AccessTokenArgumentName, err)
	}

	return t, nil
}
