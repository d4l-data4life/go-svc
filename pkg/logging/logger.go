package logging

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	golog "github.com/d4l-data4life/go-svc/pkg/log"
)

type LoggerOption func(*Config)

func Unnamed() LoggerOption {
	return func(c *Config) {
		c.SvcName = "unnamed"
		c.SvcVersion = "unknown"
		c.Hostname = ""
		c.HumanReadable = false
		c.Debug = false
		c.OutputFile = os.Stdout
	}
}

func ServiceName(value string) LoggerOption {
	return func(c *Config) {
		c.SvcName = value
	}
}

func ServiceVersion(value string) LoggerOption {
	return func(c *Config) {
		c.SvcVersion = value
	}
}

func Hostname(value string) LoggerOption {
	return func(c *Config) {
		c.Hostname = value
	}
}

func Environment(value string) LoggerOption {
	return func(c *Config) {
		c.Environment = value
	}
}

func PodName(value string) LoggerOption {
	return func(c *Config) {
		c.PodName = value
	}
}

func HumanReadable(value bool) LoggerOption {
	return func(c *Config) {
		c.HumanReadable = value
	}
}

func Debug(value bool) LoggerOption {
	return func(c *Config) {
		c.Debug = value
	}
}

func OutputFile(file *os.File) LoggerOption {
	return func(c *Config) {
		c.OutputFile = file
	}
}

type Config struct {
	HumanReadable bool
	Debug         bool
	SvcName       string
	SvcVersion    string
	Hostname      string
	Environment   string
	PodName       string
	OutputFile    *os.File
}

var onceLogger sync.Once
var instance *golog.Logger

var onceLoggerConfig sync.Once
var loggerConfig *Config

func LoggerConfig(opts ...LoggerOption) *Config {
	onceLoggerConfig.Do(func() {
		loggerConfig = &Config{}

		Unnamed()(loggerConfig)
	})
	for _, opt := range opts {
		opt(loggerConfig)
	}
	return loggerConfig
}

// Logger returns a global singleton logger object to access go-log logger
func Logger(opts ...LoggerOption) *golog.Logger {
	onceLogger.Do(func() {
		lConf := LoggerConfig(opts...)

		var encoder golog.Encoder
		if lConf.OutputFile == nil {
			// If OutputFile is explicitly set to nil, disable logging
			encoder = golog.NewNullEncoder()
		} else {
			// Use the configured output file (defaults to os.Stdout)
			if lConf.HumanReadable {
				encoder = golog.NewPrettyEncoder(lConf.OutputFile)
			} else {
				encoder = golog.NewJSONEncoder(lConf.OutputFile)
			}
		}

		instance = golog.NewLogger(
			lConf.SvcName,
			lConf.SvcVersion,
			lConf.Hostname,
			golog.WithPodName(lConf.PodName),
			golog.WithEnv(lConf.Environment),
			golog.WithEncoder(encoder),
		)
	})
	return instance
}

// LogError (DEPRECATED in favour of LogErrorfCtx) logs an error with the singleton logger with message and error
func LogError(message string, err error) {
	LogErrorf(err, "%s", message)
}

// LogErrorf (DEPRECATED in favour of LogErrorfCtx) logs an error with the singleton logger with message and error
func LogErrorf(err error, format string, fields ...interface{}) {
	msg := fmt.Sprintf(format, fields...)
	if err := Logger().ErrMessage(context.TODO(), strings.TrimRight(msg, "\n"), err); err != nil {
		fmt.Printf("Logging error (LogErrorf): %s\n", err.Error())
	}
}

// LogErrorfCtx logs an error with the singleton logger with message, error, and context
func LogErrorfCtx(ctx context.Context, err error, format string, fields ...interface{}) {
	msg := fmt.Sprintf(format, fields...)
	if err := Logger().ErrMessage(ctx, strings.TrimRight(msg, "\n"), err); err != nil {
		fmt.Printf("Logging error (LogErrorf): %s\n", err.Error())
	}
}

// LogWarning (DEPRECATED in favour of LogWarningfCtx) logs a warning with the singleton logger with message and error
func LogWarning(message string, err error) {
	LogWarningf(err, "%s", message)
}

// LogWarningf (DEPRECATED in favour of LogWarningfCtx) logs a warning with the singleton logger with message and error
func LogWarningf(err error, format string, fields ...interface{}) {
	msg := fmt.Sprintf(format, fields...)
	if err := Logger().WarnGeneric(context.TODO(), strings.TrimRight(msg, "\n"), err); err != nil {
		fmt.Printf("Logging error (LogWarningf): %s\n", err.Error())
	}
}

// LogWarningfCtx logs a warning with the singleton logger with message and error
func LogWarningfCtx(ctx context.Context, err error, format string, fields ...interface{}) {
	msg := fmt.Sprintf(format, fields...)
	if err := Logger().WarnGeneric(ctx, strings.TrimRight(msg, "\n"), err); err != nil {
		fmt.Printf("Logging error (LogWarningf): %s\n", err.Error())
	}
}

