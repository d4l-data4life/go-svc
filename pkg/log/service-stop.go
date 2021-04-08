package log

import (
	"time"
)

func (l *Logger) ServiceStop() error {
	return l.Log(logEntry{
		Timestamp:      time.Now(),
		LogLevel:       LevelInfo,
		ServiceName:    l.serviceName,
		ServiceVersion: l.serviceVersion,
		Hostname:       l.hostname,
		EventType:      "service-stop",
		ClientID:       "",
		TenantID:       l.tenantID,
	})
}
