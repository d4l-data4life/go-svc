package jwt

import "context"

type logger interface {
	ErrUserAuth(context.Context, error) error
	InfoGeneric(context.Context, string) error
	ErrGeneric(context.Context, error) error
}
