package log

import (
	"bytes"
	"net/http"
	"time"
)

type inResponseLog struct {
	Timestamp       time.Time `json:"timestamp"`
	LogLevel        logLevel  `json:"log-level"`
	TraceID         string    `json:"trace-id"`
	ServiceName     string    `json:"service-name"`
	ServiceVersion  string    `json:"service-version"`
	Hostname        string    `json:"hostname"`
	EventType       string    `json:"event-type"`
	UserID          string    `json:"user-id,omitempty"`
	ReqMethod       string    `json:"req-method"`
	ReqURL          string    `json:"req-url"`
	ResponseCode    int       `json:"response-code"`
	ResponseBody    string    `json:"response-body"`
	PayloadLength   int64     `json:"payload-length"`
	ContentType     string    `json:"content-type"`
	ContentEncoding string    `json:"content-encoding"`
	Duration        int64     `json:"roundtrip-duration"`
	// OAuth client ID
	ClientID string `json:"client-id,omitempty"`
	// TenantID is the ID of the tenant to which the log belongs to
	TenantID string `json:"tenant-id"`
}

func (l *HTTPLogger) httpInResponse(
	req *http.Request,
	responseHeader http.Header,
	responseCode int,
	responseBody *bytes.Buffer,
	payloadLength int64,
	requestTimestamp time.Time,
) error {
	traceID, userID, clientID := parseContext(req.Context())

	bodyStr := filteredBodyStrFromBuffer(responseBody, responseHeader)

	level := LevelInfo
	if responseCode >= http.StatusBadRequest {
		level = LevelError
	}

	now := time.Now()

	log := inResponseLog{
		Timestamp:       now,
		LogLevel:        level,
		TraceID:         traceID,
		ServiceName:     l.log.serviceName,
		ServiceVersion:  l.log.serviceVersion,
		Hostname:        l.log.hostname,
		EventType:       "http-in-response",
		UserID:          userID,
		ReqMethod:       req.Method,
		ReqURL:          req.URL.String(),
		ResponseCode:    responseCode,
		ResponseBody:    bodyStr,
		PayloadLength:   payloadLength,
		ContentType:     responseHeader.Get("Content-Type"),
		ContentEncoding: responseHeader.Get("Content-Encoding"),
		Duration:        now.Sub(requestTimestamp).Milliseconds(),
		ClientID:        clientID,
		TenantID:        getFromContextWithDefault(req.Context(), TenantIDContextKey, l.log.tenantID),
	}

	log = l.obfuscateInResponse(log)

	return l.log.Log(log)
}

func (h *HTTPLogger) obfuscateInResponse(rlog inResponseLog) inResponseLog {
	obfKey := h.obfuscatorKey(HTTPInResponse, rlog.ReqMethod)

	obf := h.obf[obfKey]

	for _, o := range obf {
		rlog = o.Obfuscate(rlog).(inResponseLog)
	}

	return rlog
}

const matchAll = ".*"

func (o *Obfuscator) obfuscateInResponse(rlog inResponseLog) inResponseLog {
	if o.ReqURL != nil && o.ReqURL.String() != matchAll && !o.ReqURL.MatchString(rlog.ReqURL) {
		return rlog
	}

	if o.Field == Body {
		rlog.ResponseBody = o.Replace.ReplaceAllString(rlog.ResponseBody, o.With)
	}
	return rlog
}
