package standard

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithPathPrefixStrip(t *testing.T) {
	t.Run("strips matching prefix", func(t *testing.T) {
		var got string
		h := WithPathPrefixStrip("/api/pillars", http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			got = r.URL.Path
		}))

		req := httptest.NewRequest(http.MethodPost, "http://example/api/pillars/proto.api.Users/GetSelf", nil)
		h.ServeHTTP(httptest.NewRecorder(), req)

		if got != "/proto.api.Users/GetSelf" {
			t.Fatalf("expected stripped path, got %q", got)
		}
	})

	t.Run("keeps non-matching path unchanged", func(t *testing.T) {
		var got string
		h := WithPathPrefixStrip("/api/pillars", http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			got = r.URL.Path
		}))

		req := httptest.NewRequest(http.MethodPost, "http://example/proto.api.Users/GetSelf", nil)
		h.ServeHTTP(httptest.NewRecorder(), req)

		if got != "/proto.api.Users/GetSelf" {
			t.Fatalf("expected unchanged path, got %q", got)
		}
	})

	t.Run("maps exact prefix to root", func(t *testing.T) {
		var got string
		h := WithPathPrefixStrip("/api/pillars", http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			got = r.URL.Path
		}))

		req := httptest.NewRequest(http.MethodGet, "http://example/api/pillars", nil)
		h.ServeHTTP(httptest.NewRecorder(), req)

		if got != "/" {
			t.Fatalf("expected root path, got %q", got)
		}
	})
}
