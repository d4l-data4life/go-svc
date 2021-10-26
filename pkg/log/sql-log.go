package log

import (
	"context"
	"fmt"
	"time"
)

func (l *Logger) SqlLog(
	ctx context.Context,
	pgxLogLevel string,
	msg string,
	data map[string]interface{},
) error {
	traceID, userID, clientID := parseContext(ctx)

	var dataMessage string

	if len(data) != 0 {
		dataMessage = fmt.Sprintf("%+v", data)
	}

	return l.Log(sqlLogEntry{
		Timestamp:      time.Now(),
		LogLevel:       LevelInfo,
		TraceID:        traceID,
		ServiceName:    l.serviceName,
		ServiceVersion: l.serviceVersion,
		Hostname:       l.hostname,
		EventType:      "sql-log",
		UserID:         userID,
		ClientID:       clientID,
		TenantID:       getFromContextWithDefault(ctx, TenantIDContextKey, l.tenantID),
		PgxLogLevel:    pgxLogLevel,
		PgxMessage:     msg,
		PgxData:        dataMessage,
	})
}
