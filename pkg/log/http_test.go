package log_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/gofrs/uuid"

	"github.com/d4l-data4life/go-svc/pkg/log"
)

const (
	userID   = "user123"
	clientID = "c1"
	tenantID = "tenant1"
)

func TestHTTPWrapper(t *testing.T) {
	t.Run("response log of a minimal handler", func(t *testing.T) {
		buf := new(bytes.Buffer)
		l := log.NewLogger("name", "version", "hostname", log.WithWriter(buf))

		// This handler does not explicitly set the response code, and
		// has a response body of 0 bytes.
		handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

		wrapped := l.WrapHTTP(
			handler,
		)

		srv := httptest.NewServer(wrapped)
		defer srv.Close()

		req, err := http.NewRequest("OPTIONS", srv.URL+"/the-path?query=param", strings.NewReader("你好，世界"))
		if err != nil {
			t.Fatalf("request creation failed: %v", err)
		}

		if _, err := srv.Client().Do(req); err != nil {
			t.Fatalf("request failed: %v", err)
		}

		logs := json.NewDecoder(buf)

		requestLog := make(map[string]json.RawMessage)
		if err := logs.Decode(&requestLog); err != nil {
			t.Fatalf("unmarshaling the request log: %v", err)
		}

		// We are not interested in the request log in this test.
		_ = requestLog

		responseLog := make(map[string]json.RawMessage)
		if err := logs.Decode(&responseLog); err != nil {
			t.Fatalf("unmarshaling the response log: %v", err)
		}

		for _, tc := range [...]struct {
			key   string
			value string
		}{
			{
				key:   "log-level",
				value: fmt.Sprintf(`"%s"`, string(log.LevelInfo)),
			},
			{
				key:   "service-name",
				value: `"name"`,
			},
			{
				key:   "service-version",
				value: `"version"`,
			},
			{
				key:   "hostname",
				value: `"hostname"`,
			},
			{
				key:   "event-type",
				value: `"http-in-response"`,
			},
			{
				key:   "response-code",
				value: `200`,
			},
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), responseLog[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}
	})

	t.Run("preexisting context is not overwritten with empty strings", func(t *testing.T) {
		buf := new(bytes.Buffer)
		l := log.NewLogger("name", "version", "hostname", log.WithWriter(buf))

		handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

		wrapped := l.WrapHTTP(
			handler,
		)

		addContextFirst := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			ctx = context.WithValue(ctx, log.UserIDContextKey, "preset userID")
			ctx = context.WithValue(ctx, log.TraceIDContextKey, "preset traceID")
			ctx = context.WithValue(ctx, log.ClientIDContextKey, "preset clientID")
			ctx = context.WithValue(ctx, log.TenantIDContextKey, "preset tenantID")
			wrapped.ServeHTTP(rw, req.WithContext(ctx))
		})

		srv := httptest.NewServer(addContextFirst)
		defer srv.Close()

		req, err := http.NewRequest("POST", srv.URL+"/the-path?query=param", strings.NewReader("你好，世界"))
		if err != nil {
			t.Fatalf("request creation failed: %v", err)
		}

		if _, err := srv.Client().Do(req); err != nil {
			t.Fatalf("request failed: %v", err)
		}

		logs := json.NewDecoder(buf)

		requestLog := make(map[string]json.RawMessage)
		if err := logs.Decode(&requestLog); err != nil {
			t.Fatalf("unmarshaling the request log: %v", err)
		}

		for _, tc := range [...]struct {
			key   string
			value string
		}{
			{
				key:   "trace-id",
				value: `"preset traceID"`,
			},
			{
				key:   "user-id",
				value: `"preset userID"`,
			},
			{
				key:   "client-id",
				value: `"preset clientID"`,
			},
			{
				key:   "tenant-id",
				value: `"preset tenantID"`,
			},
		} {
			t.Run("contains preset "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), requestLog[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}
	})

	t.Run("nil req body is handled properly ", func(t *testing.T) {
		buf := new(bytes.Buffer)
		l := log.NewLogger("name", "version", "hostname", log.WithWriter(buf))

		handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

		wrapped := l.WrapHTTP(
			handler,
		)

		srv := httptest.NewServer(wrapped)
		defer srv.Close()

		req := httptest.NewRequest("OPTIONS", srv.URL+"/the-path?query=param", nil)

		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		logs := json.NewDecoder(buf)

		requestLog := make(map[string]json.RawMessage)
		if err := logs.Decode(&requestLog); err != nil {
			t.Fatalf("unmarshaling the request log: %v", err)
		}

		for _, tc := range [...]struct {
			key   string
			value string
		}{
			{
				key:   "service-name",
				value: `"name"`,
			},
			{
				key:   "log-level",
				value: fmt.Sprintf(`"%s"`, string(log.LevelInfo)),
			},
			{
				key:   "service-version",
				value: `"version"`,
			},
			{
				key:   "hostname",
				value: `"hostname"`,
			},
			{
				key:   "req-method",
				value: `"OPTIONS"`,
			},
			{
				key:   "req-body",
				value: `""`,
			},
		} {
			t.Run("contains preset "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), requestLog[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}
	})

	// Now we test the full chain: HTTP request, an error, HTTP response.

	buf := new(bytes.Buffer)
	l := log.NewLogger("name", "version", "hostname", log.WithWriter(buf))

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if err := l.ErrGeneric(req.Context(), http.ErrHijacked); err != nil {
			t.Fatalf("unexpected log error: %v", err)
		}
		// for header obfuscation/ignore
		rw.Header().Set("Set-Cookie", "jwt=jwt")
		rw.WriteHeader(409)
		_, _ = rw.Write([]byte("this is the response. 再见!"))
	})

	wrapped := l.WrapHTTP(
		handler,
		log.WithUserParser(func(_ *http.Request) string {
			return userID
		}),
		log.WithClientIDParser(func(_ *http.Request) string {
			return clientID
		}),
		log.WithTenantIDParser(func(_ *http.Request) string {
			return tenantID
		}),
	)

	srv := httptest.NewServer(wrapped)
	defer srv.Close()

	req, err := http.NewRequest("OPTIONS", srv.URL+"/the-path?query=param", strings.NewReader("你好，世界"))
	if err != nil {
		t.Fatalf("request creation failed: %v", err)
	}

	req.Header.Set("trace-id", "t1")
	req.Header.Set("client-id", clientID)
	req.Header.Set("x-real-ip", "10.0.0.2")
	req.Header.Set("Authorization", "Bearer PrayThatThisIsObfuscated")
	req.Header.Set("Cookie", "jwt=jwt")

	if _, err := srv.Client().Do(req); err != nil {
		t.Fatalf("request failed: %v", err)
	}

	logs := json.NewDecoder(buf)

	t.Run("request log", func(t *testing.T) {
		requestLog := make(map[string]json.RawMessage)
		if err := logs.Decode(&requestLog); err != nil {
			t.Fatalf("unmarshaling the request log: %v", err)
		}

		for _, tc := range [...]struct {
			key   string
			value string
		}{
			{
				key:   "trace-id",
				value: `"t1"`,
			},
			{
				key:   "client-id",
				value: `"c1"`,
			},
			{
				key:   "tenant-id",
				value: `"` + tenantID + `"`,
			},
			{
				key:   "log-level",
				value: fmt.Sprintf(`"%s"`, string(log.LevelInfo)),
			},
			{
				key:   "service-name",
				value: `"name"`,
			},
			{
				key:   "service-version",
				value: `"version"`,
			},
			{
				key:   "hostname",
				value: `"hostname"`,
			},
			{
				key:   "req-method",
				value: `"OPTIONS"`,
			},
			{
				key:   "req-body",
				value: `"你好，世界"`, // hello world
			},
			{
				key: "req-form",
				value: fmt.Sprintf("\"%s\"", map[string]string{
					"query": "[param]",
				}),
			},
			{
				key:   "req-url",
				value: `"/the-path?query=param"`,
			},
			{
				key:   "real-ip",
				value: `"10.0.0.2"`,
			},
			{
				key:   "event-type",
				value: `"http-in-request"`,
			},
			{
				key:   "user-id",
				value: fmt.Sprintf("%q", userID),
			},
			{
				key:   "payload-length",
				value: `15`,
			}, {
				key:   "header",
				value: `{"Authorization":["Obfuscated{31}"],"Client-Id":["c1"],"Cookie":["jwt=Obfuscated{3};"],"User-Agent":["Go-http-client/1.1"]}`,
			},
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), requestLog[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}
	})

	t.Run("error log", func(t *testing.T) {
		errorLog := make(map[string]json.RawMessage)
		if err := logs.Decode(&errorLog); err != nil {
			t.Fatalf("unmarshaling the error log: %v", err)
		}

		for _, tc := range [...]struct {
			key   string
			value string
		}{
			{
				key:   "trace-id",
				value: `"t1"`,
			},
			{
				key:   "client-id",
				value: `"c1"`,
			},
			{
				key:   "tenant-id",
				value: `"` + tenantID + `"`,
			},
			{
				key:   "log-level",
				value: fmt.Sprintf(`"%s"`, string(log.LevelError)),
			},
			{
				key:   "service-name",
				value: `"name"`,
			},
			{
				key:   "service-version",
				value: `"version"`,
			},
			{
				key:   "user-id",
				value: fmt.Sprintf("%q", userID),
			},
			{
				key:   "message",
				value: `"` + http.ErrHijacked.Error() + `"`,
			},
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), errorLog[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}
	})

	t.Run("response log", func(t *testing.T) {
		responseLog := make(map[string]json.RawMessage)
		if err := logs.Decode(&responseLog); err != nil {
			t.Fatalf("unmarshaling the response log: %v", err)
		}

		for _, tc := range [...]struct {
			key   string
			value string
		}{
			{
				key:   "trace-id",
				value: `"t1"`,
			},
			{
				key:   "client-id",
				value: `"c1"`,
			},
			{
				key:   "tenant-id",
				value: `"` + tenantID + `"`,
			},
			{
				key:   "log-level",
				value: fmt.Sprintf(`"%s"`, string(log.LevelError)),
			},
			{
				key:   "service-name",
				value: `"name"`,
			},
			{
				key:   "service-version",
				value: `"version"`,
			},
			{
				key:   "hostname",
				value: `"hostname"`,
			},
			{
				key:   "event-type",
				value: `"http-in-response"`,
			},
			{
				key:   "user-id",
				value: fmt.Sprintf("%q", userID),
			},
			{
				key:   "response-code",
				value: `409`,
			},
			{
				key:   "response-body",
				value: `"this is the response. 再见!"`, // good bye
			},
			{
				key:   "payload-length",
				value: `29`,
			}, {
				key:   "header",
				value: `{"Set-Cookie":["jwt=Obfuscated{3}"]}`,
			},
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), responseLog[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}
	})
}

