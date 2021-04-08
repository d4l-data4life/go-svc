package log

import (
	"time"
)

func (l *Logger) ServiceStart() error {
	return l.Log(logEntry{
		Timestamp:      time.Now(),
		LogLevel:       LevelInfo,
		ServiceName:    l.serviceName,
		ServiceVersion: l.serviceVersion,
		Hostname:       l.hostname,
		EventType:      "service-start",
		ClientID:       "",
		TenantID:       l.tenantID,
	})
}
