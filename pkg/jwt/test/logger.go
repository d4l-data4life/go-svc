package test

import (
	"context"
	"fmt"
)

type Logger struct{}

func (Logger) ErrUserAuth(ctx context.Context, err error) error {
	fmt.Println(err)

	return nil
}
