package log

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type AuditLogType string

const (
	AccessLog   AuditLogType = "access"
	ChangeLog   AuditLogType = "change"
	SecurityLog AuditLogType = "security"
)

type baseAuditLog struct {
	// Log Data
	Timestamp    time.Time    `json:"timestamp"`
	LogType      LogType      `json:"log-type"`
	AuditLogType AuditLogType `json:"audit-log-type"`

	// Channel Data
	TraceID        string `json:"trace-id,omitempty"`
	ServiceName    string `json:"service-name,omitempty"`
	ServiceVersion string `json:"service-version,omitempty"`
	Hostname       string `json:"hostname,omitempty"`
	PodName        string `json:"pod-name,omitempty"`
	Environment    string `json:"environment,omitempty"`
	RequestURL     string `json:"req-url,omitempty"`
	RequestDomain  string `json:"req-domain,omitempty"`

	// OAuth client ID
	ClientID string `json:"client-id,omitempty"`

	// TenantID is the ID of the tenant to which the log belongs to
	TenantID string `json:"tenant-id,omitempty"`

	// The IP address of the caller
	CallerIPAddress string `json:"caller-ip,omitempty"`

	// subject is the user initiating the event (the `sub` claim)
	SubjectID string `json:"subject-id,omitempty"`
}

func (l *Logger) createBaseAuditLog(ctx context.Context, logType AuditLogType) baseAuditLog {
	return baseAuditLog{
		Timestamp:       time.Now(),
		LogType:         Audit,
		AuditLogType:    logType,
		TraceID:         getFromContext(ctx, TraceIDContextKey),
		ServiceName:     l.serviceName,
		ServiceVersion:  l.serviceVersion,
		Hostname:        l.hostname,
		PodName:         l.podName,
		Environment:     l.environment,
		ClientID:        getFromContext(ctx, ClientIDContextKey),
		RequestURL:      getFromContext(ctx, RequestURLContextKey),
		RequestDomain:   getFromContext(ctx, RequestDomainContextKey),
		CallerIPAddress: getFromContext(ctx, CallerIPContextKey),
		SubjectID:       getFromContext(ctx, UserIDContextKey),
		TenantID:        getFromContextWithDefault(ctx, TenantIDContextKey, l.tenantID),
	}
}

// Audit logs the audit message, along with an audit object.
// The expected context keys are "trace-id" and "user-id".
// This is the log type to use when a message should be accompanied
// with an object relevant for auditing, e.g., new set of permissions.
// Deprecated: use AuditSecurity instead.
func (l *Logger) Audit(
	ctx context.Context,
	message string,
	object interface{},
) error {
	traceID, userID, clientID := parseContext(ctx)
	objectString := ""

	if object != nil && object != struct{}{} {
		marshaledObject, err := json.Marshal(object)
		if err != nil {
			return fmt.Errorf("cannot marshal audited object '%v' to JSON: %w", object, err)
		}
		objectString = string(marshaledObject)
	}

	return l.Log(logEntry{
		Timestamp:      time.Now(),
		LogLevel:       LevelAudit,
		TraceID:        traceID,
		ServiceName:    l.serviceName,
		ServiceVersion: l.serviceVersion,
		Hostname:       l.hostname,
		EventType:      "audit",
		UserID:         userID,
		Message:        message,
		Object:         objectString,
		ClientID:       clientID,
		TenantID:       getFromContextWithDefault(ctx, TenantIDContextKey, l.tenantID),
	})
}

type ExtraAuditInfoProvider func(interface{})

// SubjectID allows to override the default value for subject ID (which is
// taken from the context).
// It can be used on any audit method.
func SubjectID(sID string) ExtraAuditInfoProvider {
	return func(l interface{}) {
		switch l := l.(type) {
		case *changeLog:
			l.SubjectID = sID
		case *securityLog:
			l.SubjectID = sID
		case *accessLog:
			l.SubjectID = sID
		}
	}
}

// ClientID allows to override the default value for client ID (which is
// taken from the context).
// It can be used on any audit method.
func ClientID(cID string) ExtraAuditInfoProvider {
	return func(l interface{}) {
		switch l := l.(type) {
		case *changeLog:
			l.ClientID = cID
		case *securityLog:
			l.ClientID = cID
		case *accessLog:
			l.ClientID = cID
		}
	}
}

// AdditionalData allows to add some additional information to an audit log.
// It can be used on any audit method.
func AdditionalData(data interface{}) ExtraAuditInfoProvider {
	return func(l interface{}) {
		switch l := l.(type) {
		case *changeLog:
			l.AdditionalData = data
		case *securityLog:
			l.AdditionalData = data
		case *accessLog:
			l.AdditionalData = data
		}
	}
}

// Message allows to add a text message to a security log.
// It has only effect when used on a security event method,
// i.e. `AuditSecurity`, AuditSecuritySuccess` or `AuditSecurityFailure`
// methods.
func Message(m string) ExtraAuditInfoProvider {
	return func(l interface{}) {
		if l, ok := l.(*securityLog); ok {
			l.Message = m
		}
	}
}

// Message allows to add the old value to a change log event.
// This is useful when the old value is available and can be logged.
// Typically an old value doesn't make sense on a create operation
// but this case is supported as well.
// It has only effect when used on a change log method,
// i.e. `AuditCreate`, AuditUpdate` or `AuditDelete` methods.
func OldValue(v interface{}) ExtraAuditInfoProvider {
	return func(l interface{}) {
		if l, ok := l.(*changeLog); ok {
			l.OldValue = v
		}
	}
}