func TestTenantIDOverriding(t *testing.T) {
	buf := new(bytes.Buffer)

	for _, tc := range [...]struct {
		name              string
		logger            *log.Logger
		httpWrapers       []func(*log.HTTPLogger)
		expectedKeyValues map[string]string
	}{
		{
			name:        "should work with the tenant ID defined in the logger instance",
			logger:      log.NewLogger("name", "version", "hostname", log.WithWriter(buf), log.WithTenantID("tenant-1")),
			httpWrapers: []func(*log.HTTPLogger){},
			expectedKeyValues: map[string]string{
				"tenant-id": `"tenant-1"`,
			},
		},
		{
			name:   "tenant ID from the context should override the constructor one",
			logger: log.NewLogger("name", "version", "hostname", log.WithWriter(buf), log.WithTenantID("tenant-1-constr")),
			httpWrapers: []func(*log.HTTPLogger){
				log.WithTenantIDParser(func(_ *http.Request) string {
					return "tenant-1-http"
				}),
			},
			expectedKeyValues: map[string]string{
				"tenant-id": `"tenant-1-http"`,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

			wrapped := tc.logger.WrapHTTP(
				handler,
				tc.httpWrapers...,
			)

			srv := httptest.NewServer(wrapped)
			defer srv.Close()

			req, err := http.NewRequest("POST", srv.URL+"/the-path?query=param", strings.NewReader("你好，世界"))
			if err != nil {
				t.Fatalf("request creation failed: %v", err)
			}

			if _, err := srv.Client().Do(req); err != nil {
				t.Fatalf("request failed: %v", err)
			}

			logs := json.NewDecoder(buf)

			requestLog := make(map[string]json.RawMessage)
			if err := logs.Decode(&requestLog); err != nil {
				t.Fatalf("unmarshaling the request log: %v", err)
			}
			for key, wantValue := range tc.expectedKeyValues {
				if want, have := []byte(wantValue), requestLog[key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", key, want, have)
				}
			}

			buf.Reset()
		})
	}
}