// LogAudit logs a generic audit event containing of a message along with an object pertaining to the message.
func LogAudit(ctx context.Context, message string, object any) {
	if err := Logger().Audit(ctx, message, object); err != nil {
		fmt.Printf("Logging error (LogAudit): %s\n", err.Error())
	}
}

// LogAuditSecuritySuccess logs a successful access with the singleton logger with message and error
func LogAuditSecuritySuccess(ctx context.Context, securityEvent string, extras ...golog.ExtraAuditInfoProvider) {
	if err := Logger().AuditSecuritySuccess(ctx, securityEvent, extras...); err != nil {
		fmt.Printf("Logging error (LogAuditSecuritySuccess): %s\n", err.Error())
	}
}

// LogAuditSecurityFailure logs a failed access with the singleton logger with message and error
func LogAuditSecurityFailure(ctx context.Context, securityEvent string, extras ...golog.ExtraAuditInfoProvider) {
	if err := Logger().AuditSecurityFailure(ctx, securityEvent, extras...); err != nil {
		fmt.Printf("Logging error (LogAuditSecurityFailure): %s\n", err.Error())
	}
}

// LogAuditCreate logs a resource creation with the singleton logger with message and error
func LogAuditCreate(ctx context.Context, ownerID string, resourceType string, resourceID string, value interface{}) {
	err := Logger().AuditCreate(ctx, ownerID, resourceType, resourceID, value)
	if err != nil {
		fmt.Printf("Logging error (LogAuditCreate): %s\n", err.Error())
	}
}

// LogAuditUpdate logs a resource modification with the singleton logger with message and error
func LogAuditUpdate(ctx context.Context, ownerID string, resourceType string, resourceID string, value interface{}) {
	err := Logger().AuditUpdate(ctx, ownerID, resourceType, resourceID, value)
	if err != nil {
		fmt.Printf("Logging error (LogAuditUpdate): %s\n", err.Error())
	}
}

// LogAuditDelete logs a resource deletion with the singleton logger with message and error
func LogAuditDelete(ctx context.Context, ownerID string, resourceType string, resourceID string) {
	err := Logger().AuditDelete(ctx, ownerID, resourceType, resourceID)
	if err != nil {
		fmt.Printf("Logging error (LogAuditDelete): %s\n", err.Error())
	}
}

// LogAuditRead logs a successful resource read access with the singleton logger with message and error
func LogAuditRead(ctx context.Context, ownerID string, resourceType string, resourceID string, extras ...golog.ExtraAuditInfoProvider) {
	err := Logger().AuditRead(ctx, ownerID, resourceType, resourceID, extras...)
	if err != nil {
		fmt.Printf("Logging error (LogAuditRead): %s\n", err.Error())
	}
}

// LogAuditBulkRead logs a successful resource bulk read access with the singleton logger with message and error
func LogAuditBulkRead(
	ctx context.Context,
	ownerID string,
	resourceType string,
	resourceIDs []string,
	extras ...golog.ExtraAuditInfoProvider,
) {
	err := Logger().AuditBulkRead(ctx, ownerID, resourceType, resourceIDs, extras...)
	if err != nil {
		fmt.Printf("Logging error (LogAuditBulkRead): %s\n", err.Error())
	}
}

// LogInfo (DEPRECATED in favour of LogInfofCtx) logs an info with the singleton logger with message and error
func LogInfo(message string) {
	LogInfof("%s", message)
}

// LogInfof (DEPRECATED in favour of LogInfofCtx) info-level log with formatting
func LogInfof(format string, fields ...interface{}) {
	msg := fmt.Sprintf(format, fields...)
	if err := Logger().InfoGeneric(context.TODO(), strings.TrimRight(msg, "\n")); err != nil {
		fmt.Printf("Logging error (LogInfof): %s\n", err.Error())
	}
}

// LogInfofCtx info-level log with formatting
func LogInfofCtx(ctx context.Context, format string, fields ...interface{}) {
	msg := fmt.Sprintf(format, fields...)
	if err := Logger().InfoGeneric(ctx, strings.TrimRight(msg, "\n")); err != nil {
		fmt.Printf("Logging error (LogInfof): %s\n", err.Error())
	}
}

// LogDebugf (DEPRECATED in favour of LogDebugfCtx) acts as log-info when config.Debug flag is enabled
func LogDebugf(format string, fields ...interface{}) {
	if LoggerConfig().Debug {
		msg := fmt.Sprintf(format, fields...)
		if err := Logger().InfoGeneric(context.TODO(), "DEBUG: "+strings.TrimRight(msg, "\n")); err != nil {
			fmt.Printf("Logging error (LogDebugf): %s\n", err.Error())
		}
	}
}

// LogDebugfCtx acts as log-info when config.Debug flag is enabled
func LogDebugfCtx(ctx context.Context, format string, fields ...interface{}) {
	if LoggerConfig().Debug {
		msg := fmt.Sprintf(format, fields...)
		if err := Logger().InfoGeneric(ctx, "DEBUG: "+strings.TrimRight(msg, "\n")); err != nil {
			fmt.Printf("Logging error (LogDebugf): %s\n", err.Error())
		}
	}
}
