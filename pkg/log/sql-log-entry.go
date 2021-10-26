package log

import (
	"time"
)

type sqlLogEntry struct {
	Timestamp      time.Time `json:"timestamp"`
	LogLevel       logLevel  `json:"log-level"`
	TraceID        string    `json:"trace-id"`
	ServiceName    string    `json:"service-name"`
	ServiceVersion string    `json:"service-version"`
	Hostname       string    `json:"hostname"`
	EventType      string    `json:"event-type"`

	// TenantID is the ID of the tenant to which the log belongs to
	TenantID string `json:"tenant-id"`

	UserID string `json:"user-id,omitempty"`

	// OAuth client ID
	ClientID string `json:"client-id,omitempty"`

	PgxLogLevel string `json:"pgx-log-level,omitempty"`
	PgxMessage  string `json:"pgx-message,omitempty"`
	PgxData     string `json:"pgx-data,omitempty"`
}
