package log

import (
	"context"
	"time"
)

// ErrInputOutput logs the input/output error message, along with context information.
// The expected context keys are "trace-id" and "user-id".
// The service must print this log line when a dependency does not respond or responds in an unexpected way.
func (l *Logger) ErrInputOutput(
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
		EventType:      "err-io",
		UserID:         userID,
		Message:        err.Error(),
		ClientID:       clientID,
		TenantID:       getFromContextWithDefault(ctx, TenantIDContextKey, l.tenantID),
	})
}
