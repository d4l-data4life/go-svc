package client

import (
	"context"

	uuid "github.com/gofrs/uuid"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
)

var _ UserPreferences = (*UserPreferencesMock)(nil)

// UserPreferencesMock is a mocked client for the user-preferences service
type UserPreferencesMock struct {
	storage AllSettings
}

func NewUserPreferencesMock() *UserPreferencesMock {
	// use some hardcoded data
	data := map[uuid.UUID]struct{ K, V string }{
		uuid.FromStringOrNil("02c172e0-1d22-41f8-a4c1-5613f9882daa"): {"key1", "02c172e0-key1-value"},
		uuid.FromStringOrNil("02c172e0-1d22-41f8-a4c1-5613f9882daa"): {"key2", "02c172e0-key2-value"},
		uuid.FromStringOrNil("4f140045-4764-47c2-a8b5-71a7d3515928"): {"key1", "4f140045-key1-value"},
		uuid.FromStringOrNil("4f140045-4764-47c2-a8b5-71a7d3515928"): {"key3", "4f140045-key1-value"},
		uuid.FromStringOrNil("c7908420-b9d3-4a1b-844c-ea0eb363f0bd"): {"key1", ""},
	}
	return NewUserPreferencesMockWithState(data)
}

func NewUserPreferencesMockWithState(data map[uuid.UUID]struct{ K, V string }) *UserPreferencesMock {
	storage := NewAllSettings()
	for accID, pair := range data {
		storage.Add(accID, pair.K, pair.V)
	}
	return &UserPreferencesMock{storage}
}

// Get fetches a single setting for a single user
func (c *UserPreferencesMock) Get(ctx context.Context, accountID uuid.UUID, key string) (string, error) {
	return c.storage[accountID][key], nil
}

// GetKeySettings fetches single setting for all users
func (c *UserPreferencesMock) GetKeySettings(ctx context.Context, key string) (GlobalSetting, error) {
	setting := NewGlobalSetting()
	// there is no such API to get a single setting for all users, so we need to compute it
	globalSettings, err := c.GetGlobal(ctx)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error fetching global settings")
		return setting, err
	}
	for accID, usrSettings := range globalSettings {
		setting[accID] = usrSettings[key]
	}
	return setting, nil
}

// GetAccountSettings fetches all settings for a single user
func (c *UserPreferencesMock) GetAccountSettings(ctx context.Context, accountID uuid.UUID) (UserSettings, error) {
	return c.storage[accountID], nil
}

// GetGlobal fetches all settings for all users
func (c *UserPreferencesMock) GetGlobal(ctx context.Context) (AllSettings, error) {
	return c.storage, nil
}

// Set sets a single setting for a single user
func (c *UserPreferencesMock) Set(ctx context.Context, accountID uuid.UUID, key, value string) error {
	c.storage.Add(accountID, key, value)
	return nil
}

func (c *UserPreferencesMock) Delete(ctx context.Context, accountID uuid.UUID, key string) error {
	if _, ok := c.storage[accountID]; !ok {
		return nil
	}
	delete(c.storage[accountID], key)
	return nil
}

func (c *UserPreferencesMock) DeleteUser(ctx context.Context, accountID uuid.UUID) error {
	delete(c.storage, accountID)
	return nil
}
