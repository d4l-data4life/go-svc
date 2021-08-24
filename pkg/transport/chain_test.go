package transport_test

import (
	"net/http"
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/transport"
)

type TestTransport struct {
	ch chan<- int
	i  int
	rt http.RoundTripper
}

func (t *TestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.ch <- t.i
	return t.rt.RoundTrip(req)
}

func NewTestTransport(ch chan<- int, i int) transport.TransportFunc {
	return func(rt http.RoundTripper) http.RoundTripper {
		return &TestTransport{
			ch: ch,
			i:  i,
			rt: rt,
		}
	}
}

type CloserTransport struct {
	ch chan<- int
}

func (t *CloserTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	close(t.ch)
	return nil, nil
}

func TestChainTransport(t *testing.T) {
	t.Parallel()

	ch := make(chan int, 5)

	rt := transport.Chain(
		NewTestTransport(ch, 0),
		NewTestTransport(ch, 1),
		NewTestTransport(ch, 2),
		NewTestTransport(ch, 3),
		NewTestTransport(ch, 4),
	)(&CloserTransport{ch: ch})

	_, err := rt.RoundTrip(nil)
	if err != nil {
		t.Fatalf("round trip failed: %v", err)
	}

	j := 0
	for i := range ch {
		if i != j {
			t.Fatalf("transports run in invalid order: %d != %d", i, j)
		}
		j++
	}
}
