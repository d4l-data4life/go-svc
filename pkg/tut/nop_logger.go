package tut

import (
	"io"

	"github.com/gesundheitscloud/go-svc/pkg/log"
)

type NopLogger struct {
	*log.Logger
}

func NewNopLogger() *NopLogger {
	return &NopLogger{
		Logger: log.NewLogger("test", "", "", log.WithWriter(io.Discard)),
	}
}
