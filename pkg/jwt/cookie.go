package jwt

import (
	"net/http"

	jwtReq "github.com/golang-jwt/jwt/v4/request"
)

const AccessCookieName string = "phdp-access-token"

// cookieExtractor extracts an access token from a cookie.
// Implements the golang-jwt/jwt/v4/request.Extractor interface.
type cookieExtractor struct {
	cookieName string
}

func newCookieExtractor() cookieExtractor {
	return cookieExtractor{
		cookieName: AccessCookieName,
	}
}

// ExtractToken implements the golang-jwt/jwt/v4/request.Extractor interface.
// If no token is present, it must return ErrNoTokenInRequest.
func (e cookieExtractor) ExtractToken(req *http.Request) (string, error) {
	tc, err := req.Cookie(e.cookieName)
	if err != nil {
		// golang-jwt/jwt/v4/request doesn't support yet wrapped errors using fmt.Errorf
		// therefore this exact error needs to be returned
		return "", jwtReq.ErrNoTokenInRequest
	}

	if tc.Value == "" {
		// if the value is empty, consider it a missing token
		return "", jwtReq.ErrNoTokenInRequest
	}

	return tc.Value, nil
}
