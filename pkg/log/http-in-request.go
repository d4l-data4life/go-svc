package log

import (
	"fmt"
	"net/http"
	"time"
)

type inRequestLog struct {
	Timestamp       time.Time           `json:"timestamp,omitempty"`
	LogLevel        logLevel            `json:"log-level,omitempty"`
	TraceID         string              `json:"trace-id,omitempty"`
	ServiceName     string              `json:"service-name,omitempty"`
	ServiceVersion  string              `json:"service-version,omitempty"`
	Hostname        string              `json:"hostname,omitempty"`
	ReqIP           string              `json:"req-ip,omitempty"`
	ReqMethod       string              `json:"req-method"`
	ReqBody         string              `json:"req-body"`
	ReqForm         string              `json:"req-form,omitempty"`
	ReqURL          string              `json:"req-url"`
	RealIP          string              `json:"real-ip,omitempty"`
	EventType       string              `json:"event-type,omitempty"`
	UserID          string              `json:"user-id,omitempty"`
	PayloadLength   int64               `json:"payload-length,omitempty"`
	Header          map[string][]string `json:"header,omitempty"`
	ContentType     string              `json:"content-type,omitempty"`
	ContentEncoding string              `json:"content-encoding,omitempty"`
	// OAuth client ID
	ClientID string `json:"client-id,omitempty"`
	// TenantID is the ID of the tenant to which the log belongs to
	TenantID string `json:"tenant-id,omitempty"`
}

func (h *HTTPLogger) httpInRequest(req *http.Request) error {
	traceID, userID, clientID := parseContext(req.Context())
	_ = req.ParseForm()

	bodyStr := filteredBodyStrFromReq(req)

	log := inRequestLog{
		Timestamp:       time.Now(),
		LogLevel:        LevelInfo,
		TraceID:         traceID,
		ServiceName:     h.log.serviceName,
		ServiceVersion:  h.log.serviceVersion,
		Hostname:        h.log.hostname,
		ReqIP:           req.RemoteAddr,
		ReqMethod:       req.Method,
		ReqBody:         bodyStr,
		ReqForm:         fmt.Sprintf("%s", req.Form),
		ReqURL:          req.URL.String(),
		RealIP:          req.Header.Get("X-Real-Ip"),
		EventType:       "http-in-request",
		UserID:          userID,
		PayloadLength:   req.ContentLength,
		Header:          hlcInRequest.processHeaders(req.Header),
		ContentType:     req.Header.Get("Content-Type"),
		ContentEncoding: req.Header.Get("Content-Encoding"),
		ClientID:        clientID,
		TenantID:        getFromContextWithDefault(req.Context(), TenantIDContextKey, h.log.tenantID),
	}

	log = h.obfuscateInRequest(log)
	log = h.anonymizeIP(log)

	return h.log.Log(log)
}

func (h *HTTPLogger) anonymizeIP(rlog inRequestLog) inRequestLog {
	for _, a := range h.ipa {
		rlog = a.anonymizeIPInRequest(rlog)
	}
	return rlog
}

func (a *IPAnonymizer) anonymizeIPInRequest(rlog inRequestLog) inRequestLog {
	switch a.IPType {
	case IPTypeReal:
		rlog.RealIP = a.With
	case IPTypeReq:
		rlog.ReqIP = a.With
	case IPTypeAll:
		rlog.ReqIP = a.With
		rlog.RealIP = a.With
	}
	return rlog
}

func (h *HTTPLogger) obfuscateInRequest(rlog inRequestLog) inRequestLog {
	obfKey := ObfuscatorKey(HTTPInRequest, rlog.ReqMethod)

	obf := h.obf[obfKey]

	for _, o := range obf {
		rlog = o.Obfuscate(rlog).(inRequestLog)
	}

	return rlog
}

func (o *Obfuscator) obfuscateInRequest(rlog inRequestLog) inRequestLog {
	if o.ReqURL != nil && o.ReqURL.String() != matchAll && !o.ReqURL.MatchString(rlog.ReqURL) {
		return rlog
	}

	switch o.Field {
	case Body:
		rlog.ReqBody = o.Replace.ReplaceAllString(rlog.ReqBody, o.With)
	case ReqForm:
		rlog.ReqForm = o.Replace.ReplaceAllString(rlog.ReqForm, o.With)
	}
	return rlog
}
