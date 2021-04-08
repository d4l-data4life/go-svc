package migrate

import "context"

type logger interface {
	InfoGeneric(context.Context, string) error
}
