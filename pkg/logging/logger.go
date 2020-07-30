package logging

import (
	"context"
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
		encoder := golog.NewJSONEncoder(os.Stdout)
		if lConf.HumanReadable {
			encoder = golog.NewPrettyEncoder(os.Stdout)
		}
		instance = golog.NewLogger(
			lConf.SvcName,
			lConf.SvcVersion,
			os.Getenv("HOSTNAME"),
			golog.WithEncoder(encoder),
		)
	})
	return instance
}

// LogError logs an error with the singleton logger with message and error
func LogError(message string, err error) {
	LogErrorf(err, message)
}

// LogErrorf logs an error with the singleton logger with message and error
func LogErrorf(err error, format string, fields ...interface{}) {
	msg := fmt.Sprintf(format, fields...)
	if err := Logger().ErrMessage(context.TODO(), strings.TrimRight(msg, "\n"), err); err != nil {
		fmt.Printf("Logging error (LogErrorf): %s", err.Error())
	}
}

// LogWarning logs a warning with the singleton logger with message and error
func LogWarning(message string, err error) {
	LogWarningf(err, message)
}

// LogWarningf logs a warning with the singleton logger with message and error
func LogWarningf(err error, format string, fields ...interface{}) {
	msg := fmt.Sprintf(format, fields...)
	if err := Logger().WarnGeneric(context.TODO(), strings.TrimRight(msg, "\n"), err); err != nil {
		fmt.Printf("Logging error (LogWarningf): %s", err.Error())
	}
}

// LogAudit logs a warning with the singleton logger with message and error
func LogAudit(message string, object ...interface{}) {
	if err := Logger().Audit(context.TODO(), strings.TrimRight(message, "\n"), object); err != nil {
		fmt.Printf("Logging error (LogAudit): %s", err.Error())
	}
}

// LogInfo logs an info with the singleton logger with message and error
func LogInfo(message string) {
	LogInfof(message)
}

// LogInfof info-level log with formatting
func LogInfof(format string, fields ...interface{}) {
	msg := fmt.Sprintf(format, fields...)
	if err := Logger().InfoGeneric(context.TODO(), strings.TrimRight(msg, "\n")); err != nil {
		fmt.Printf("Logging error (LogInfof): %s", err.Error())
	}
}

// LogDebugf acts as log-info when config.Debug flag is enabled
func LogDebugf(format string, fields ...interface{}) {
	if LoggerConfig().Debug {
		msg := fmt.Sprintf(format, fields...)
		if err := Logger().InfoGeneric(context.TODO(), "DEBUG: "+strings.TrimRight(msg, "\n")); err != nil {
			fmt.Printf("Logging error (LogDebugf): %s", err.Error())
		}
	}
}
