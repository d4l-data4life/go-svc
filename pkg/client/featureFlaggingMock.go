package client

import (
	"context"
)

var _ FeatureFlagging = (*FeatureFlaggingMock)(nil)

type FeatureFlaggingMock struct {
	EnabledKeys []string
}

// Get fetches a single setting for a single user
func (f *FeatureFlaggingMock) Get(ctx context.Context, key string) (bool, error) {
	for _, enabledKey := range f.EnabledKeys {
		if key == enabledKey {
			return true, nil
		}
	}
	return false, nil
}
