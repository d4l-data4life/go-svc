package db

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
)

var (
	errSlowSQL = errors.New("Slow sql")
)

type Logger struct {
	logger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
	LogLevel                            LogLevel
}

type LogLevel int

const (
	Silent LogLevel = iota + 1
	Error
	Warn
	Info
)

func NewLogger(config logger.Config) *Logger {
	var (
		infoStr      = "%s"
		warnStr      = "%s"
		errStr       = "%s"
		traceStr     = "%s [%.3fms] [rows:%v] %s"
		traceWarnStr = "%s [%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s [%.3fms] [rows:%v] %s"
	)
	return &Logger{
		Config:       config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
		LogLevel:     LogLevel(config.LogLevel),
	}
}

// LogMode is needed for implementing the logger.Interface
// dummy replacement, as log level is handeld by our logger implementation go-svc/pkg/logging
func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = LogLevel(level)
	return &newlogger
}

func (l Logger) Info(ctx context.Context, msg string, fields ...interface{}) {
	if l.LogLevel >= Info {
		logging.LogInfofCtx(ctx, msg, fields...)
	}
}
func (l Logger) Warn(ctx context.Context, msg string, fields ...interface{}) {
	if l.LogLevel >= Warn {
		logging.LogWarningfCtx(ctx, nil, msg, fields...)
	}
}

func (l Logger) Error(ctx context.Context, msg string, fields ...interface{}) {
	if l.LogLevel >= Error {
		logging.LogErrorfCtx(ctx, nil, msg, fields...)
	}
}
func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError) && l.LogLevel >= Error:
		sql, rows := fc()
		if rows == -1 {
			logging.LogErrorfCtx(ctx, err, l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)

		} else {
			logging.LogErrorfCtx(ctx, err, l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= Warn:
		sql, rows := fc()
		if rows == -1 {
			logging.LogWarningfCtx(ctx, errSlowSQL, l.traceWarnStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			logging.LogWarningfCtx(ctx, errSlowSQL, l.traceWarnStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel >= Info:
		sql, rows := fc()
		if rows == -1 {
			logging.LogInfofCtx(ctx, l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			logging.LogInfofCtx(ctx, l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}