func buildContext(options ...func(context.Context) context.Context) context.Context {
	ctx := context.Background()
	for _, f := range options {
		ctx = f(ctx)
	}

	return ctx
}

func withValue(key interface{}, val interface{}) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, key, val)
	}
}

func TestContextOverriding(t *testing.T) {
	buf := new(bytes.Buffer)

	baseReq, err := http.NewRequest("POST", "/path", strings.NewReader("你好，世界"))
	if err != nil {
		t.Fatalf("request creation failed: %v", err)
	}

	userID := uuid.Must(uuid.NewV4())

	for _, tc := range [...]struct {
		name              string
		logger            *log.Logger
		req               *http.Request
		httpWrapers       []func(*log.HTTPLogger)
		expectedKeyValues map[string]string
	}{
		{
			name:   "explicit values extractors should override the values from the d4lcontext",
			logger: log.NewLogger("name", "version", "hostname", log.WithWriter(buf)),
			req: baseReq.WithContext(buildContext(
				withValue(log.UserIDContextKey, userID),
				withValue(log.ClientIDContextKey, "ctx-client-id"),
				withValue(log.TenantIDContextKey, "ctx-tenant-id"),
			)),
			httpWrapers: []func(*log.HTTPLogger){
				log.WithTenantIDParser(func(_ *http.Request) string {
					return "tenant-1-req"
				}),
				log.WithClientIDParser(func(_ *http.Request) string {
					return "client-1-req"
				}),
				log.WithUserParser(func(_ *http.Request) string {
					return "user-1-req"
				}),
			},
			expectedKeyValues: map[string]string{
				"tenant-id": `"tenant-1-req"`,
				"user-id":   `"user-1-req"`,
				"client-id": `"client-1-req"`,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

			wrapper := tc.logger.HTTPMiddleware(
				tc.httpWrapers...,
			)

			wrapper(handler).ServeHTTP(httptest.NewRecorder(), tc.req)

			logs := json.NewDecoder(buf)

			requestLog := make(map[string]json.RawMessage)
			if err := logs.Decode(&requestLog); err != nil {
				t.Fatalf("unmarshaling the request log: %v", err)
			}
			for key, wantValue := range tc.expectedKeyValues {
				if want, have := []byte(wantValue), requestLog[key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", key, want, have)
				}
			}

			buf.Reset()
		})
	}
}

