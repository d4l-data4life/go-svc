package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
	"github.com/gesundheitscloud/go-svc/pkg/transport"
)

type caller struct {
	client *http.Client
	name   string
}

func NewCaller(timeout time.Duration, name string) *caller {
	return &caller{
		client: &http.Client{
			Transport: transport.Chain(
				transport.Timeout(timeout),
				transport.Prometheus(name),
				transport.Log(logging.Logger()),
				transport.TraceID,
				transport.JSON,
			)(nil),
		},
		name: name,
	}
}

func (c *caller) call(ctx context.Context, URL, method, secret, userAgent string, payload *bytes.Buffer, expectedCodes ...int) ([]byte, int, http.Header, error) {
	request, err := http.NewRequestWithContext(ctx, method, URL, payload)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error creating HTTP request")
		return nil, 0, nil, err
	}
	request.Header.Add("Authorization", secret)
	request.Header.Set("User-Agent", userAgent)
	request.Close = true

	response, err := c.client.Do(request)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error sending '%s' request to '%s'", method, URL)
		return nil, 0, nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if !existsIn(response.StatusCode, expectedCodes) {
		if err == nil {
			err = fmt.Errorf("method = '%s', URL = '%s' error: unexpected return code %d (wanted one of: %s), body = %s",
				method, URL, response.StatusCode, prettyPrint(expectedCodes), string(body))
		}
		logging.LogErrorfCtx(ctx, err, "error sending request to service. Status: %s", http.StatusText(response.StatusCode))
		return nil, response.StatusCode, nil, err
	}
	return body, response.StatusCode, response.Header, nil
}

func existsIn(value int, array []int) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}

func prettyPrint(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		fmt.Println("error:", err)
	}
	return string(b)
}
