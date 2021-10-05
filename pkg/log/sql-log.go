package log

import (
	"context"
	"errors"
	"time"
)

type SqlLogData struct {
	Action   string
	Duration time.Duration
	Error    error
	Sql      string
	Args     string
}

func (l *Logger) SqlLog(
	ctx context.Context,
	sqlLogData SqlLogData,
) error {
	traceID, userID, clientID := parseContext(ctx)

	if sqlLogData.Error == nil {
		sqlLogData.Error = errors.New("")
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
		Action:         sqlLogData.Action,
		Duration:       sqlLogData.Duration.Milliseconds(),
		Error:          sqlLogData.Error.Error(),
		Sql:            sqlLogData.Sql,
		Args:           sqlLogData.Args,
	})
}
