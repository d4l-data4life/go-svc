package transport_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/log"
	"github.com/gesundheitscloud/go-svc/pkg/transport"
)

type NopTransport struct{}

func (t *NopTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, nil
}

func TestTraceIDTransport(t *testing.T) {
	t.Parallel()

	rt := transport.TraceID(&NopTransport{})

	ctx := context.WithValue(
		context.Background(),
		log.TraceIDContextKey,
		"some-trace-id",
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "", nil)
	if err != nil {
		t.Fatalf("error creating http request: %v", err)
	}

	_, err = rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("round trip failed: %v", err)

	}

	if req.Header.Get(log.TraceIDHeaderKey) != "some-trace-id" {
		t.Fatalf("invalid trace-id")
	}
}
