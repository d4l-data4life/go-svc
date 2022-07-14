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

// Get fetches a single setting
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

// GetForUser fetches a single setting for a user
func (f *FeatureFlaggingMock) GetForUser(ctx context.Context, key string, authorization string) (bool, error) {
	return f.Get(ctx, key)
}

type FeatureFlaggingErrorMock struct{}

// Get fetches a single setting
func (f *FeatureFlaggingErrorMock) Get(ctx context.Context, key string) (bool, error) {
	return false, errors.New("Not reachable")
}

// Get fetches a single setting for a user
func (f *FeatureFlaggingErrorMock) GetForUser(ctx context.Context, key string, authorization string) (bool, error) {
	return f.Get(ctx, key)
}
