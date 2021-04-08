package log

import (
	"context"
	"fmt"
	"time"
)

type securityLog struct {
	baseAuditLog

	// Event Data
	SecurityEvent  string      `json:"security-event"`
	Successful     bool        `json:"successful"`
	Message        string      `json:"message,omitempty"`
	AdditionalData interface{} `json:"additional-data,omitempty"`
}

// AuditSecurity should be used to log any security-relevant event
// It will attempt to get `TraceID`, `ClientID`, `RequestURL`, `RequestDomain`, `CallerIPAddress` and `SubjectID`
// from the context.
// `securityEvent` should be the category of the event
// `successful` should be true is the event occurred successfully or false if it was rejected/unsuccessful
// `extras` allows to add optional information or override default values:
//    - `SubjectID` allows to override the ID of the user performing the action (by default it is expected in the context)
//    - `ClientID` allows to override the oauth client ID (by default it is expected in the context)
//    - `Message` allows to provide an instance-specific more detailed message about the event
//    - `AdditionalData` allows to provide any extra information relevant for the audit log
func (l *Logger) AuditSecurity(
	ctx context.Context,
	securityEvent fmt.Stringer,
	successful bool,
	extras ...ExtraAuditInfoProvider,
) error {
	log := securityLog{
		baseAuditLog: baseAuditLog{
			Timestamp:       time.Now(),
			LogType:         Audit,
			AuditLogType:    SecurityLog,
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
		},
		SecurityEvent: securityEvent.String(),
		Successful:    successful,
	}

	for _, f := range extras {
		f(&log)
	}

	return l.Log(log)
}

// AuditSecuritySuccess should be used to log any successful security-relevant event
// It will attempt to get `TraceID`, `ClientID`, `RequestURL`, `RequestDomain`, `CallerIPAddress` and `SubjectID`
// from the context.
// `securityEvent` should be the category of the event
// `extras` allows to add optional information or override default values:
//    - `SubjectID` allows to override the ID of the user performing the action (by default it is expected in the context)
//    - `ClientID` allows to override the oauth client ID (by default it is expected in the context)
//    - `Message` allows to provide an instance-specific more detailed message about the event
//    - `AdditionalData` allows to provide any extra information relevant for the audit log
func (l *Logger) AuditSecuritySuccess(
	ctx context.Context,
	securityEvent fmt.Stringer,
	extras ...ExtraAuditInfoProvider,
) error {
	return l.AuditSecurity(
		ctx,
		securityEvent,
		true,
		extras...,
	)
}

// AuditSecurityFailure should be used to log any unsuccessful security-relevant event
// It will attempt to get `TraceID`, `ClientID`, `RequestURL`, `RequestDomain`, `CallerIPAddress` and `SubjectID`
// from the context.
// `securityEvent` should be the category of the event
// `extras` allows to add optional information or override default values:
//    - `SubjectID` allows to override the ID of the user performing the action (by default it is expected in the context)
//    - `ClientID` allows to override the oauth client ID (by default it is expected in the context)
//    - `Message` allows to provide an instance-specific more detailed message about the event
//    - `AdditionalData` allows to provide any extra information relevant for the audit log
func (l *Logger) AuditSecurityFailure(
	ctx context.Context,
	securityEvent fmt.Stringer,
	extras ...ExtraAuditInfoProvider,
) error {
	return l.AuditSecurity(
		ctx,
		securityEvent,
		false,
		extras...,
	)
}
