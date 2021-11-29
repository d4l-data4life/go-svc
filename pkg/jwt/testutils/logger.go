package testutils

import (
	"context"
	"fmt"
)

type Logger struct{}

func (Logger) ErrUserAuth(ctx context.Context, err error) error {
	fmt.Println(err)
	return nil
}
func (Logger) InfoGeneric(ctx context.Context, msg string) error {
	fmt.Println(msg)
	return nil
}
func (Logger) ErrGeneric(ctx context.Context, err error) error {
	fmt.Println(err)
	return nil
}
