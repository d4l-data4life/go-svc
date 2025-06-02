package log

import (
	"bytes"
	"net/http"
	"time"
)

type inResponseLog struct {
	Timestamp       time.Time           `json:"timestamp,omitempty"`
	LogLevel        logLevel            `json:"log-level,omitempty"`
	TraceID         string              `json:"trace-id,omitempty"`
	ServiceName     string              `json:"service-name,omitempty"`
	ServiceVersion  string              `json:"service-version,omitempty"`
	Hostname        string              `json:"hostname,omitempty"`
	EventType       string              `json:"event-type,omitempty"`
	UserID          string              `json:"user-id,omitempty"`
	ReqMethod       string              `json:"req-method"`
	ReqURL          string              `json:"req-url"`
	ResponseCode    int                 `json:"response-code"`
	ResponseBody    string              `json:"response-body"`
	PayloadLength   int64               `json:"payload-length,omitempty"`
	Header          map[string][]string `json:"header,omitempty"`
	ContentType     string              `json:"content-type,omitempty"`
	ContentEncoding string              `json:"content-encoding,omitempty"`
	Duration        int64               `json:"roundtrip-duration,omitempty"`
	// OAuth client ID
	ClientID string `json:"client-id,omitempty"`
	// TenantID is the ID of the tenant to which the log belongs to
	TenantID string `json:"tenant-id,omitempty"`
}

func (h *HTTPLogger) httpInResponse(
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
		ServiceName:     h.log.serviceName,
		ServiceVersion:  h.log.serviceVersion,
		Hostname:        h.log.hostname,
		EventType:       "http-in-response",
		UserID:          userID,
		ReqMethod:       req.Method,
		ReqURL:          req.URL.String(),
		ResponseCode:    responseCode,
		ResponseBody:    bodyStr,
		PayloadLength:   payloadLength,
		Header:          hlcInResponse.processHeaders(responseHeader),
		ContentType:     responseHeader.Get("Content-Type"),
		ContentEncoding: responseHeader.Get("Content-Encoding"),
		Duration:        now.Sub(requestTimestamp).Milliseconds(),
		ClientID:        clientID,
		TenantID:        getFromContextWithDefault(req.Context(), TenantIDContextKey, h.log.tenantID),
	}

	log = h.obfuscateInResponse(log)

	return h.log.Log(log)
}

func (h *HTTPLogger) obfuscateInResponse(rlog inResponseLog) inResponseLog {
	obfKey := ObfuscatorKey(HTTPInResponse, rlog.ReqMethod)

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
