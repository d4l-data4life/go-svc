package middlewares_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/log"
	"github.com/gesundheitscloud/go-svc/pkg/middlewares"

	"github.com/stretchr/testify/assert"
)

func TestTraceNewID(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/route", nil)
	res := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Context().Value(log.TraceIDContextKey).(string)
		assert.NotEqual(t, "", traceID)
	})
	traceMiddleware := middlewares.Trace(handler)
	traceMiddleware.ServeHTTP(res, req)
}

func TestTraceExistingID(t *testing.T) {
	expectedTraceID := "b24caeb7-4250-428f-be62-844e041a5109"

	req, _ := http.NewRequest(http.MethodGet, "/route", nil)
	req.Header.Add(log.TraceIDHeaderKey, expectedTraceID)
	res := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Context().Value(log.TraceIDContextKey).(string)
		assert.Equal(t, expectedTraceID, traceID)
	})

	traceMiddleware := middlewares.Trace(handler)
	traceMiddleware.ServeHTTP(res, req)
}

func TestTraceTransportWithTraceIdInContext(t *testing.T) {
	expectedTraceID := "b24caeb7-4250-428f-be62-844e041a5109"
	expectedResponseBody := "Some content"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get(log.TraceIDHeaderKey)
		w.WriteHeader(200)
		_, err := w.Write([]byte(expectedResponseBody))
		assert.NoError(t, err)
		assert.Equal(t, expectedTraceID, traceID)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client := &http.Client{Transport: &middlewares.TraceTransport{}}

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	c := context.WithValue(req.Context(), log.TraceIDContextKey, expectedTraceID)
	req = req.WithContext(c)

	res, err := client.Do(req)
	assert.NoError(t, err)

	body, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, expectedResponseBody, string(body))
}

func TestTraceTransportWithOutTraceIdInContext(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get(log.TraceIDHeaderKey)
		w.WriteHeader(200)
		assert.NotEqual(t, "", traceID)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client := &http.Client{Transport: &middlewares.TraceTransport{}}

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)

	_, err := client.Do(req)
	assert.NoError(t, err)
}

func TestGenerateTraceID(t *testing.T) {
	traceID := middlewares.GenerateTraceID()
	assert.Equal(t, 32, len(traceID))
}
