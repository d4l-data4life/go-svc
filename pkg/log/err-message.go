package log

import (
	"context"
	"time"
)

// ErrMessage logs the error message, along with additional message and context information.
// The expected context keys are "trace-id" and "user-id".
func (l *Logger) ErrMessage(
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
		LogLevel:       LevelError,
		TraceID:        traceID,
		ServiceName:    l.serviceName,
		ServiceVersion: l.serviceVersion,
		Hostname:       l.hostname,
		EventType:      "err-message",
		UserID:         userID,
		Message:        message,
		Error:          errorMsg,
		ClientID:       clientID,
		TenantID:       getFromContextWithDefault(ctx, TenantIDContextKey, l.tenantID),
	})
}