func TestFilteredContentType(t *testing.T) {
	buf := new(bytes.Buffer)
	l := log.NewLogger("name", "version", "hostname", log.WithWriter(buf))

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if err := l.ErrGeneric(req.Context(), http.ErrHijacked); err != nil {
			t.Fatalf("unexpected log error: %v", err)
		}
		rw.Header().Set("Content-Type", "application/octet-stream")
		// for header logging
		rw.Header().Set("trace-id", "t1")
		rw.Header().Set("client-id", "c1")
		rw.WriteHeader(http.StatusConflict)
		_, _ = rw.Write([]byte("this is the response. 再见!"))
	})

	wrapped := l.WrapHTTP(
		handler,
		log.WithUserParser(func(_ *http.Request) string {
			return userID
		}),
		log.WithClientIDParser(func(_ *http.Request) string {
			return clientID
		}),
	)

	srv := httptest.NewServer(wrapped)
	defer srv.Close()

	req, err := http.NewRequest("OPTIONS", srv.URL+"/the-path?query=param", strings.NewReader("你好，世界"))
	if err != nil {
		t.Fatalf("request creation failed: %v", err)
	}

	req.Header.Set("trace-id", "t1")
	req.Header.Set("client-id", "c1")
	req.Header.Set("x-real-ip", "10.0.0.2")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	if _, err := srv.Client().Do(req); err != nil {
		t.Fatalf("request failed: %v", err)
	}

	logs := json.NewDecoder(buf)

	t.Run("request log", func(t *testing.T) {
		requestLog := make(map[string]json.RawMessage)
		if err := logs.Decode(&requestLog); err != nil {
			t.Fatalf("unmarshaling the request log: %v", err)
		}

		for _, tc := range [...]struct {
			key   string
			value string
		}{
			{
				key:   "trace-id",
				value: `"t1"`,
			},
			{
				key:   "client-id",
				value: `"c1"`,
			},
			{
				key:   "log-level",
				value: fmt.Sprintf(`"%s"`, string(log.LevelInfo)),
			},
			{
				key:   "service-name",
				value: `"name"`,
			},
			{
				key:   "service-version",
				value: `"version"`,
			},
			{
				key:   "hostname",
				value: `"hostname"`,
			},
			{
				key:   "req-method",
				value: `"OPTIONS"`,
			},
			{
				key:   "req-body",
				value: `"Content-Type: application/json, Content-Encoding: gzip is excluded from logging"`, // hello world
			},
			{
				key: "req-form",
				value: fmt.Sprintf("\"%s\"", map[string]string{
					"query": "[param]",
				}),
			},
			{
				key:   "req-url",
				value: `"/the-path?query=param"`,
			},
			{
				key:   "real-ip",
				value: `"10.0.0.2"`,
			},
			{
				key:   "event-type",
				value: `"http-in-request"`,
			},
			{
				key:   "user-id",
				value: fmt.Sprintf("%q", userID),
			},
			{
				key:   "payload-length",
				value: `15`,
			},
			{
				key:   "content-type",
				value: `"application/json"`,
			},
			{
				key:   "content-encoding",
				value: `"gzip"`,
			}, {
				key:   "header",
				value: `{"Client-Id":["c1"],"User-Agent":["Go-http-client/1.1"]}`,
			},
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), requestLog[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}
	})

	t.Run("error log", func(t *testing.T) {
		errorLog := make(map[string]json.RawMessage)
		if err := logs.Decode(&errorLog); err != nil {
			t.Fatalf("unmarshaling the error log: %v", err)
		}

		for _, tc := range [...]struct {
			key   string
			value string
		}{
			{
				key:   "trace-id",
				value: `"t1"`,
			},
			{
				key:   "client-id",
				value: `"c1"`,
			},
			{
				key:   "log-level",
				value: fmt.Sprintf(`"%s"`, string(log.LevelError)),
			},
			{
				key:   "service-name",
				value: `"name"`,
			},
			{
				key:   "service-version",
				value: `"version"`,
			},
			{
				key:   "user-id",
				value: fmt.Sprintf("%q", userID),
			},
			{
				key:   "message",
				value: `"` + http.ErrHijacked.Error() + `"`,
			},
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), errorLog[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}
	})

	t.Run("response log", func(t *testing.T) {
		responseLog := make(map[string]json.RawMessage)
		if err := logs.Decode(&responseLog); err != nil {
			t.Fatalf("unmarshaling the response log: %v", err)
		}

		for _, tc := range [...]struct {
			key   string
			value string
		}{
			{
				key:   "trace-id",
				value: `"t1"`,
			},
			{
				key:   "client-id",
				value: `"c1"`,
			},
			{
				key:   "log-level",
				value: fmt.Sprintf(`"%s"`, string(log.LevelError)),
			},
			{
				key:   "service-name",
				value: `"name"`,
			},
			{
				key:   "service-version",
				value: `"version"`,
			},
			{
				key:   "hostname",
				value: `"hostname"`,
			},
			{
				key:   "event-type",
				value: `"http-in-response"`,
			},
			{
				key:   "user-id",
				value: fmt.Sprintf("%q", userID),
			},
			{
				key:   "response-code",
				value: `409`,
			},
			{
				key:   "response-body",
				value: `"Content-Type: application/octet-stream, Content-Encoding:  is excluded from logging"`, // good bye
			},
			{
				key:   "payload-length",
				value: `29`,
			},
			{
				key:   "content-type",
				value: `"application/octet-stream"`,
			},
			{
				key:   "req-method",
				value: `"OPTIONS"`,
			},
			{
				key:   "req-url",
				value: `"/the-path?query=param"`,
			}, {
				key:   "header",
				value: `{"Client-Id":["c1"]}`,
			},
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), responseLog[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}
	})
}

