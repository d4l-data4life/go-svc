package log

import (
	"context"
	"time"
)

// WarnGeneric logs the warning message, along with error and context information.
// The expected context keys are "trace-id" and "user-id".
func (l *Logger) WarnGeneric(
	ctx context.Context,
	message string,
	err error,
) error {
	traceID, userID, clientID := parseContext(ctx)
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	}

	return l.Log(logEntry{
		Timestamp:      time.Now(),
		LogLevel:       LevelWarning,
		TraceID:        traceID,
		ServiceName:    l.serviceName,
		ServiceVersion: l.serviceVersion,
		Hostname:       l.hostname,
		EventType:      "warning-generic",
		UserID:         userID,
		Error:          errorMsg,
		Message:        message,
		ClientID:       clientID,
		TenantID:       getFromContextWithDefault(ctx, TenantIDContextKey, l.tenantID),
	})
}
