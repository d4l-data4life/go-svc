package log

import (
	"context"
	"time"
)

// ErrUserAuth logs the user authentication failure error message, along with context information.
// The expected context keys are "trace-id" and "user-id".
// The service must use this log line before responding to an unsuccessful authentication.
// Includes:
// credentials mismatch for user authentication
func (l *Logger) ErrUserAuth(
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
		EventType:      "err-user-auth",
		UserID:         userID,
		Message:        err.Error(),
		ClientID:       clientID,
		TenantID:       getFromContextWithDefault(ctx, TenantIDContextKey, l.tenantID),
	})
}
