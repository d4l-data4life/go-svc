package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
	"github.com/gesundheitscloud/go-svc/pkg/middlewares"
)

func call(ctx context.Context, URL, method, secret, userAgent string, payload *bytes.Buffer, expectedCodes ...int) ([]byte, int, error) {
	request, err := http.NewRequestWithContext(ctx, method, URL, payload)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error creating HTTP request")
		return nil, 0, err
	}
	request.Header.Add("Authorization", secret)
	request.Header.Set("User-Agent", userAgent)
	request.Close = true

	client := &http.Client{Transport: &middlewares.TraceTransport{}}
	response, err := client.Do(request)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error sending %s request to user-preferences service", method)
		return nil, 0, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if !existsIn(response.StatusCode, expectedCodes) {
		if err == nil {
			err = fmt.Errorf("method = '%s', URL = '%s' error: unexpected return code %d (wanted one of: %s), body = %s",
				method, URL, response.StatusCode, prettyPrint(expectedCodes), string(body))
		}
		logging.LogErrorfCtx(ctx, err, "error sending request to service. Status: %s", http.StatusText(response.StatusCode))
		return nil, response.StatusCode, err
	}
	return body, response.StatusCode, nil
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
