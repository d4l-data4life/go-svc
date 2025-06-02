package log

import (
	"net/http"
	"time"
)

type outResponseLog struct {
	Timestamp       time.Time           `json:"timestamp,omitempty"`
	LogLevel        logLevel            `json:"log-level,omitempty"`
	TraceID         string              `json:"trace-id,omitempty"`
	ServiceName     string              `json:"service-name,omitempty"`
	ServiceVersion  string              `json:"service-version,omitempty"`
	Hostname        string              `json:"hostname,omitempty"`
	ReqMethod       string              `json:"req-method"`
	ReqURL          string              `json:"req-url"`
	EventType       string              `json:"event-type,omitempty"`
	UserID          string              `json:"user-id,omitempty"`
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

func (l *Logger) HTTPOutResponse(
	req *http.Request,
	resp *http.Response,
	requestTimestamp time.Time,
	obf map[string][]HTTPObfuscator,
) error {
	traceID, userID, clientID := parseContext(req.Context())
	if traceID == "" {
		traceID = req.Header.Get(TraceIDHeaderKey)
	}

	var bodyStr, ct, ce string
	var code int
	var cl int64
	if resp != nil {
		bodyStr = filteredBodyStrFromResp(resp)
		ct = resp.Header.Get("Content-Type")
		ce = resp.Header.Get("Content-Encoding")
		code = resp.StatusCode
		cl = resp.ContentLength
	}

	now := time.Now()
	level := LevelInfo
	if code >= http.StatusBadRequest {
		level = LevelError
	}

	var header map[string][]string
	if resp != nil {
		header = hlcOutResponse.processHeaders(resp.Header)
	}

	outLog := outResponseLog{
		Timestamp:       now,
		LogLevel:        level,
		TraceID:         traceID,
		ServiceName:     l.serviceName,
		ServiceVersion:  l.serviceVersion,
		Hostname:        l.hostname,
		ReqMethod:       req.Method,
		ReqURL:          req.URL.String(),
		EventType:       HTTPOutResponse.String(),
		UserID:          userID,
		ResponseCode:    code,
		ResponseBody:    bodyStr,
		PayloadLength:   cl,
		ContentType:     ct,
		ContentEncoding: ce,
		Header:          header,
		Duration:        now.Sub(requestTimestamp).Milliseconds(),
		ClientID:        clientID,
		TenantID:        getFromContextWithDefault(req.Context(), TenantIDContextKey, l.tenantID),
	}

	o := obf[ObfuscatorKey(HTTPOutRequest, req.Method)]
	for _, obfuscator := range o {
		outLog = obfuscator.Obfuscate(outLog).(outResponseLog)
	}

	return l.Log(outLog)
}

func (o *Obfuscator) obfuscateOutResponse(rlog outResponseLog) outResponseLog {
	if o.ReqURL != nil && o.ReqURL.String() != matchAll && !o.ReqURL.MatchString(rlog.ReqURL) {
		return rlog
	}

	if o.Field == Body {
		rlog.ResponseBody = o.Replace.ReplaceAllString(rlog.ResponseBody, o.With)
	}
	return rlog
}