func TestObfuscator(t *testing.T) {
	buf := new(bytes.Buffer)
	l := log.NewLogger("name", "version", "hostname", log.WithWriter(buf))

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			rw.Header().Set("Content-Type", "application/json")
			_, _ = rw.Write([]byte(`{"password":"MySuperSecurePassword123", "otherStuff":"someValue", "secret":"dontexposethis"}`))
		case http.MethodPut:
			rw.Header().Set("Content-Type", "application/json")
			_, _ = rw.Write([]byte(`{"super-safe":"init123", "amazingly-secret":"password"}`))
		default:
			rw.WriteHeader(http.StatusForbidden)
		}
	})

	wrapped := l.WrapHTTP(
		handler,
		log.WithObfuscators(
			log.Obfuscator{
				EventType: log.HTTPInRequest,
				ReqMethod: http.MethodPost,
				ReqURL:    regexp.MustCompile("^/the-path/[a-zA-Z0-9]*/something$"),
				Field:     log.Body,
				Replace:   regexp.MustCompile(`"password":".*?"`),
				With:      `"password":"obfuscatedPostReqBody"`,
			},
			log.Obfuscator{
				EventType: log.HTTPInRequest,
				ReqMethod: http.MethodPost,
				ReqURL:    regexp.MustCompile("^/the-path/[a-zA-Z0-9]*/something$"),
				Field:     log.Body,
				Replace:   regexp.MustCompile(`"secret":".*?"`),
				With:      `"secret":"obfuscatedPostReqBody"`,
			},
			log.Obfuscator{
				EventType: log.HTTPInRequest,
				ReqMethod: http.MethodPost,
				ReqURL:    regexp.MustCompile("^/the-path/[a-zA-Z0-9]*/something$"),
				Field:     log.ReqForm,
				Replace:   regexp.MustCompile(`"secret":".*?"`),
				With:      `"secret":"obfuscatedPostReqForm"`,
			},
			log.Obfuscator{
				EventType: log.HTTPInRequest,
				ReqMethod: http.MethodPost,
				ReqURL:    regexp.MustCompile("^/other-path$"),
				Field:     log.Body,
				Replace:   regexp.MustCompile(`"otherStuff":".*?"`),
				With:      `"otherStuff":"obfuscatedPostReqBody"`,
			},
			log.Obfuscator{
				EventType: log.HTTPInRequest,
				ReqMethod: http.MethodGet,
				ReqURL:    regexp.MustCompile("^/the-path/[a-zA-Z0-9]*/something$"),
				Field:     log.Body,
				Replace:   regexp.MustCompile(`"secret":".*?"`),
				With:      `"secret":"obfuscatedGetReqBody"`,
			},
			log.Obfuscator{
				EventType: log.HTTPInResponse,
				ReqMethod: http.MethodPost,
				ReqURL:    regexp.MustCompile("^/the-path/[a-zA-Z0-9]*/something$"),
				Field:     log.Body,
				Replace:   regexp.MustCompile(`"secret":".*?"`),
				With:      `"secret":"obfuscatedPostRespBody"`,
			},
			log.Obfuscator{
				EventType: log.HTTPInRequest,
				ReqMethod: http.MethodPost,
				ReqURL:    regexp.MustCompile("^/post-form-data$"),
				Field:     log.ReqForm,
				Replace:   regexp.MustCompile(`password:\[.*?\]`),
				With:      `password:[obfuscatedReqForm]`,
			},
			log.Obfuscator{
				EventType: log.HTTPInResponse,
				ReqMethod: http.MethodPost,
				ReqURL:    regexp.MustCompile("^/post-form-data$"),
				Field:     log.Body,
				Replace:   regexp.MustCompile(`"secret":".*?"`),
				With:      `"secret":"obfuscatedPostFormRespBody"`,
			},
			log.Obfuscator{
				EventType: log.HTTPInRequest,
				ReqMethod: http.MethodPut,
				Field:     log.Body,
				Replace:   regexp.MustCompile(`"super-safe":".*?"`),
				With:      `"super-safe":"obfuscatedNilReqURL"`,
			},
			log.Obfuscator{
				EventType: log.HTTPInRequest,
				ReqMethod: http.MethodPut,
				ReqURL:    regexp.MustCompile(".*"),
				Field:     log.Body,
				Replace:   regexp.MustCompile(`"amazingly-secret":".*?"`),
				With:      `"amazingly-secret":"obfuscatedMatchAllReqURL"`,
			},
			log.Obfuscator{
				EventType: log.HTTPInResponse,
				ReqMethod: http.MethodPut,
				Field:     log.Body,
				Replace:   regexp.MustCompile(`"super-safe":".*?"`),
				With:      `"super-safe":"obfuscatedRespNilReqURL"`,
			},
			log.Obfuscator{
				EventType: log.HTTPInResponse,
				ReqMethod: http.MethodPut,
				ReqURL:    regexp.MustCompile(".*"),
				Field:     log.Body,
				Replace:   regexp.MustCompile(`"amazingly-secret":".*?"`),
				With:      `"amazingly-secret":"obfuscatedRespMatchAllReqURL"`,
			},
		),
	)

	srv := httptest.NewServer(wrapped)
	defer srv.Close()

	t.Run("post request with body", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/the-path/id/something",
			strings.NewReader(`{"password":"MySuperSecurePassword123", "otherStuff":"someValue", "secret":"dontexposethis"}`))
		if err != nil {
			t.Fatalf("request creation failed: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")

		if _, err := srv.Client().Do(req); err != nil {
			t.Fatalf("request failed: %v", err)
		}

		logs := json.NewDecoder(buf)

		requestLog := make(map[string]json.RawMessage)
		if err := logs.Decode(&requestLog); err != nil {
			t.Fatalf("unmarshaling the request log: %v", err)
		}

		responseLog := make(map[string]json.RawMessage)
		if err := logs.Decode(&responseLog); err != nil {
			t.Fatalf("unmarshaling the response log: %v", err)
		}

		for _, tc := range [...]struct {
			key   string
			value string
		}{
			{
				key:   "req-body",
				value: `"{\"password\":\"obfuscatedPostReqBody\", \"otherStuff\":\"someValue\", \"secret\":\"obfuscatedPostReqBody\"}"`,
			},
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), requestLog[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}

		for _, tc := range [...]struct {
			key   string
			value string
		}{
			{
				key:   "response-body",
				value: `"{\"password\":\"MySuperSecurePassword123\", \"otherStuff\":\"someValue\", \"secret\":\"obfuscatedPostRespBody\"}"`,
			},
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), responseLog[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}
	})

	t.Run("post request with body - test nil and matchAll regex for ReqURL", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPut, srv.URL+"/the-path/id/something",
			strings.NewReader(`{"super-safe":"MySuperSecurePassword123", "amazingly-secret":"someValue"}`))
		if err != nil {
			t.Fatalf("request creation failed: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")

		if _, err := srv.Client().Do(req); err != nil {
			t.Fatalf("request failed: %v", err)
		}

		logs := json.NewDecoder(buf)

		requestLog := make(map[string]json.RawMessage)
		if err := logs.Decode(&requestLog); err != nil {
			t.Fatalf("unmarshaling the request log: %v", err)
		}

		responseLog := make(map[string]json.RawMessage)
		if err := logs.Decode(&responseLog); err != nil {
			t.Fatalf("unmarshaling the response log: %v", err)
		}

		for _, tc := range [...]struct {
			key   string
			value string
		}{
			{
				key:   "req-body",
				value: `"{\"super-safe\":\"obfuscatedNilReqURL\", \"amazingly-secret\":\"obfuscatedMatchAllReqURL\"}"`,
			},
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), requestLog[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}

		for _, tc := range [...]struct {
			key   string
			value string
		}{
			{
				key:   "response-body",
				value: `"{\"super-safe\":\"obfuscatedRespNilReqURL\", \"amazingly-secret\":\"obfuscatedRespMatchAllReqURL\"}"`,
			},
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), responseLog[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}
	})

	t.Run("post request with form data", func(t *testing.T) {
		_, err := http.PostForm(srv.URL+"/post-form-data", url.Values{"someKey": {"someValue 123"}, "password": {"123"}})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		logs := json.NewDecoder(buf)

		requestLog := make(map[string]json.RawMessage)
		if err := logs.Decode(&requestLog); err != nil {
			t.Fatalf("unmarshaling the request log: %v", err)
		}

		responseLog := make(map[string]json.RawMessage)
		if err := logs.Decode(&responseLog); err != nil {
			t.Fatalf("unmarshaling the response log: %v", err)
		}

		for _, tc := range [...]struct {
			key   string
			value string
		}{
			{
				key:   "req-form",
				value: `"map[password:[obfuscatedReqForm] someKey:[someValue 123]]"`,
			},
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), requestLog[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}

		for _, tc := range [...]struct {
			key   string
			value string
		}{
			{
				key:   "response-body",
				value: `"{\"password\":\"MySuperSecurePassword123\", \"otherStuff\":\"someValue\", \"secret\":\"obfuscatedPostFormRespBody\"}"`,
			},
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), responseLog[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}
	})
}

