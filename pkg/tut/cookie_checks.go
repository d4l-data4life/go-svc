package tut

import (
	"fmt"
	"net/http"
)

type CookieCheckFunc func(cookie *http.Cookie) error

// CookieValueCheck runs checks on the cookie's value.
func CookieValueCheck(checks ...ValueCheckFunc) CookieCheckFunc {
	return func(cookie *http.Cookie) error {
		for _, check := range checks {
			if err := check(cookie.Value); err != nil {
				return err
			}
		}

		return nil
	}
}

// CookieHasSecure checks if the cookie's Secure attribute is the given value.
func CookieHasSecure(secure bool) CookieCheckFunc {
	return func(cookie *http.Cookie) error {
		if cookie.Secure != secure {
			return fmt.Errorf("cookie is %t secure but expected %t", cookie.Secure, secure)
		}

		return nil
	}
}

// CookieHasHttpOnly checks if the cookie's HttpOnly attribute is the given value.
func CookieHasHttpOnly(httpOnly bool) CookieCheckFunc {
	return func(cookie *http.Cookie) error {
		if cookie.HttpOnly != httpOnly {
			return fmt.Errorf("cookie is %t httpOnly but expected %t", cookie.HttpOnly, httpOnly)
		}

		return nil
	}
}
