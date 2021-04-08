package cors_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/cors"
)

func TestWrap(t *testing.T) {
	h := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Custom-Header", "value")
		rw.WriteHeader(http.StatusTeapot)
		_, _ = rw.Write([]byte("So much fun."))
	})

	for _, tc := range [...]struct {
		name         string
		wrapped      http.Handler
		origin       string
		corsExpected bool
	}{
		{
			"Default allows some Origin",
			cors.Wrap(h),
			"https://app.gesundheitscloud.de",
			true,
		},
		{
			"Default allows no Origin",
			cors.Wrap(h),
			"",
			true,
		},
		{
			"Domain restriction allows listed",
			cors.Wrap(h, cors.WithDomainList("https://example.com")),
			"https://example.com",
			true,
		},
		{
			"Domain restriction denies not listed",
			cors.Wrap(h, cors.WithDomainList("https://example.com")),
			"https://gesundheitscloud.de",
			false,
		},
		{
			"Domain restriction denies not listed in multiple list",
			cors.Wrap(h, cors.WithDomainList("https://example.com, http://hpsgc.de")),
			"https://gesundheitscloud.de",
			false,
		},
		{
			"Domain restriction allows listed in multiple list",
			cors.Wrap(h, cors.WithDomainList("https://example.com, http://hpsgc.de")),
			"http://hpsgc.de",
			true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}

			rec := httptest.NewRecorder()

			tc.wrapped.ServeHTTP(rec, req)

			res := rec.Result()

			corsFound := res.Header.Get("Access-Control-Allow-Origin") != ""

			if tc.corsExpected && !corsFound {
				t.Error("CORS header not found")
			}

			if !tc.corsExpected && corsFound {
				t.Error("unexpected CORS header found")
			}

			if want, have := 418, res.StatusCode; want != have {
				t.Errorf("expected status code %d, found %d", want, have)
			}
		})
	}

	t.Run("Only the current Origin is shown as allowed", func(t *testing.T) {
		wrapped := cors.Wrap(h, cors.WithDomainList("http://example.com, https://example.com,https://gesundheitscloud.de"))

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Origin", "https://gesundheitscloud.de")
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if want, have := "https://gesundheitscloud.de", rec.Result().Header.Get("Access-Control-Allow-Origin"); want != have {
			t.Errorf("expected allowed origin %q, found %q", want, have)
		}
	})

	t.Run("An OPTIONS call does not trigger the wrapped handler", func(t *testing.T) {
		rec := httptest.NewRecorder()
		cors.Wrap(h).ServeHTTP(rec, httptest.NewRequest("OPTIONS", "/", nil))
		if want, have := 200, rec.Result().StatusCode; want != have {
			t.Errorf("expected status code %d, found %d", want, have)
		}
	})

	t.Run("Custom allowed methods", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("OPTIONS", "/", nil)
		cors.Wrap(h, cors.WithMethodList("OPTIONS, HEAD, PATCH")).ServeHTTP(rec, req)

		if want, have := "OPTIONS, HEAD, PATCH", rec.Result().Header.Get("Access-Control-Allow-Methods"); want != have {
			t.Errorf("expected allowed methods %q, found %q", want, have)
		}

	})

	t.Run("Custom allowed headers", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("OPTIONS", "/", nil)
		cors.Wrap(h, cors.WithHeaderList("X-CSRF-Token, X-Api-Version")).ServeHTTP(rec, req)

		if want, have := "X-CSRF-Token, X-Api-Version", rec.Result().Header.Get("Access-Control-Allow-Headers"); want != have {
			t.Errorf("expected allowed headers %q, found %q", want, have)
		}

	})
}
func TestMiddleware(t *testing.T) {
	h := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Custom-Header", "value")
		rw.WriteHeader(http.StatusTeapot)
		_, _ = rw.Write([]byte("So much fun."))
	})

	for _, tc := range [...]struct {
		name         string
		options      cors.Options
		origin       string
		corsExpected bool
	}{
		{
			"Default allows some Origin",
			cors.Options{},
			"https://app.gesundheitscloud.de",
			true,
		},
		{
			"Default allows no Origin",
			cors.Options{},
			"",
			true,
		},
		{
			"Domain restriction allows listed",
			cors.Options{
				AllowedDomains: []string{"https://example.com"},
			},
			"https://example.com",
			true,
		},
		{
			"Domain restriction denies not listed",
			cors.Options{
				AllowedDomains: []string{"https://example.com"},
			},
			"https://gesundheitscloud.de",
			false,
		},
		{
			"Domain restriction denies not listed in multiple list",
			cors.Options{
				AllowedDomains: []string{"https://example.com", "http://hpsgc.de"},
			},
			"https://gesundheitscloud.de",
			false,
		},
		{
			"Domain restriction allows listed in multiple list",
			cors.Options{
				AllowedDomains: []string{"https://example.com", "http://hpsgc.de"},
			},
			"http://hpsgc.de",
			true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}

			rec := httptest.NewRecorder()

			corsHander := cors.New(tc.options)
			corsHander.MiddleWare(h).ServeHTTP(rec, req)

			res := rec.Result()

			corsFound := res.Header.Get("Access-Control-Allow-Origin") != ""

			if tc.corsExpected && !corsFound {
				t.Error("CORS header not found")
			}

			if !tc.corsExpected && corsFound {
				t.Error("unexpected CORS header found")
			}

			if want, have := 418, res.StatusCode; want != have {
				t.Errorf("expected status code %d, found %d", want, have)
			}
		})
	}
}
