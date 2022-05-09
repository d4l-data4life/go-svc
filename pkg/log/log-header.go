package log

import (
	"fmt"
	"net/http"
	"net/textproto"
	"strings"
)

var headerSetCookie string = textproto.CanonicalMIMEHeaderKey("Set-Cookie")
var headerCookie string = textproto.CanonicalMIMEHeaderKey("Cookie")

type headerObfuscator struct {
	obfuscateHeader map[string]bool
	ignoreHeader    map[string]bool
}

func newHeaderObfuscator() *headerObfuscator {
	return &headerObfuscator{
		obfuscateHeader: make(map[string]bool),
		ignoreHeader:    make(map[string]bool),
	}
}

// Add header keys that should be obfuscated by the value of its length. An already ignored header can not be obfuscated. `Set-Cookie` and `Cookie` is handled by default
func (heob *headerObfuscator) obfuscateHeaders(keys []string) *headerObfuscator {
	for _, key := range keys {
		if canon := textproto.CanonicalMIMEHeaderKey(key); canon != headerSetCookie && canon != headerCookie && !heob.ignoreHeader[canon] {
			heob.obfuscateHeader[canon] = true
		}
	}
	return heob
}

// Add header keys that should be ignored. An already obfuscated header can not be ignored. `Set-Cookie` and `Cookie` is handled by default
func (heob *headerObfuscator) ignoreHeaders(keys []string) *headerObfuscator {
	for _, key := range keys {
		if canon := textproto.CanonicalMIMEHeaderKey(key); canon != headerSetCookie && canon != headerCookie && !heob.obfuscateHeader[canon] {
			heob.ignoreHeader[canon] = true
		}
	}
	return heob
}

// ProcessHeaders either obfuscates, ignores or does nothing to headers. Obfuscation happens by replacing the header value by its length. The header `Cookie` and `Set-Cookie` values' are always obfuscated
func (heob *headerObfuscator) processHeaders(header http.Header) http.Header {

	// we do not want to interfer with the existing request header
	processedHeaders := header.Clone()

	// Ignoring headers
	for key := range heob.ignoreHeader {
		processedHeaders.Del(key)
	}

	// Obfuscation of headers
	for key := range heob.obfuscateHeader {
		value := processedHeaders.Values(key)
		for i, singularValue := range value {
			value[i] = fmt.Sprintf("Obfuscated{%d}", len(singularValue))
		}
	}

	// modify value's in Set-cookie
	setCookie := processedHeaders.Values(headerSetCookie)
	for i, value := range setCookie {
		// according to https://datatracker.ietf.org/doc/html/rfc6265#section-4.1.1, `Set-Cookie: key=value` is the minimal format
		parts := strings.SplitN(value, "=", 2)

		// malformed header, clear value when in doubt
		if len(parts) != 2 {
			setCookie[i] = fmt.Sprintf("Invalid{%d}", len(value))
			continue
		}

		// the Set-Cookie can have different parameters like domain which are seperated by;
		semicolonSign := strings.Index(parts[1], ";")
		if semicolonSign > 0 {
			newValue := parts[1][:semicolonSign]
			// rebuild the cookie
			setCookie[i] = fmt.Sprintf("%s=Obfuscated{%d}%s", parts[0], len(newValue), parts[1][semicolonSign:])
		} else {
			setCookie[i] = fmt.Sprintf("%s=Obfuscated{%d}", parts[0], len(parts[1]))
		}
	}

	// parsing cookie values via a fake request
	if header.Get(headerCookie) != "" {

		cookieString := strings.Builder{}
		request := http.Request{Header: header}

		for _, cookie := range request.Cookies() {
			// be RFC compliant with the white space (https://datatracker.ietf.org/doc/html/rfc6265#section-4.2)
			cookieString.WriteString(fmt.Sprintf("%s=Obfuscated{%d}; ", cookie.Name, len(cookie.Value)))
		}

		// remove last whitespace
		processedHeaders[headerCookie] = []string{cookieString.String()[:cookieString.Len()-1]}
	}

	return processedHeaders
}
