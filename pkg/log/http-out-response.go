package log

import (
	"net/http"
	"time"
)

var hlcOutResponse *headerObfuscator = newHeaderObfuscator().
	obfuscateHeaders([]string{"Authorization", "WWW-Authenticate"}).
	ignoreHeaders([]string{"X-Real-Ip", "Content-Encoding", "Content-Type", "Accept-Encoding", "Content-Length", "Date"})

type outResponseLog struct {
	Timestamp       time.Time           `json:"timestamp"`
	LogLevel        logLevel            `json:"log-level"`
	TraceID         string              `json:"trace-id"`
	ServiceName     string              `json:"service-name"`
	ServiceVersion  string              `json:"service-version"`
	Hostname        string              `json:"hostname"`
	ReqMethod       string              `json:"req-method"`
	ReqURL          string              `json:"req-url"`
	EventType       string              `json:"event-type"`
	UserID          string              `json:"user-id,omitempty"`
	ResponseCode    int                 `json:"response-code"`
	ResponseBody    string              `json:"response-body"`
	PayloadLength   int64               `json:"payload-length"`
	Header          map[string][]string `json:"header"`
	ContentType     string              `json:"content-type"`
	ContentEncoding string              `json:"content-encoding"`
	Duration        int64               `json:"roundtrip-duration"`
	// OAuth client ID
	ClientID string `json:"client-id,omitempty"`
	// TenantID is the ID of the tenant to which the log belongs to
	TenantID string `json:"tenant-id"`
}

func (l *Logger) HttpOutResponse(
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
		Header:          hlcOutResponse.processHeaders(resp.Header),
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

	switch o.Field {
	case Body:
		rlog.ResponseBody = o.Replace.ReplaceAllString(rlog.ResponseBody, o.With)
	}
	return rlog
}
