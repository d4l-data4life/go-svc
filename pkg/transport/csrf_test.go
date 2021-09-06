package transport_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/transport"
)

type FakeTransport struct {
	res *http.Response
}

func (t *FakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.res, nil
}

func TestCSRFTransport(t *testing.T) {
	t.Parallel()

	rt := transport.CSRF("some-token")(&FakeTransport{
		res: &http.Response{
			Header: http.Header{
				"x-csrf-token": []string{"some-other-token"},
			},
		},
	})

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"",
		nil,
	)
	if err != nil {
		t.Fatalf("error creating http request: %v", err)
	}

	_, err = rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("round trip failed: %v", err)
	}

	if req.Header.Get("x-csrf-token") != "some-token" {
		t.Fatalf("invalid request csrf token")
	}
}
