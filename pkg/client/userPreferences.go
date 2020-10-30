package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	uuid "github.com/satori/go.uuid"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
)

// AllSettings store all settings (it is a map-of-maps, so it requires care when adding new entries)
type AllSettings map[uuid.UUID]UserSettings

func NewAllSettings() AllSettings {
	return make(map[uuid.UUID]UserSettings)
}

// Add adds safely new values to AllSettings
func (as *AllSettings) Add(accID uuid.UUID, key, value string) {
	if _, ok := (*as)[accID]; !ok {
		(*as)[accID] = NewUserSettings()
	}
	(*as)[accID][key] = value
}

// UserSettings store all settings for single user
type UserSettings map[string]string

func NewUserSettings() UserSettings {
	return make(map[string]string)
}

// GlobalSetting store single setting for all users
type GlobalSetting map[uuid.UUID]string

func NewGlobalSetting() GlobalSetting {
	return make(map[uuid.UUID]string)
}

type UserPreferences interface {
	// Get fetches a given setting for a particular account
	Get(ctx context.Context, accountID uuid.UUID, key string) (string, error)
	// GetKeySettings fetches a given setting for all accounts
	GetKeySettings(ctx context.Context, key string) (GlobalSetting, error)
	// GetAccountSettings fetches all settings for single accounts
	GetAccountSettings(ctx context.Context, accountID uuid.UUID) (UserSettings, error)
	// GetGlobal fetches all settings for all accounts
	GetGlobal(ctx context.Context) (AllSettings, error)
	// Set sets a given setting for a particular account
	Set(ctx context.Context, accountID uuid.UUID, key, value string) error
	// Delete deletes a single key for a user
	Delete(ctx context.Context, accountID uuid.UUID, key string) error
	// Delete DeleteUser all setting keys for a user
	DeleteUser(ctx context.Context, accountID uuid.UUID) error
}

var _ UserPreferences = (*UserPreferencesService)(nil)
var userAgentUserPrefs = "go-svc.client.UserPreferencesService"

// UserPreferencesService is a client for the cds-notification
// it implements UserPreferences and UserPreferencesV2 interfaces
type UserPreferencesService struct {
	svcAddr   string
	svcSecret string
	caller    string
}

func NewUserPreferencesService(svcAddr, svcSecret, caller string) *UserPreferencesService {
	if caller == "" {
		caller = "unknown"
	}
	return &UserPreferencesService{
		svcAddr:   svcAddr,
		svcSecret: svcSecret,
		caller:    caller,
	}
}

// Get fetches a single setting for a single user
func (c *UserPreferencesService) Get(ctx context.Context, accountID uuid.UUID, key string) (string, error) {
	contentURL := fmt.Sprintf("%s/api/v1/internal/users/%s/settings/%s", c.svcAddr, accountID.String(), key)
	byteSettings, _, err := call(ctx, contentURL, "GET", c.svcSecret, userAgentUserPrefs, &bytes.Buffer{}, http.StatusOK)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "fetching single setting failed")
		return "", err
	}
	return string(byteSettings), nil
}

// GetKeySettings fetches single setting for all users
func (c *UserPreferencesService) GetKeySettings(ctx context.Context, key string) (GlobalSetting, error) {
	setting := GlobalSetting{}

	// there is no such API to get a single setting for all users, so we need to compute it
	globalSettings, err := c.GetGlobal(ctx)
	if err != nil {
		return setting, err
	}
	for accID, usrSettings := range globalSettings {
		setting[accID] = usrSettings[key]
	}
	return setting, nil
}

// GetAccountSettings fetches all settings for a single user
func (c *UserPreferencesService) GetAccountSettings(ctx context.Context, accountID uuid.UUID) (UserSettings, error) {
	contentURL := fmt.Sprintf("%s/api/v1/internal/users/%s/settings/", c.svcAddr, accountID.String())
	settings := UserSettings{}
	byteSettings, _, err := call(ctx, contentURL, "GET", c.svcSecret, userAgentUserPrefs, &bytes.Buffer{}, http.StatusOK)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "fetching user settings failed")
		return UserSettings{}, err
	}
	if err := json.Unmarshal(byteSettings, &settings); err != nil {
		logging.LogErrorfCtx(ctx, err, "error transforming user-preferences service reply to an object")
		return nil, err
	}
	return settings, nil
}

// GetGlobal fetches all settings for all users
func (c *UserPreferencesService) GetGlobal(ctx context.Context) (AllSettings, error) {
	contentURL := fmt.Sprintf("%s/api/v1/internal/global/settings", c.svcAddr)
	settings := AllSettings{}
	byteSettings, _, err := call(ctx, contentURL, "GET", c.svcSecret, userAgentUserPrefs, &bytes.Buffer{}, http.StatusOK)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "fetching global settings failed")
		return AllSettings{}, err
	}
	if err := json.Unmarshal(byteSettings, &settings); err != nil {
		logging.LogErrorfCtx(ctx, err, "error transforming user-preferences service reply to an object")
		return nil, err
	}
	return settings, nil
}

// Set sets a single setting for a single user
func (c *UserPreferencesService) Set(ctx context.Context, accountID uuid.UUID, key, value string) error {
	contentURL := fmt.Sprintf("%s/api/v1/internal/users/%s/settings/%s", c.svcAddr, accountID.String(), key)
	// may feel weird, but PUT returns 204 on success
	_, _, err := call(ctx, contentURL, "PUT", c.svcSecret, userAgentUserPrefs, bytes.NewBuffer([]byte(value)), http.StatusNoContent)
	return err
}

func (c *UserPreferencesService) Delete(ctx context.Context, accountID uuid.UUID, key string) error {
	contentURL := fmt.Sprintf("%s/api/v1/internal/users/%s/settings/%s", c.svcAddr, accountID.String(), key)
	_, _, err := call(ctx, contentURL, "DELETE", c.svcSecret, userAgentUserPrefs, &bytes.Buffer{}, http.StatusNoContent)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "deleting single setting failed")
	}
	return err
}

func (c *UserPreferencesService) DeleteUser(ctx context.Context, accountID uuid.UUID) error {
	contentURL := fmt.Sprintf("%s/api/v1/internal/users/%s/settings", c.svcAddr, accountID.String())
	_, _, err := call(ctx, contentURL, "DELETE", c.svcSecret, userAgentUserPrefs, &bytes.Buffer{}, http.StatusNoContent)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "deleting all user settings failed")
	}
	return err
}
