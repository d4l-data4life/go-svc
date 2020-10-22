package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	golog "github.com/gesundheitscloud/go-log/v2/log"
)

type LoggerOption func(*Config)

func Unnamed() LoggerOption {
	return func(c *Config) {
		c.SvcName = "unnamed"
		c.SvcVersion = "unknown"
		c.Hostname = ""
		c.HumanReadable = false
		c.Debug = false
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

type Config struct {
	HumanReadable bool
	Debug         bool
	SvcName       string
	SvcVersion    string
	Hostname      string
	Environment   string
	PodName       string
}

var onceLogger sync.Once
var instance *golog.Logger

var onceLoggerConfig sync.Once
var loggerConfig *Config

type stringer string

func (s stringer) String() string { return string(s) }

type jsonStringer struct {
	obj interface{}
}

func newJSONStringer(obj interface{}) jsonStringer {
	return jsonStringer{obj}
}

func (js jsonStringer) String() string {
	json, _ := json.Marshal(js.obj)
	return string(json)
}

func newStringer(obj interface{}) fmt.Stringer {
	switch obj := obj.(type) {
	case string:
		return stringer(obj)
	case fmt.Stringer:
		return obj
	default:
		return newJSONStringer(obj)
	}
}

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
		encoder := golog.NewJSONEncoder(os.Stdout)
		if lConf.HumanReadable {
			encoder = golog.NewPrettyEncoder(os.Stdout)
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
	LogErrorf(err, message)
}

// LogErrorf (DEPRECATED in favour of LogErrorfCtx) logs an error with the singleton logger with message and error
func LogErrorf(err error, format string, fields ...interface{}) {
	msg := fmt.Sprintf(format, fields...)
	if err := Logger().ErrMessage(context.TODO(), strings.TrimRight(msg, "\n"), err); err != nil {
		fmt.Printf("Logging error (LogErrorf): %s", err.Error())
	}
}

// LogErrorfCtx logs an error with the singleton logger with message, error, and context
func LogErrorfCtx(ctx context.Context, err error, format string, fields ...interface{}) {
	msg := fmt.Sprintf(format, fields...)
	if err := Logger().ErrMessage(ctx, strings.TrimRight(msg, "\n"), err); err != nil {
		fmt.Printf("Logging error (LogErrorf): %s", err.Error())
	}
}

// LogWarning (DEPRECATED in favour of LogWarningfCtx) logs a warning with the singleton logger with message and error
func LogWarning(message string, err error) {
	LogWarningf(err, message)
}

// LogWarningf (DEPRECATED in favour of LogWarningfCtx) logs a warning with the singleton logger with message and error
func LogWarningf(err error, format string, fields ...interface{}) {
	msg := fmt.Sprintf(format, fields...)
	if err := Logger().WarnGeneric(context.TODO(), strings.TrimRight(msg, "\n"), err); err != nil {
		fmt.Printf("Logging error (LogWarningf): %s", err.Error())
	}
}

// LogWarningfCtx logs a warning with the singleton logger with message and error
func LogWarningfCtx(ctx context.Context, err error, format string, fields ...interface{}) {
	msg := fmt.Sprintf(format, fields...)
	if err := Logger().WarnGeneric(ctx, strings.TrimRight(msg, "\n"), err); err != nil {
		fmt.Printf("Logging error (LogWarningf): %s", err.Error())
	}
}

// LogAudit logs a warning with the singleton logger with message and error
func LogAudit(ctx context.Context, message string, object ...interface{}) {
	if err := Logger().Audit(ctx, strings.TrimRight(message, "\n"), object); err != nil {
		fmt.Printf("Logging error (LogAudit): %s", err.Error())
	}
}

// LogAuditCreate logs a resource creation with the singleton logger with message and error
func LogAuditCreate(ctx context.Context, ownerID fmt.Stringer, resourceType string, resourceID interface{}, value interface{}) {
	err := Logger().AuditCreate(ctx, ownerID, newStringer(resourceType), newStringer(resourceID), value)
	if err != nil {
		fmt.Printf("Logging error (LogAuditCreate): %s", err.Error())
	}
}

// LogAuditUpdate logs a resource modification with the singleton logger with message and error
func LogAuditUpdate(ctx context.Context, ownerID fmt.Stringer, resourceType string, resourceID interface{}, value interface{}) {
	err := Logger().AuditUpdate(ctx, ownerID, newStringer(resourceType), newStringer(resourceID), value)
	if err != nil {
		fmt.Printf("Logging error (LogAuditUpdate): %s", err.Error())
	}
}

// LogAuditDelete logs a resource deletion with the singleton logger with message and error
func LogAuditDelete(ctx context.Context, ownerID fmt.Stringer, resourceType string, resourceID interface{}) {
	err := Logger().AuditDelete(ctx, ownerID, newStringer(resourceType), newStringer(resourceID))
	if err != nil {
		fmt.Printf("Logging error (LogAuditDelete): %s", err.Error())
	}
}

// LogInfo (DEPRECATED in favour of LogInfofCtx) logs an info with the singleton logger with message and error
func LogInfo(message string) {
	LogInfof(message)
}

// LogInfof (DEPRECATED in favour of LogInfofCtx) info-level log with formatting
func LogInfof(format string, fields ...interface{}) {
	msg := fmt.Sprintf(format, fields...)
	if err := Logger().InfoGeneric(context.TODO(), strings.TrimRight(msg, "\n")); err != nil {
		fmt.Printf("Logging error (LogInfof): %s", err.Error())
	}
}

// LogInfofCtx info-level log with formatting
func LogInfofCtx(ctx context.Context, format string, fields ...interface{}) {
	msg := fmt.Sprintf(format, fields...)
	if err := Logger().InfoGeneric(ctx, strings.TrimRight(msg, "\n")); err != nil {
		fmt.Printf("Logging error (LogInfof): %s", err.Error())
	}
}

// LogDebugf (DEPRECATED in favour of LogDebugfCtx) acts as log-info when config.Debug flag is enabled
func LogDebugf(format string, fields ...interface{}) {
	if LoggerConfig().Debug {
		msg := fmt.Sprintf(format, fields...)
		if err := Logger().InfoGeneric(context.TODO(), "DEBUG: "+strings.TrimRight(msg, "\n")); err != nil {
			fmt.Printf("Logging error (LogDebugf): %s", err.Error())
		}
	}
}

// LogDebugfCtx acts as log-info when config.Debug flag is enabled
func LogDebugfCtx(ctx context.Context, format string, fields ...interface{}) {
	if LoggerConfig().Debug {
		msg := fmt.Sprintf(format, fields...)
		if err := Logger().InfoGeneric(ctx, "DEBUG: "+strings.TrimRight(msg, "\n")); err != nil {
			fmt.Printf("Logging error (LogDebugf): %s", err.Error())
		}
	}
}
