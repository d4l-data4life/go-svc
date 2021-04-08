package jwt

import "context"

type logger interface {
	ErrUserAuth(context.Context, error) error
}
