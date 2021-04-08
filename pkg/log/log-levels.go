package log

type logLevel string

const (
	LevelError   logLevel = "error"
	LevelWarning logLevel = "warn"
	LevelInfo    logLevel = "info"
	LevelAudit   logLevel = "audit"
)
