package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	uuid "github.com/gofrs/uuid"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
)

// define errors
var (
	ErrEmptyKey        = errors.New("required key parameter was empty")
	ErrEmptyAccountIDs = errors.New("required accountIDs were missing")
)

// AllSettings store all settings (it is a map-of-maps, so it requires care when adding new entries)
type AllSettings map[uuid.UUID]UserSettings

func NewAllSettings() AllSettings {
	return make(map[uuid.UUID]UserSettings)
}

// Add adds safely new values to AllSettings
func (as *AllSettings) Add(accID uuid.UUID, key string, value interface{}) {
	if _, ok := (*as)[accID]; !ok {
		(*as)[accID] = NewUserSettings()
	}
	(*as)[accID][key] = value
}

// UserSettings store all settings for single user
type UserSettings map[string]interface{}

func NewUserSettings() UserSettings {
	return make(map[string]interface{})
}

// GlobalSetting store single setting for all users
type GlobalSetting map[uuid.UUID]interface{}

func NewGlobalSetting() GlobalSetting {
	return make(map[uuid.UUID]interface{})
}

type UserPreferences interface {
	// Get fetches a setting for a user
	Get(ctx context.Context, accountID uuid.UUID, key string) (interface{}, error)
	// GetKeySettings fetches a specific setting for all users
	GetKeySettings(ctx context.Context, key string) (GlobalSetting, error)
	// GetKeySettingsForUsers fetches a specific setting for a set of users
	GetKeySettingsForUsers(ctx context.Context, key string, accountIDs []uuid.UUID) (GlobalSetting, error)
	// GetUserSettings fetches all settings for a user
	GetUserSettings(ctx context.Context, accountID uuid.UUID) (UserSettings, error)
	// GetGlobal fetches all settings for all users
	GetGlobal(ctx context.Context) (AllSettings, error)
	// Set sets a setting for a user
	Set(ctx context.Context, accountID uuid.UUID, key string, value interface{}) error
	// Delete deletes a setting for a user
	Delete(ctx context.Context, accountID uuid.UUID, key string) error
	// DeleteUser deletes all settings for a user
	DeleteUser(ctx context.Context, accountID uuid.UUID) error
}

var _ UserPreferences = (*UserPreferencesService)(nil)
var userAgentUserPrefs = "go-svc.client.UserPreferencesService"

// UserPreferencesService is a client for the cds-notification
// it implements the UserPreferences APIv2 interface
type UserPreferencesService struct {
	svcAddr   string
	svcSecret string
	caller    *caller
}

func NewUserPreferencesService(svcAddr, svcSecret, caller string) *UserPreferencesService {
	if caller == "" {
		caller = "unknown"
	}
	return &UserPreferencesService{
		svcAddr:   svcAddr,
		svcSecret: svcSecret,
		caller:    NewCaller(30*time.Second, caller),
	}
}

// Get fetches a setting for a user
func (c *UserPreferencesService) Get(ctx context.Context, accountID uuid.UUID, key string) (interface{}, error) {
	if key == "" {
		return nil, ErrEmptyKey
	}
	contentURL := fmt.Sprintf("%s/api/v2/internal/users/%s/settings/%s", c.svcAddr, accountID.String(), key)
	byteSettings, _, err := c.caller.call(ctx, contentURL, "GET", c.svcSecret, userAgentUserPrefs, &bytes.Buffer{}, http.StatusOK)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "fetching single setting failed")
		return nil, err
	}
	var value interface{}
	err = json.Unmarshal(byteSettings, &value)
	return value, err
}

// GetKeySettings fetches a specific setting for all users
func (c *UserPreferencesService) GetKeySettings(ctx context.Context, key string) (GlobalSetting, error) {
	settings := GlobalSetting{}
	if key == "" {
		return settings, ErrEmptyKey
	}
	contentURL := fmt.Sprintf("%s/api/v2/internal/global/settings/%s", c.svcAddr, key)
	byteSettings, _, err := c.caller.call(ctx, contentURL, "GET", c.svcSecret, userAgentUserPrefs, &bytes.Buffer{}, http.StatusOK)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "fetching specific setting of all users failed")
		return GlobalSetting{}, err
	}
	if err := json.Unmarshal(byteSettings, &settings); err != nil {
		logging.LogErrorfCtx(ctx, err, "error transforming user-preferences service reply to an object")
		return nil, err
	}
	return settings, nil
}

