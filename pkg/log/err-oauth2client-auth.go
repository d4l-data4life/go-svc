package log

import (
	"context"
	"time"
)

// ErrOauth2ClientAuth logs the oauth2 client authentication failure error message, along with context information.
// The expected context keys are "trace-id" and "user-id".
// The service must use this before responding to an unsuccessful authentication.
// Includes:
// credentials mismatch for OAUTH2 client authentication
func (l *Logger) ErrOauth2ClientAuth(
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
		EventType:      "err-oauth2client-auth",
		UserID:         userID,
		Message:        err.Error(),
		ClientID:       clientID,
		TenantID:       getFromContextWithDefault(ctx, TenantIDContextKey, l.tenantID),
	})
}
