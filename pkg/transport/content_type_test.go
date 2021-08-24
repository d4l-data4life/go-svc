package transport_test

import (
	"net/http"
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/transport"
)

func TestJSONTransport(t *testing.T) {
	t.Parallel()

	rt := transport.JSON(&NopTransport{})

	req, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		t.Fatalf("error creating http request: %v", err)
	}

	_, err = rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("round trip failed: %v", err)
	}

	if req.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("content-type header != application/json")
	}
}
