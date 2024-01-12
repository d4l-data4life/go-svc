package log

import (
	"time"
)

// logEntry is the format for generic logs
type logEntry struct {
	Timestamp      time.Time `json:"timestamp,omitempty"`
	LogLevel       logLevel  `json:"log-level,omitempty"`
	TraceID        string    `json:"trace-id,omitempty"`
	ServiceName    string    `json:"service-name,omitempty"`
	ServiceVersion string    `json:"service-version,omitempty"`
	Hostname       string    `json:"hostname,omitempty"`
	EventType      string    `json:"event-type,omitempty"`

	// TenantID is the ID of the tenant to which the log belongs to
	TenantID string `json:"tenant-id,omitempty"`

	UserID  string `json:"user-id,omitempty"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
	Object  string `json:"object,omitempty"`

	// OAuth client ID
	ClientID string `json:"client-id,omitempty"`
}