// GetKeySettingsForUsers fetches a specific setting for a set of users
func (c *UserPreferencesService) GetKeySettingsForUsers(ctx context.Context, key string, accountIDs []uuid.UUID) (GlobalSetting, error) {
	settings := GlobalSetting{}
	if key == "" {
		return settings, ErrEmptyKey
	}
	if len(accountIDs) == 0 {
		return settings, ErrEmptyAccountIDs
	}
	contentURL := fmt.Sprintf("%s/api/v2/internal/global/settings/%s", c.svcAddr, key)
	jsonBytes, err := json.Marshal(accountIDs)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "cannot marshal account array")
		return settings, nil
	}
	byteSettings, _, err := c.caller.call(ctx, contentURL, "POST", c.svcSecret, userAgentUserPrefs, bytes.NewBuffer(jsonBytes), http.StatusOK)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "fetching specific setting of all users failed")
		return settings, err
	}
	if err := json.Unmarshal(byteSettings, &settings); err != nil {
		logging.LogErrorfCtx(ctx, err, "error transforming user-preferences service reply to an object")
		return nil, err
	}
	return settings, nil
}

// GetUserSettings fetches all settings for a user
func (c *UserPreferencesService) GetUserSettings(ctx context.Context, accountID uuid.UUID) (UserSettings, error) {
	contentURL := fmt.Sprintf("%s/api/v2/internal/users/%s/settings/", c.svcAddr, accountID.String())
	settings := UserSettings{}
	byteSettings, _, err := c.caller.call(ctx, contentURL, "GET", c.svcSecret, userAgentUserPrefs, &bytes.Buffer{}, http.StatusOK)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "fetching user settings failed")
		return settings, err
	}
	if err := json.Unmarshal(byteSettings, &settings); err != nil {
		logging.LogErrorfCtx(ctx, err, "error transforming user-preferences service reply to an object")
		return nil, err
	}
	return settings, nil
}

// GetGlobal fetches all settings for all users
func (c *UserPreferencesService) GetGlobal(ctx context.Context) (AllSettings, error) {
	contentURL := fmt.Sprintf("%s/api/v2/internal/global/settings", c.svcAddr)
	settings := AllSettings{}
	byteSettings, _, err := c.caller.call(ctx, contentURL, "GET", c.svcSecret, userAgentUserPrefs, &bytes.Buffer{}, http.StatusOK)
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

// Set sets a setting for a user
func (c *UserPreferencesService) Set(ctx context.Context, accountID uuid.UUID, key string, value interface{}) error {
	contentURL := fmt.Sprintf("%s/api/v2/internal/users/%s/settings/%s", c.svcAddr, accountID.String(), key)
	json, err := json.Marshal(value)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "cannot marshal value")
		return err
	}
	// may feel weird, but PUT returns 204 on success
	_, _, err = c.caller.call(ctx, contentURL, "PUT", c.svcSecret, userAgentUserPrefs, bytes.NewBuffer(json), http.StatusNoContent)
	return err
}

// Delete deletes a setting for a user
func (c *UserPreferencesService) Delete(ctx context.Context, accountID uuid.UUID, key string) error {
	if key == "" {
		return ErrEmptyKey
	}
	contentURL := fmt.Sprintf("%s/api/v2/internal/users/%s/settings/%s", c.svcAddr, accountID.String(), key)
	_, _, err := c.caller.call(ctx, contentURL, "DELETE", c.svcSecret, userAgentUserPrefs, &bytes.Buffer{}, http.StatusNoContent)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "deleting single setting failed")
	}
	return err
}

// DeleteUser deletes all settings for a user
func (c *UserPreferencesService) DeleteUser(ctx context.Context, accountID uuid.UUID) error {
	contentURL := fmt.Sprintf("%s/api/v2/internal/users/%s/settings", c.svcAddr, accountID.String())
	_, _, err := c.caller.call(ctx, contentURL, "DELETE", c.svcSecret, userAgentUserPrefs, &bytes.Buffer{}, http.StatusNoContent)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "deleting all user settings failed")
	}
	return err
}
