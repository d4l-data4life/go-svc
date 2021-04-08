package log

import (
	"context"
	"time"
)

// ErrInternal logs the internal error message, along with context information.
// The expected context keys are "trace-id" and "user-id".
// The service must print this log line when an error arises while processing sanitised inputs.
// Includes:
// generating new IDs
// parsing data that has already been processed (e.g. error parsing data coming from the DB)
func (l *Logger) ErrInternal(
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
		EventType:      "err-internal",
		UserID:         userID,
		Message:        err.Error(),
		ClientID:       clientID,
		TenantID:       getFromContextWithDefault(ctx, TenantIDContextKey, l.tenantID),
	})
}
