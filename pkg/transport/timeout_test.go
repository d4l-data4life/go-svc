package transport_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/d4l-data4life/go-svc/pkg/transport"
)

type TimeoutCheckTransport struct{}

func (t *TimeoutCheckTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	_, ok := req.Context().Deadline()
	if !ok {
		return nil, fmt.Errorf("no deadline set")
	}

	return nil, nil
}

func TestTimeoutTransport(t *testing.T) {
	t.Parallel()

	rt := transport.Timeout(5 * time.Second)(&TimeoutCheckTransport{})

	req, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		t.Fatalf("error creating http request: %v", err)
	}

	_, err = rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("round trip failed: %v", err)
	}
}
