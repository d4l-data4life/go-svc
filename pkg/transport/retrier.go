package transport

import (
	"fmt"
	"net/http"
	"time"

	"github.com/eapache/go-resiliency/retrier"
)

type StatusCode int

func (e StatusCode) Error() string {
	return fmt.Sprintf("HTTP Status code of %d", int(e))
}

type RetrierTransport struct {
	allowedStatusCodes map[StatusCode]bool
	retries            int
	retrySleep         time.Duration
	rt                 http.RoundTripper
}

func (rt *RetrierTransport) Classify(err error) retrier.Action {
	// can happen that the request itself is erroneous, therefore having a different error type
	if responseStatusCode, ok := err.(StatusCode); ok {
		// 2xx and 3xx are okay
		if responseStatusCode >= 200 && responseStatusCode < 400 {
			return retrier.Succeed
		}

		if _, ok := rt.allowedStatusCodes[responseStatusCode]; ok {
			return retrier.Retry
		}
	}

	// no retry defined
	return retrier.Fail
}

func (rt *RetrierTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	retry := retrier.New(retrier.ExponentialBackoff(rt.retries, rt.retrySleep), rt)
	var response *http.Response

	err := retry.Run(func() error {
		resp, err := rt.rt.RoundTrip(req)
		if err == nil {
			response = resp
			return StatusCode(resp.StatusCode)
		} else {
			return err
		}
	})

	// roundtrip must return a response if the error is nil even if we retried
	if _, ok := err.(StatusCode); ok {
		return response, nil
	} else {
		return response, err
	}
}

// Retrier will retry requesting a resource depending on some http status code. Generally, status codes from 200 to 399 are always viewed as a success
func Retrier(retries int, retrySleep time.Duration, retriableStatusCodes []int) TransportFunc {
	return func(rt http.RoundTripper) http.RoundTripper {
		if rt == nil {
			rt = http.DefaultTransport
		}

		var statusCodeSet map[StatusCode]bool

		if retriableStatusCodes == nil {
			// based on https://developer.mozilla.org/en-US/docs/Web/HTTP/Status#server_error_responses
			statusCodeSet = map[StatusCode]bool{
				// 4xx
				http.StatusRequestTimeout:  true,
				http.StatusTooManyRequests: true,
				// 5xx
				http.StatusInternalServerError: true,
				http.StatusBadGateway:          true,
				http.StatusServiceUnavailable:  true,
				http.StatusGatewayTimeout:      true,
			}
		} else {
			statusCodeSet = make(map[StatusCode]bool)
			for sc := range retriableStatusCodes {
				statusCodeSet[StatusCode(sc)] = true
			}
		}

		return &RetrierTransport{
			allowedStatusCodes: statusCodeSet,
			retries:            retries,
			retrySleep:         retrySleep,
			rt:                 rt,
		}
	}
}
