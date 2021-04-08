package log

import (
	"context"
	"net/http"
	"time"
)

type outRequestLog struct {
	Timestamp       time.Time `json:"timestamp"`
	LogLevel        logLevel  `json:"log-level"`
	TraceID         string    `json:"trace-id"`
	ServiceName     string    `json:"service-name"`
	ServiceVersion  string    `json:"service-version"`
	Hostname        string    `json:"hostname"`
	ReqMethod       string    `json:"req-method"`
	ReqURL          string    `json:"req-url"`
	EventType       string    `json:"event-type"`
	UserID          string    `json:"user-id,omitempty"`
	PayloadLength   int64     `json:"payload-length"`
	ReqBody         string    `json:"req-body"`
	ContentType     string    `json:"content-type"`
	ContentEncoding string    `json:"content-encoding"`
	// OAuth client ID
	ClientID string `json:"client-id,omitempty"`
	// TenantID is the ID of the tenant to which the log belongs to
	TenantID string `json:"tenant-id"`
}

func (l *Logger) HttpOutRequest(
	ctx context.Context,
	reqMethod string,
	reqURL string,
	payloadLength int64,
) error {
	traceID, userID, clientID := parseContext(ctx)

	return l.Log(outRequestLog{
		Timestamp:      time.Now(),
		LogLevel:       LevelInfo,
		TraceID:        traceID,
		ServiceName:    l.serviceName,
		ServiceVersion: l.serviceVersion,
		Hostname:       l.hostname,
		ReqMethod:      reqMethod,
		ReqURL:         reqURL,
		EventType:      "http-out-request",
		UserID:         userID,
		PayloadLength:  payloadLength,
		ClientID:       clientID,
		TenantID:       getFromContextWithDefault(ctx, TenantIDContextKey, l.tenantID),
	})
}

func (l *Logger) HttpOutReq(req *http.Request) error {
	traceID, userID, clientID := parseContext(req.Context())
	if traceID == "" {
		traceID = req.Header.Get(TraceIDHeaderKey)
	}
	bodyStr := filteredBodyStrFromReq(req)

	return l.Log(outRequestLog{
		Timestamp:       time.Now(),
		LogLevel:        LevelInfo,
		TraceID:         traceID,
		ServiceName:     l.serviceName,
		ServiceVersion:  l.serviceVersion,
		Hostname:        l.hostname,
		ReqMethod:       req.Method,
		ReqURL:          req.URL.String(),
		EventType:       "http-out-request",
		UserID:          userID,
		PayloadLength:   req.ContentLength,
		ReqBody:         bodyStr,
		ContentType:     req.Header.Get("Content-Type"),
		ContentEncoding: req.Header.Get("Content-Encoding"),
		ClientID:        clientID,
		TenantID:        getFromContextWithDefault(req.Context(), TenantIDContextKey, l.tenantID),
	})
}
