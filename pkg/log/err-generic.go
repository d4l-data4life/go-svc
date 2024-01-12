package log

import (
	"context"
	"time"
)

// ErrGeneric logs the error message, along with context information.
// The expected context keys are "trace-id" and "user-id".
// This is the error type to use when a more specific type is not available.
func (l *Logger) ErrGeneric(
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
		UserID:         userID,
		Message:        err.Error(),
		ClientID:       clientID,
		TenantID:       getFromContextWithDefault(ctx, TenantIDContextKey, l.tenantID),
	})
}
