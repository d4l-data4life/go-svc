package log

import (
	"context"
	"time"
)

// InfoGeneric logs the info message, along with context information.
// The expected context keys are "trace-id" and "user-id".
// This is the log type to use for info messages when a more specific type is not available.
func (l *Logger) InfoGeneric(
	ctx context.Context,
	message string,
) error {
	traceID, userID, clientID := parseContext(ctx)

	return l.Log(logEntry{
		Timestamp:      time.Now(),
		LogLevel:       LevelInfo,
		TraceID:        traceID,
		ServiceName:    l.serviceName,
		ServiceVersion: l.serviceVersion,
		Hostname:       l.hostname,
		UserID:         userID,
		Message:        message,
		ClientID:       clientID,
		TenantID:       getFromContextWithDefault(ctx, TenantIDContextKey, l.tenantID),
	})
}
