package transport_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/transport"
)

var errFoo error = errors.New("some unrelated error")

type ProxyRoundTrip struct {
	statusCodes  []int
	currentIndex int
}

func (ppt *ProxyRoundTrip) RoundTrip(req *http.Request) (*http.Response, error) {
	if ppt.currentIndex >= len(ppt.statusCodes) {
		return nil, errFoo
	}

	rec := httptest.NewRecorder()
	rec.WriteHeader(ppt.statusCodes[ppt.currentIndex])
	ppt.currentIndex += 1

	return rec.Result(), nil
}

func TestRetrierTransport_RoundTrip(t *testing.T) {
	tests := []struct {
		name             string
		proxy            *ProxyRoundTrip
		successfulRetry  bool
		numberOfRequests int
	}{
		{
			"Happy path - no retry",
			&ProxyRoundTrip{[]int{http.StatusAccepted}, 0},
			true, 1,
		}, {
			"Happy path - with max retries",
			&ProxyRoundTrip{
				[]int{http.StatusTooManyRequests, http.StatusTooManyRequests, http.StatusTooManyRequests, http.StatusAccepted},
				0,
			},
			true, 4,
		}, {
			"Happy path - 3xx are also okay",
			&ProxyRoundTrip{[]int{http.StatusTooManyRequests, http.StatusPermanentRedirect}, 0},
			true, 2,
		}, {
			"Bad path - After max retries follows successful response",
			&ProxyRoundTrip{
				[]int{
					http.StatusTooManyRequests,
					http.StatusTooManyRequests,
					http.StatusTooManyRequests,
					http.StatusTooManyRequests,
					http.StatusAccepted,
				},
				0,
			},
			false, 4,
		}, {
			"Bad path - retry but only errors with capped retries",
			&ProxyRoundTrip{
				[]int{
					http.StatusTooManyRequests,
					http.StatusTooManyRequests,
					http.StatusTooManyRequests,
					http.StatusTooManyRequests,
					http.StatusTooManyRequests,
				},
				0,
			},
			false, 4,
		}, {
			"Bad path - instant failure",
			&ProxyRoundTrip{[]int{http.StatusUnavailableForLegalReasons, http.StatusAccepted}, 0},
			false, 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// runnig with default configuration
			retrier := transport.Retrier(3, time.Millisecond, nil)
			roundtripper := retrier(tt.proxy)

			req, err := http.NewRequest(http.MethodGet, "", nil)
			if err != nil {
				t.Fatalf("error creating http request: %v", err)
			}

			resp, err := roundtripper.RoundTrip(req)

			if tt.successfulRetry && (err != nil || resp.StatusCode < 200 || resp.StatusCode >= 400) {
				t.Fatalf("Expected no error and positive status code (200-399), got error: %v and status code: %d", err, resp.StatusCode)
			}

			if !tt.successfulRetry && (err != nil || resp.StatusCode < 400) {
				t.Fatalf("Expected no error and bad status code (>399), got error: %v and status code: %d", err, resp.StatusCode)
			}

			if tt.numberOfRequests != tt.proxy.currentIndex {
				t.Fatalf("Expected %v request, want: %v", tt.numberOfRequests, tt.proxy.currentIndex)
			}
		})
	}

	t.Run("Bad path - error in sending request", func(t *testing.T) {
		// runnig with default configuration
		retrier := transport.Retrier(3, time.Millisecond, nil)
		// will throw some error
		errorProxy := &ProxyRoundTrip{[]int{}, 1}
		roundtripper := retrier(errorProxy)

		req, err := http.NewRequest(http.MethodGet, "", nil)
		if err != nil {
			t.Fatalf("error creating http request: %v", err)
		}

		resp, err := roundtripper.RoundTrip(req)

		if err == nil || resp != nil {
			t.Fatalf("Expected no response and a non status code related error, got error %v and response %v", err, resp)
		}
	})
}
