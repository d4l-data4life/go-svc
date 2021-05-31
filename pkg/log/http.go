package log

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/d4lcontext"
)

type contextKey string

const (
	// UserIDContextKey is the key to store the user ID in the context
	UserIDContextKey contextKey = "user-id"

	// TraceIDContextKey is the key to store the trace ID in the context
	TraceIDContextKey contextKey = "trace-id"

	// CallerIPContextKey is the key to store the caller IP in the context
	CallerIPContextKey contextKey = "caller-ip"

	// ClientIDContextKey is the key to store the client ID in the context
	ClientIDContextKey contextKey = "client-id"

	// TenantIDContextKey is the key to store the caller IP in the context
	TenantIDContextKey contextKey = "tenant-id"

	// RequestURLContextKey is the key to store the request URL in the context
	RequestURLContextKey contextKey = "req-url"

	// RequestDomainContextKey is the key to store the request domain in the context
	RequestDomainContextKey contextKey = "req-domain"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode    int
	contentLength int64
	body          *bytes.Buffer
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.contentLength += int64(n)
	rw.body.Write(b)
	return n, err
}

type HTTPLogger struct {
	next http.Handler

	log *Logger

	userParser     func(*http.Request) string
	clientParser   func(*http.Request) string
	ipParser       func(*http.Request) string
	tenantIDParser func(*http.Request) string

	obf map[string][]Obfuscator
	ipa []IPAnonymizer
}

func (h *HTTPLogger) obfuscatorKey(et EventType, reqMethod string) string {
	return strings.ToLower(string(et) + ":" + reqMethod)
}

// TraceIDHeaderKey is the key to store the trace ID in the header
const TraceIDHeaderKey = "Trace-Id"

func (l HTTPLogger) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if traceID := req.Header.Get(TraceIDHeaderKey); traceID != "" {
		req = req.WithContext(context.WithValue(req.Context(), TraceIDContextKey, traceID))
	}

	if userID := l.userParser(req); userID != "" {
		req = req.WithContext(context.WithValue(req.Context(), UserIDContextKey, userID))
	} else if userID = d4lcontext.GetUserID(req); userID != "" {
		req = req.WithContext(context.WithValue(req.Context(), UserIDContextKey, userID))
	}

	if clientID := l.clientParser(req); clientID != "" {
		req = req.WithContext(context.WithValue(req.Context(), ClientIDContextKey, clientID))
	} else if clientID = d4lcontext.GetClientID(req); clientID != "" {
		req = req.WithContext(context.WithValue(req.Context(), ClientIDContextKey, clientID))
	}

	if callerIP := l.ipParser(req); callerIP != "" {
		req = req.WithContext(context.WithValue(req.Context(), CallerIPContextKey, callerIP))
	}

	if tenantID := l.tenantIDParser(req); tenantID != "" {
		req = req.WithContext(context.WithValue(req.Context(), TenantIDContextKey, tenantID))
	} else if tenantID = d4lcontext.GetTenantID(req); tenantID != "" {
		req = req.WithContext(context.WithValue(req.Context(), TenantIDContextKey, tenantID))
	}

	req = req.WithContext(context.WithValue(req.Context(), RequestURLContextKey, req.URL.Path))
	req = req.WithContext(context.WithValue(req.Context(), RequestDomainContextKey, req.Host))

	reqTime := time.Now()
	_ = l.httpInRequest(req)

	recorder := responseWriter{ResponseWriter: rw, body: new(bytes.Buffer)}
	l.next.ServeHTTP(&recorder, req)

	if recorder.statusCode == 0 {
		recorder.statusCode = 200
	}

	_ = l.httpInResponse(
		req,
		recorder.Header(),
		recorder.statusCode,
		recorder.body,
		recorder.contentLength,
		reqTime,
	)
}

// WithUserParser is an option for NewHTTPLogger, that lets the caller pass a
// function for extracting the userID out of the HTTP request.
// That function will be called to inject the userID into the log line.
func WithUserParser(userParser func(*http.Request) string) func(*HTTPLogger) {
	return func(l *HTTPLogger) {
		l.userParser = userParser
	}
}

// WithClientIDParser is an option for NewHTTPLogger, that lets the caller pass a
// function for extracting the client ID out of the HTTP request.
func WithClientIDParser(clientParser func(*http.Request) string) func(*HTTPLogger) {
	return func(l *HTTPLogger) {
		l.clientParser = clientParser
	}
}

