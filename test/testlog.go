package test

import (
	"context"
	"log"
)

type testLog struct{}

func (l *testLog) InfoGeneric(_ context.Context, msg string) error {
	log.Println(msg)
	return nil
}
