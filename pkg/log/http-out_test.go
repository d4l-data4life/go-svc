package log_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/log"
	"github.com/gesundheitscloud/go-svc/pkg/transport"
)

func TestLogHttpOutRequest(t *testing.T) {
	t.Run("with full context", func(t *testing.T) {
		buf := new(bytes.Buffer)
		l := log.NewLogger("vega", "v1.0.0", "vega-123-123", log.WithWriter(buf))

		ctx := context.Background()
		ctx = context.WithValue(ctx, log.UserIDContextKey, "u1")
		ctx = context.WithValue(ctx, log.TraceIDContextKey, "t1")
		ctx = context.WithValue(ctx, log.ClientIDContextKey, "c1")

		if err := l.HttpOutRequest(
			ctx,
			"GET",
			"http://example.com/hey",
			145,
		); err != nil {
			t.Fatalf("logging: %v", err)
		}

		m := make(map[string]json.RawMessage)
		if err := json.NewDecoder(buf).Decode(&m); err != nil {
			t.Fatalf("unmarshaling JSON: %v", err)
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
				key:   "user-id",
				value: `"u1"`,
			},
			{
				key:   "service-name",
				value: `"vega"`,
			},
			{
				key:   "service-version",
				value: `"v1.0.0"`,
			},
			{
				key:   "hostname",
				value: `"vega-123-123"`,
			},
			{
				key:   "event-type",
				value: `"http-out-request"`,
			},
			{
				key:   "payload-length",
				value: `145`,
			},
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), m[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}
	})

	t.Run("with empty context", func(t *testing.T) {
		buf := new(bytes.Buffer)
		l := log.NewLogger("vega", "v1.0.0", "vega-123-123", log.WithWriter(buf))

		if err := l.HttpOutRequest(
			context.Background(),
			"GET",
			"http://example.com/hey",
			145,
		); err != nil {
			t.Fatalf("logging: %v", err)
		}

		m := make(map[string]json.RawMessage)
		if err := json.NewDecoder(buf).Decode(&m); err != nil {
			t.Fatalf("unmarshaling JSON: %v", err)
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
				value: `"vega"`,
			},
			{
				key:   "service-version",
				value: `"v1.0.0"`,
			},
			{
				key:   "hostname",
				value: `"vega-123-123"`,
			},
			{
				key:   "event-type",
				value: `"http-out-request"`,
			},
			{
				key:   "payload-length",
				value: `145`,
			},
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), m[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}
	})
}

func TestLogHTTPOutReqResp(t *testing.T) {

	buf := new(bytes.Buffer)
	l := log.NewLogger("name", "version", "hostname", log.WithWriter(buf))

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(400)
		_, _ = rw.Write([]byte("400 Bad Request"))
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequest("GET", srv.URL+"/the-path?query=param", strings.NewReader("Hello, PHDP!"))
	if err != nil {
		t.Fatalf("request creation failed: %v", err)
	}

	ctx := req.Context()
	ctx = context.WithValue(ctx, log.UserIDContextKey, "user123")
	ctx = context.WithValue(ctx, log.TraceIDContextKey, "t1")
	req = req.WithContext(ctx)
	req.Header.Add("trace-id", "t1")
	req.Header.Add("Content-Type", "text/plain")

	client := http.Client{
		Transport: transport.Log(l)(nil),
		Timeout:   10 * time.Second,
	}

	if _, err := client.Do(req); err != nil {
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
				value: `"GET"`,
			},
			{
				key:   "req-body",
				value: `"Hello, PHDP!"`,
			},
			{
				key:   "req-url",
				value: `"` + srv.URL + `/the-path?query=param"`,
			},
			{
				key:   "event-type",
				value: `"http-out-request"`,
			},
			{
				key:   "user-id",
				value: `"user123"`,
			},
			{
				key:   "payload-length",
				value: `12`,
			},
			{
				key:   "content-type",
				value: `"text/plain"`,
			},
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), requestLog[tc.key]; !bytes.Equal(want, have) {
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
				value: `"http-out-response"`,
			},
			{
				key:   "user-id",
				value: `"user123"`,
			},
			{
				key:   "response-code",
				value: `400`,
			},
			{
				key:   "response-body",
				value: `"400 Bad Request"`,
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
				key:   "req-method",
				value: `"GET"`,
			},
			{
				key:   "req-url",
				value: `"` + srv.URL + `/the-path?query=param"`,
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

func TestLogHTTPOutReqRespHeader(t *testing.T) {

	buf := new(bytes.Buffer)
	l := log.NewLogger("name", "version", "hostname", log.WithWriter(buf))

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(400)
		_, _ = rw.Write([]byte("400 Bad Request"))
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequest("GET", srv.URL+"/the-path?query=param", strings.NewReader("Hello, PHDP!"))
	if err != nil {
		t.Fatalf("request creation failed: %v", err)
	}

	req.Header.Add(log.TraceIDHeaderKey, "t1")

	client := http.Client{
		Transport: transport.Log(l)(nil),
		Timeout:   10 * time.Second,
	}

	if _, err := client.Do(req); err != nil {
		t.Fatalf("request failed: %v", err)
	}

	logs := json.NewDecoder(buf)

	t.Run("request: trace-id in header instead of ctx", func(t *testing.T) {
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
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), requestLog[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}
	})

	t.Run("response: trace-id in header instead of ctx", func(t *testing.T) {
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
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), responseLog[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}
	})
}

func TestLogHTTPOutReqRespHeaderLogging(t *testing.T) {

	buf := new(bytes.Buffer)
	l := log.NewLogger("name", "version", "hostname", log.WithWriter(buf))

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.Header().Set("authorization", "Bearer TakeMeOn")
		rw.WriteHeader(400)
		_, _ = rw.Write([]byte("400 Bad Request"))
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequest("GET", srv.URL+"/the-path?query=param", strings.NewReader("Hello, PHDP!"))
	if err != nil {
		t.Fatalf("request creation failed: %v", err)
	}

	req.Header.Add(log.TraceIDHeaderKey, "t1")
	req.Header.Add("authorization", "Bearer ThinkingAboutYou")
	client := http.Client{
		Transport: transport.Log(l)(nil),
		Timeout:   10 * time.Second,
	}

	if _, err := client.Do(req); err != nil {
		t.Fatalf("request failed: %v", err)
	}

	logs := json.NewDecoder(buf)

	t.Run("request header", func(t *testing.T) {
		requestLog := make(map[string]json.RawMessage)
		if err := logs.Decode(&requestLog); err != nil {
			t.Fatalf("unmarshaling the request log: %v", err)
		}

		for _, tc := range [...]struct {
			key   string
			value string
		}{
			{
				key:   "header",
				value: `{"Authorization":["Obfuscated{23}"]}`,
			},
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), requestLog[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}
	})

	t.Run("response header", func(t *testing.T) {
		responseLog := make(map[string]json.RawMessage)
		if err := logs.Decode(&responseLog); err != nil {
			t.Fatalf("unmarshaling the response log: %v", err)
		}

		for _, tc := range [...]struct {
			key   string
			value string
		}{
			{
				key:   "header",
				value: `{"Authorization":["Obfuscated{15}"]}`,
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
