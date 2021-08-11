package dynamic

import (
	"context"
	"fmt"

	"github.com/gesundheitscloud/go-svc/pkg/log"
)

var _ Logger = (*DefaultLogger)(nil)

type Logger interface {
	SetName(name string)
	LogDebug(format string, a ...interface{})
	LogInfo(format string, a ...interface{})
	LogError(err error, format string, a ...interface{})
}

type DefaultLogger struct {
	name string
	log  *log.Logger
}

func NewDefaultLogger(name string, log *log.Logger) *DefaultLogger {
	return &DefaultLogger{name: name, log: log}
}

// Rename updates the viper-config name that should be printed with every message
// Useful to update the name after vc.Merge()
func (dl *DefaultLogger) SetName(name string) {
	dl.name = name
}

func (dl *DefaultLogger) LogInfo(format string, a ...interface{}) {
	dl.LogDebug(format, a...)
}

func (dl *DefaultLogger) LogDebug(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if dl.log != nil {
		_ = dl.log.InfoGeneric(context.Background(), fmt.Sprintf("[vc '%s']: %s", dl.name, msg))
	}
}

func (dl *DefaultLogger) LogError(err error, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if dl.log != nil {
		_ = dl.log.ErrGeneric(context.Background(), fmt.Errorf("[vc '%s']: %s %w", dl.name, msg, err))
	}
}