// WithCallerIPParser is an option for NewHTTPLogger, that lets the caller pass a
// function for extracting the caller IP out of the HTTP request.
func WithCallerIPParser(ipParser func(*http.Request) string) func(*HTTPLogger) {
	return func(l *HTTPLogger) {
		l.ipParser = ipParser
	}
}

// WithTenantIDParser is an option for NewHTTPLogger, that lets the caller pass a
// function for extracting the caller IP out of the HTTP request.
func WithTenantIDParser(tenantIDParser func(*http.Request) string) func(*HTTPLogger) {
	return func(l *HTTPLogger) {
		l.tenantIDParser = tenantIDParser
	}
}

type Field int

const (
	Body Field = iota
	ReqForm
)

type EventType string

const (
	HTTPInRequest  EventType = "http-in-request"
	HTTPInResponse EventType = "http-in-response"
)

// Obfuscator used to define which HTTP log event shall be obfuscated and how
type Obfuscator struct {
	EventType EventType
	ReqMethod string
	// ReqURL must match the request URL in order to apply this obfuscator
	// If nil or ".*" the request URL is ignored (i.e. all logs for given EventType and ReqMethod are obfuscated)
	ReqURL *regexp.Regexp

	Field   Field
	Replace *regexp.Regexp
	With    string
}

func WithObfuscators(o ...Obfuscator) func(*HTTPLogger) {
	return func(l *HTTPLogger) {
		for _, obf := range o {
			key := l.obfuscatorKey(obf.EventType, obf.ReqMethod)
			l.obf[key] = append(l.obf[key], obf)
		}
	}
}

type IPType string

const (
	IPTypeReal IPType = "real"
	IPTypeReq  IPType = "req"
	IPTypeAll  IPType = "all"
)

// IPAnonymizer define how IP addresses should be obfuscated from HTTP log event
type IPAnonymizer struct {
	IPType IPType
	With   string
}

func WithIPAnonymizers(s ...IPAnonymizer) func(*HTTPLogger) {
	return func(l *HTTPLogger) {
		l.ipa = append(l.ipa, s...)
	}
}

func (l *Logger) WrapHTTP(h http.Handler, options ...func(*HTTPLogger)) HTTPLogger {
	httpLogger := HTTPLogger{
		next:           h,
		log:            l,
		userParser:     func(_ *http.Request) string { return "" },
		clientParser:   func(_ *http.Request) string { return "" },
		ipParser:       func(_ *http.Request) string { return "" },
		tenantIDParser: func(_ *http.Request) string { return "" },
		obf:            make(map[string][]Obfuscator),
		ipa:            make([]IPAnonymizer, 0),
	}

	for _, apply := range options {
		apply(&httpLogger)
	}

	return httpLogger
}

func (l *Logger) HTTPMiddleware(options ...func(*HTTPLogger)) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return l.WrapHTTP(next, options...)
	}
}

// excludedContentType filters out content types that we do not want to log
// (e.g. doesn't make much sense to log application/octet-stream bodies consisting of binary data)
func excludedContentType(ct string) bool {
	return ct == "application/octet-stream"
}

// excludedContentEncoding filters out content encodings that we do not want to log
// (e.g. doesn't make much sense to log gzip'ed bodies consisting of binary data)
// also see: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Encoding
func excludedContentEncoding(ce string) bool {
	return ce == "gzip" || ce == "compress" || ce == "deflate" || ce == "br"
}

func filteredBodyStrFromBuffer(body *bytes.Buffer, header http.Header) string {
	bodyStr := excludedBodyStr(header)
	if bodyStr != "" {
		return bodyStr
	}
	return body.String()
}

func filteredBodyStrFromReq(req *http.Request) string {
	bodyStr := excludedBodyStr(req.Header)
	if bodyStr != "" {
		return bodyStr
	}

	if req.Body == nil {
		return ""
	}

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return fmt.Sprintf("error reading body: %v", err)
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	return string(bodyBytes)
}

func filteredBodyStrFromResp(resp *http.Response) string {
	bodyStr := excludedBodyStr(resp.Header)
	if bodyStr != "" {
		return bodyStr
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("error reading body: %v", err)
	}

	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	return string(bodyBytes)
}

func excludedBodyStr(header http.Header) string {
	ct := header.Get("Content-Type")
	ce := header.Get("Content-Encoding")
	if excludedContentType(ct) || excludedContentEncoding(ce) {
		return fmt.Sprintf("Content-Type: %s, Content-Encoding: %s is excluded from logging", ct, ce)
	}
	return ""
}
