package dynamic

import (
	"crypto/rsa"
	"fmt"
	"io"

	"github.com/golang-jwt/jwt/v4"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// PublicKey represents an object storing information about public key from Viper config
type PublicKey struct {
	Name      string
	Comment   string
	NotBefore string
	NotAfter  string
	Key       *rsa.PublicKey
}

var _ KeyProvider = (*ViperConfig)(nil)

type KeyProvider interface {
	PublicKeys() ([]PublicKey, error)
}

// PublicKeyConfig models the structure of the Viper config fragment responsible for defining JWT public keys
type PublicKeyConfig struct {
	Entries []struct {
		Name      string `mapstructure:"name"`
		Comment   string `mapstructure:"comment"`
		Key       string `mapstructure:"key"`
		NotBefore string `mapstructure:"not_before"`
		NotAfter  string `mapstructure:"not_after"`
	} `mapstructure:"jwtpublickey"`
}

type ViperConfigOpt func(vp *ViperConfig)

func ConfigSource(configSrc io.Reader) ViperConfigOpt {
	return func(vp *ViperConfig) {
		vp.useCustomSource = true
		vp.configSrc = configSrc
	}
}
func ConfigFilePaths(configPaths ...string) ViperConfigOpt {
	return func(vp *ViperConfig) {
		vp.configPaths = append(vp.configPaths, configPaths...)
	}
}
func ConfigFileName(configFileName string) ViperConfigOpt {
	return func(vp *ViperConfig) {
		vp.configFileName = configFileName
	}
}
func ConfigFormat(configFormat string) ViperConfigOpt {
	return func(vp *ViperConfig) {
		vp.configFormat = configFormat
	}
}
func AutoBootstrap(b bool) ViperConfigOpt {
	return func(vp *ViperConfig) {
		vp.autoBootstrap = b
	}
}
func WatchChanges(b bool) ViperConfigOpt {
	return func(vp *ViperConfig) {
		vp.watchChanges = b
	}
}
func UpdateErrorHandler(fn func(error)) ViperConfigOpt {
	return func(vp *ViperConfig) {
		vp.updateErrorHandler = fn
	}
}

type ViperConfig struct {
	// Error stores config bootstraping and loading errors
	Error error

	// updateErrorHandler will be called if an error occurs while auto-updating viper config
	updateErrorHandler func(error)

	viperKeyJWT  string
	bootstrapped bool
	// autoBootstrap set to true will bootstrap Viper config in the constructor and ignore all errors that may happen
	// set to false to have full controll over errors
	autoBootstrap bool
	// watchChanges set to true will enable default onChange behavior already in the constructor
	watchChanges bool

	// useCustomSource set to true results in using 'configSrc' instead of default config files
	useCustomSource bool
	// configSrc holds a reader form which the config should be read. If empty, the default files will be used
	configSrc io.Reader
	// configFileName default is config (config.yaml) - name of file where config will be read from
	configFileName string
	// configPaths defines where to search for the config file - defaults contain: `.`, `/etc/config`, `test/config`
	configPaths []string
	// configFormat default is yaml - see Viper docs for r formats supported
	configFormat string
	// stores the parsed keys from the config
	keys []PublicKey
	// v stores a viper instance (default is global)
	v *viper.Viper
}

// NewViperConfig provides configuration using Viper
// Developers using autobootstrap should check Error field for possible bootstrapping errors
func NewViperConfig(opts ...ViperConfigOpt) *ViperConfig {
	vp := &ViperConfig{
		viperKeyJWT:        "JWTPublicKey", // must be named JWTPublicKey due to annotation in PublicKeyConfig: `mapstructure:"jwtpublickey"`
		configSrc:          nil,
		configPaths:        []string{"."},
		configFileName:     "config",
		configFormat:       "yaml", // default is yaml
		bootstrapped:       false,
		useCustomSource:    false,
		autoBootstrap:      true,
		watchChanges:       true,
		keys:               make([]PublicKey, 0),
		v:                  viper.New(),
		Error:              nil,
		updateErrorHandler: func(e error) {},
	}
	for _, opt := range opts {
		opt(vp)
	}
	if vp.autoBootstrap {
		if err := vp.Bootstrap(); err != nil {
			vp.Error = fmt.Errorf("bootstrap error: %w", err)
		}
	}
	if vp.watchChanges {
		vp.WatchConfig()
	}
	return vp
}

// Bootstrap initializes ViperConfig form a config.yaml file being read from configurable locations
// Providing config in form of a file is a prerequisite for using 'WatchConfig' and automated loading of changes
func (vc *ViperConfig) Bootstrap() error {
	if vc.bootstrapped {
		return nil
	}

	vc.v.SetConfigType(vc.configFormat)
	// do not read from files - use the reader instead
	if vc.useCustomSource {
		err := vc.v.ReadConfig(vc.configSrc)
		vc.bootstrapped = err == nil
		return err
	}

	vc.v.SetConfigName(vc.configFileName)
	for _, p := range vc.configPaths { // read from the reader if provided
		vc.v.AddConfigPath(p)
	}
	err := vc.v.ReadInConfig()
	if err != nil {
		if err2, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found
			// this error can be optionally ignored - maybe in a moment the config file will be fixed and auto-reloaded?
			return fmt.Errorf("cannot find config file: %s \n", vc.v.ConfigFileUsed())
		} else {
			// Config file was found but another error was produced
			return fmt.Errorf("error reading config file: %w \n", err2)
		}
	}
	vc.bootstrapped = true
	return nil
}

// WatchConfig enables a listener that will execute `updatePublicKeys` automatically on every secret/config change
func (vc *ViperConfig) WatchConfig() {
	vc.v.WatchConfig()
	// Define what to do when config changes
	vc.v.OnConfigChange(func(e fsnotify.Event) {
		vc.updateErrorHandler(vc.updatePublicKeys())
	})
}

func (vc *ViperConfig) PublicKeys() ([]PublicKey, error) {
	if len(vc.keys) > 0 {
		return vc.keys, nil
	}
	err := vc.updatePublicKeys()
	return vc.keys, err
}
func (vc *ViperConfig) updatePublicKeys() error {
	if err := vc.Bootstrap(); err != nil {
		return err
	}
	keys, err := parsePublicKeys(vc.v, vc.viperKeyJWT)
	if err != nil {
		return err
	}
	vc.keys = keys // store the keys to speed-up access in the future
	return nil
}

// parsePublicKeys un-marshals viper config and parses PEM public keys from it
// The Keys and their metadata are stored in the PublicKey struct
func parsePublicKeys(vp *viper.Viper, key string) ([]PublicKey, error) {
	keys := make([]PublicKey, 0)
	var config PublicKeyConfig
	err := vp.Unmarshal(&config)
	if err != nil {
		return keys, fmt.Errorf("unable to decode into struct, %v", err)
	}
	for _, entry := range config.Entries {
		if entry.Key == "" {
			return keys, fmt.Errorf("error parsing the public key entry '%v': key is empty", entry)
		}
		publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(entry.Key))
		if err != nil {
			return keys, fmt.Errorf("error parsing the public key '%s': %w", entry.Name, err)
		}
		keys = append(keys, PublicKey{Key: publicKey, Name: entry.Name, Comment: entry.Comment, NotBefore: entry.NotBefore, NotAfter: entry.NotAfter})
	}
	return keys, nil
}
