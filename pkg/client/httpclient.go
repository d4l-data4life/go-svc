package client

import (
	"context"
	"net/http"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/log"
	"github.com/gesundheitscloud/go-svc/pkg/prom"
)

const (
	DefaultTimeout = time.Second * 30
)

type Client struct {
	httpClient *http.Client
	timeout    time.Duration
}

func (c *Client) DoWithContext(ctx context.Context, req *http.Request) (*http.Response, error) {
	traceID, _ := ctx.Value(log.TraceIDContextKey).(string)
	req.Header.Set(log.TraceIDHeaderKey, traceID)

	ctxWTimeout, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req = req.WithContext(ctxWTimeout)

	return c.httpClient.Do(req)
}

func NewInstrumentedClient(name string, timeoutSec int, logger *log.Logger) *Client {
	return &Client{
		httpClient: newInstrumentedHTTPClient(name, logger),
		timeout:    timeout(timeoutSec),
	}
}

func NewInstrumentedClientNoRedirect(name string, timeoutSec int, logger *log.Logger) *Client {
	client := newInstrumentedHTTPClient(name, logger)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		// we do not want to follow any redirects but return them
		// back to the client instead
		return http.ErrUseLastResponse
	}
	return &Client{
		httpClient: client,
		timeout:    timeout(timeoutSec),
	}
}

func timeout(timeoutSec int) time.Duration {
	if timeoutSec <= 0 {
		return DefaultTimeout
	}
	return time.Second * time.Duration(timeoutSec)
}

func newInstrumentedHTTPClient(name string, logger *log.Logger) *http.Client {
	monitored := prom.NewRoundTripperInstrumenter().Instrument(name, http.DefaultTransport)
	loggedMonitored := logger.LoggedTransport(monitored)

	return &http.Client{
		Transport: loggedMonitored,
	}
}
