package client

import (
	"context"
	"errors"
	"time"
)

var _ FeatureFlagging = (*FeatureFlaggingMock)(nil)

type FeatureFlaggingMock struct {
	EnabledKeys []string
	Delay       time.Duration
}

// Get fetches a single setting for a single user
func (f *FeatureFlaggingMock) Get(ctx context.Context, key string) (bool, error) {
	select {
	case <-time.After(f.Delay):
		for _, enabledKey := range f.EnabledKeys {
			if key == enabledKey {
				return true, nil
			}
		}
		return false, nil

	case <-ctx.Done():
		return false, ctx.Err()
	}

}

type FeatureFlaggingErrorMock struct{}

// Get fetches a single setting for a single user
func (f *FeatureFlaggingErrorMock) Get(ctx context.Context, key string) (bool, error) {
	return false, errors.New("Not reachable")
}
