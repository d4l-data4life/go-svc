package jwt

import (
	"context"
	"os"

	"github.com/gesundheitscloud/go-svc/pkg/log"
)

type logger interface {
	ErrUserAuth(context.Context, error) error
	InfoGeneric(context.Context, string) error
	ErrGeneric(context.Context, error) error
}

var pkgLogger = log.NewLogger("go-svc", "", os.Getenv("HOSTNAME"))