func TestAnonymizeIP(t *testing.T) {
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"data":"cool!"}`))
	})

	t.Run("GET request with ReqIP", func(t *testing.T) {
		for _, tc := range [...]struct {
			name          string
			key           string
			value         string
			anonymizer    log.IPAnonymizer
			expectedMatch bool
		}{
			{
				name:          "change request IP",
				key:           "req-ip",
				value:         `"1.1.1.1"`,
				anonymizer:    log.IPAnonymizer{IPType: log.IPTypeReq, With: "1.1.1.1"},
				expectedMatch: true,
			},
			{
				name:          "Change real IP",
				key:           "real-ip",
				value:         `"1.1.1.2"`,
				anonymizer:    log.IPAnonymizer{IPType: log.IPTypeReal, With: "1.1.1.2"},
				expectedMatch: true,
			},
			{
				name:          "Change both IPs - check request IP",
				key:           "req-ip",
				value:         `"1.1.2.2"`,
				anonymizer:    log.IPAnonymizer{IPType: log.IPTypeAll, With: "1.1.2.2"},
				expectedMatch: true,
			},
			{
				name:          "Change both IPs - check real IP",
				key:           "real-ip",
				value:         `"1.1.2.2"`,
				anonymizer:    log.IPAnonymizer{IPType: log.IPTypeAll, With: "1.1.2.2"},
				expectedMatch: true,
			},
			{
				name:          "Change real IP to non-IP string",
				key:           "real-ip",
				value:         `"foobardummy"`,
				anonymizer:    log.IPAnonymizer{IPType: log.IPTypeAll, With: "foobardummy"},
				expectedMatch: true,
			},
			{
				name:          "Use empty IPAnonymizer for req-ip",
				key:           "req-ip",
				value:         `""`, // expect real local IP, for local tests: 127.0.0.1:port
				anonymizer:    log.IPAnonymizer{},
				expectedMatch: false,
			},
			{
				name:          "Use empty IPAnonymizer for real-ip",
				key:           "real-ip",
				value:         `""`, // expect empty string
				anonymizer:    log.IPAnonymizer{},
				expectedMatch: false, // but it will not show up due to omitempty
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				// CASE SETUP
				buf := new(bytes.Buffer)
				l := log.NewLogger("name", "version", "hostname", log.WithWriter(buf))
				wrapped := l.WrapHTTP(
					handler,
					log.WithIPAnonymizers(tc.anonymizer),
				)

				srv := httptest.NewServer(wrapped)
				defer srv.Close()

				_, err := http.Get(srv.URL)
				if err != nil {
					t.Fatalf("request failed: %v", err)
				}
				logs := json.NewDecoder(buf)

				requestLog := make(map[string]json.RawMessage)
				if err := logs.Decode(&requestLog); err != nil {
					t.Fatalf("unmarshaling the request log: %v", err)
				}

				// TEST
				want, have := []byte(tc.value), requestLog[tc.key]
				if tc.expectedMatch && !bytes.Equal(want, have) {
					t.Errorf("Expected a requestLog match in: %s", tc.name)
				}
				if !tc.expectedMatch && bytes.Equal(want, have) {
					t.Errorf("Expected no requestLog match in: %s", tc.name)
				}
			})
		}
	})
}
