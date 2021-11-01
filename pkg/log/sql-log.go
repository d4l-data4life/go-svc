package log

import (
	"context"
	"fmt"
	"time"
)

// SqlObfuscator applies obfuscation to an SQL Log
type SqlObfuscator interface {
	Obfuscate(interface{}) interface{}
}

func (l *Logger) SqlLog(
	ctx context.Context,
	pgxLogLevel string,
	msg string,
	data map[string]interface{},
	obfuscators ...SqlObfuscator,
) error {
	traceID, userID, clientID := parseContext(ctx)

	var dataMessage string

	if len(data) != 0 {
		dataMessage = fmt.Sprintf("%+v", data)
	}

	log := sqlLogEntry{
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
	}

	for _, obf := range obfuscators {
		log = obf.Obfuscate(log).(sqlLogEntry)
	}

	return l.Log(log)
}
