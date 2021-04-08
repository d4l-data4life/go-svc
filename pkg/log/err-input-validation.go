package log

import (
	"context"
	"time"
)

// ErrInputValidation logs the error message, along with context information.
// The expected context keys are "trace-id" and "user-id".
// The service must print this log line before responding upon a malformed or unexpected input.
func (l *Logger) ErrInputValidation(
	ctx context.Context,
	err error,
) error {
	traceID, userID, clientID := parseContext(ctx)

	return l.Log(logEntry{
		Timestamp:      time.Now(),
		LogLevel:       LevelError,
		TraceID:        traceID,
		ServiceName:    l.serviceName,
		ServiceVersion: l.serviceVersion,
		Hostname:       l.hostname,
		EventType:      "err-input-validation",
		UserID:         userID,
		Message:        err.Error(),
		ClientID:       clientID,
		TenantID:       getFromContextWithDefault(ctx, TenantIDContextKey, l.tenantID),
	})
}
