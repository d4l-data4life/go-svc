package tut

import (
	"io/ioutil"

	"github.com/gesundheitscloud/go-svc/pkg/log"
)

type NopLogger struct {
	*log.Logger
}

func NewNopLogger() *NopLogger {
	return &NopLogger{
		Logger: log.NewLogger("test", "", "", log.WithWriter(ioutil.Discard)),
	}
}
