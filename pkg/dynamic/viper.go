package dynamic

import (
	"crypto/rsa"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v4"

	"github.com/gesundheitscloud/go-svc/pkg/log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

/*
Example Viper config expected to be delivered by `phdp-shared-secrets` with annotations (number - comment)

---
JWTPublicKey: (1 - the name must match `mapstructure:"jwtpublickey"`) (2 - represented by struct JWTPublicKeyConfig)
- name: "generated-for-test-1" (3 - represented by struct JWTPublicKey)
  comment: "generated with: openssl rsa -in private.pem -pubout -outform PEM -out public.pem"
  not_before: "2020-01-01"
  not_after: "2022-01-01"
  key: |
    -----BEGIN PUBLIC KEY-----
    MIIBIjANBg...
    -----END PUBLIC KEY-----
- name: "dev-vault-previous"
  comment: "copied from somewhere"
  not_before: "2020-01-01"
  not_after: "2022-01-01"
  key: |
    -----BEGIN PUBLIC KEY-----
    MIICIjAN...
    -----END PUBLIC KEY-----
- name: "integrationTest"
  comment: "copied from this repo: /test/config"
  not_before: "2020-01-01"
  not_after: "2022-01-01"
  key: |
    -----BEGIN PUBLIC KEY-----
    MIICI...
    -----END PUBLIC KEY-----
--- (4 - can be a separate yaml file in different path - if split into many files, we need to use many viper instances to parse and then use 'ViperConfig.Merge()' )
JWTPrivateKey: (5 - the name must match `mapstructure:"jwtprivatekey"`) (6 - represented by struct JWTPrivateKeyConfig)
- name: "generated-for-test-1"  (7 - represented by struct JWTPrivateKey)
  comment: "generated with: openssl genrsa 2048 -out private.pem"
  enabled: true
  created_at: "unknown"
  author: "Jenkins"
  key: |
    -----BEGIN RSA PRIVATE KEY-----
    ...
    -----END RSA PRIVATE KEY-----
- name: "the-same-but-copied"
  comment: ""
  enabled: false
  created_at: "unknown"
  author: "Jenkins"
  key: |
    -----BEGIN RSA PRIVATE KEY-----
    ...
    -----END RSA PRIVATE KEY-----
*/

// JWTPublicKey represents an object storing information about a single public key entry from Viper config.yaml
type JWTPublicKey struct {
	Name      string
	Comment   string
	NotBefore string
	NotAfter  string
	Key       *rsa.PublicKey
}

func (j *JWTPublicKey) String() string {
	return fmt.Sprintf("Name: %s, Comment: %s, Key: (hidden), NotBefore: %s, NotAfter: %s",
		j.Name, j.Comment, j.NotBefore, j.NotAfter)
}

// JWTPrivateKey represents an object storing information about a single private key entry from Viper config.yaml
type JWTPrivateKey struct {
	Name      string
	Comment   string
	Enabled   bool
	CreatedAt string
	Author    string
	Key       *rsa.PrivateKey
}

func (j *JWTPrivateKey) String() string {
	return fmt.Sprintf("Name: %s, Comment: %s, Key: (hidden), Enabled: %t, CreatedAt: %s, Author: %s",
		j.Name, j.Comment, j.Enabled, j.CreatedAt, j.Author)
}

var _ JWTKeyProvider = (*ViperConfig)(nil)

type JWTKeyProvider interface {
	JWTPublicKeys() ([]JWTPublicKey, error)
	JWTPrivateKey() (JWTPrivateKey, error)
}

var _ JWTKeyProvider = (*ViperConfig)(nil)

// JWTPublicKeyConfig models the structure of the Viper config fragment responsible for defining JWT public keys
type JWTPublicKeyConfig struct {
	Entries []struct {
		Name      string `mapstructure:"name"`
		Comment   string `mapstructure:"comment"`
		Key       string `mapstructure:"key"`
		NotBefore string `mapstructure:"not_before"`
		NotAfter  string `mapstructure:"not_after"`
	} `mapstructure:"jwtpublickey"`
}

func (j *JWTPublicKeyConfig) String() string {
	s := ""
	for i, val := range j.Entries {
		s += fmt.Sprintf("[%d] Name: %s, Comment: %s, Key: <truncated>, NotBefore: %s, NotAfter: %s",
			i, val.Name, val.Comment, val.NotBefore, val.NotAfter)
	}
	return s
}

// JWTPrivateKeyConfig models the structure of the Viper config fragment responsible for defining JWT private keys
type JWTPrivateKeyConfig struct {
	Entries []struct {
		Name      string `mapstructure:"name"`
		Comment   string `mapstructure:"comment"`
		Key       string `mapstructure:"key"`
		Enabled   bool   `mapstructure:"enabled"`
		CreatedAt string `mapstructure:"created_at"`
		Author    string `mapstructure:"author"`
	} `mapstructure:"jwtprivatekey"`
}

func (j *JWTPrivateKeyConfig) String() string {
	s := ""
	for i, val := range j.Entries {
		s += fmt.Sprintf("[%d] Name: %s, Comment: %s, Key: <hidden>, Enabled: %t, CreatedAt: %s, Author: %s",
			i, val.Name, val.Comment, val.Enabled, val.CreatedAt, val.Author)
	}
	return s
}

type ViperConfigOpt func(vp *ViperConfig)

func WithConfigSource(configSrc io.Reader) ViperConfigOpt {
	return func(vp *ViperConfig) {
		vp.useCustomSource = true
		vp.configSrc = configSrc
	}
}

// WithConfigFilePaths sets the paths where the config files could be located
// the values should be directory paths. The name of the files are configured in other options
// only the first existing valid file out of the input files will be read
func WithConfigFilePaths(configPaths ...string) ViperConfigOpt {
	return func(vp *ViperConfig) {
		vp.configPaths = append(vp.configPaths, configPaths...)
	}
}

// WithConfigFileName sets the config file name (without extension) to read from - default is 'config'
func WithConfigFileName(configFileName string) ViperConfigOpt {
	return func(vp *ViperConfig) {
		vp.configFileName = configFileName
	}
}

// WithConfigFormat sets the config format to read from - default is yaml
func WithConfigFormat(configFormat string) ViperConfigOpt {
	return func(vp *ViperConfig) {
		vp.configFormat = configFormat
	}
}
func WithServiceName(s string) ViperConfigOpt {
	return func(vp *ViperConfig) {
		vp.svcName = s
	}
}

// WithAutoBootstrap setting this to false disables WithWatchChanges(true)
func WithAutoBootstrap(b bool) ViperConfigOpt {
	return func(vp *ViperConfig) {
		vp.autoBootstrap = b
	}
}
func WithWatchChanges(b bool) ViperConfigOpt {
	return func(vp *ViperConfig) {
		vp.watchChanges = b
	}
}

// WithDefaultLogger sets the log.Logger that will be used in the dynamic.DefaultLogger
func WithDefaultLogger(log *log.Logger) ViperConfigOpt {
	return func(vp *ViperConfig) {
		vp.log = log
	}
}

// WithLogger configures the custom logger. Will replace the dynamic.DefaultLogger
func WithLogger(logger Logger) ViperConfigOpt {
	return func(vp *ViperConfig) {
		vp.logger = logger
		vp.loggerSet = true
	}
}

func WithViperVerbose(b bool) ViperConfigOpt {
	return func(vp *ViperConfig) {
		vp.viperVerbose = b
	}
}

// ViperConfig represents Viper configuration for shared secrets and configs
// due to usage of sync.Once must not be copied after boostraping
type ViperConfig struct {
	// name of the ViperConfig entity
	name string
	// serviceName is the identifier of the service using ViperConfig entity
	svcName string
	// Error stores config bootstraping and loading errors
	Error error

	// logger will be used to log events, expose errors, and allow simple debugging
	logger Logger
	// flag to indicate whether the user has provided custom logger
	loggerSet bool
	// log will produce log messages (if set) for the default logger
	log *log.Logger

	// viperVerbose is a flag to set viper in verbose mode
	viperVerbose bool

	// bootstrapOnce ensures that the bootstrap function is called at most once
	bootstrapOnce sync.Once

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
	// stores the parsed pubKeys from the config
	pubKeys []JWTPublicKey
	// stores the parsed privKeys from the config
	privKeys []JWTPrivateKey
	// stores the active (enabled) private key
	activePrivKey *JWTPrivateKey
	// v stores a viper instance (default is global)
	v *viper.Viper
}

// NewViperConfig provides configuration using Viper
// Developers using autobootstrap should check Error field for possible bootstrapping errors
func NewViperConfig(name string, opts ...ViperConfigOpt) *ViperConfig {
	vp := &ViperConfig{
		name:            name,
		svcName:         "unknown-svc",
		configSrc:       nil,
		configPaths:     []string{"."},
		configFileName:  "config",
		configFormat:    "yaml", // default is yaml
		viperVerbose:    false,
		useCustomSource: false,
		autoBootstrap:   true,
		watchChanges:    true,
		pubKeys:         make([]JWTPublicKey, 0),
		privKeys:        make([]JWTPrivateKey, 0),
		activePrivKey:   nil,
		v:               viper.New(),
		Error:           nil,
		loggerSet:       false,
		logger:          nil,
		log:             nil,
	}
	for _, opt := range opts {
		opt(vp)
	}
	// set Viper verbosity as soon as possible to see potential bootstrapping issues
	vp.v.Set("Verbose", vp.viperVerbose)

	// set default event handlers - only if no custom handlers were configured by the user
	defaultLogger := NewDefaultLogger(vp.name, vp.log)
	if vp.logger == nil && !vp.loggerSet {
		vp.logger = defaultLogger
	}

	if vp.autoBootstrap {
		if err := vp.Bootstrap(); err != nil {
			vp.Error = fmt.Errorf("bootstrap error: %w", err)
		}
	}
	if vp.watchChanges {
		vp.WatchConfig()
	}
	AddDynamicPkgMetrics(vp.svcName, vp.v.ConfigFileUsed(), vp.autoBootstrap)
	return vp
}

func (vc *ViperConfig) Name() string {
	return vc.name
}

// MergeAndDisableHotReload merges values from another viper instance into the current one
// to be used when configs are provided in multiple files, as one viper instance can read config only from a single file
// WARNING: MERGING DISABLES HOT-RELOAD of the 'other' config - use with caution!
// the name MergeAndDisableHotReload shall ensure that users understand the consequences of using it
func (vc *ViperConfig) MergeAndDisableHotReload(other *ViperConfig) error {
	orgName := vc.name
	_ = other.v.ReadInConfig()
	vc.logger.LogDebug("Merging '%s' into '%s'", other.name, vc.name)
	err := vc.v.MergeConfigMap(other.v.AllSettings())
	vc.name = fmt.Sprintf("merged-%s+%s", orgName, other.name)
	vc.logger.SetName(vc.name)
	vc.logger.LogDebug("Merge finished. Setting keys (secrets): %s", strings.Join(vc.v.AllKeys(), ","))
	vc.handleConfigChange(fsnotify.Event{Name: "ViperConfig.Merge event"})
	return err
}

// Bootstrap initializes ViperConfig form a config.yaml file being read from configurable locations
// Providing config in form of a file is a prerequisite for using 'WatchConfig' and automated loading of changes
func (vc *ViperConfig) Bootstrap() (err error) {
	vc.bootstrapOnce.Do(func() {
		err = vc.bootstrap()
	})
	return err
}
func (vc *ViperConfig) bootstrap() error {
	vc.v.SetConfigType(vc.configFormat)
	// do not read from files - use the reader instead (useful for tests)
	if vc.useCustomSource {
		err := vc.v.ReadConfig(vc.configSrc)
		if err == nil {
			vc.logger.LogDebug("Boostrap finished. Setting keys (secrets): %s", strings.Join(vc.v.AllKeys(), ","))
			MetricBootstrapStatus.WithLabelValues(string(vc.name), string(vc.svcName)).Set(1)
			vc.handleConfigChange(fsnotify.Event{Name: "ViperConfig.Bootstrap event with useCustomSource"})
		} else {
			vc.logger.LogDebug("Boostrap failed. Error: %s", err.Error())
			MetricBootstrapStatus.WithLabelValues(string(vc.name), string(vc.svcName)).Set(0)
		}
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
			MetricBootstrapStatus.WithLabelValues(string(vc.name), string(vc.svcName)).Set(0)
			return fmt.Errorf("cannot find config file: %s \n", vc.v.ConfigFileUsed())
		} else {
			// Config file was found but another error was produced
			MetricBootstrapStatus.WithLabelValues(string(vc.name), string(vc.svcName)).Set(0)
			return fmt.Errorf("error reading config file: %w \n", err2)
		}
	}

	MetricBootstrapStatus.WithLabelValues(string(vc.name), string(vc.svcName)).Set(1)
	vc.logger.LogDebug("Boostrap finished. Setting keys (secrets): %s", strings.Join(vc.v.AllKeys(), ","))
	vc.handleConfigChange(fsnotify.Event{Name: "ViperConfig.Bootstrap event"})
	return nil
}

// logSummary produces a string that summarizes the status of loading all secrets.
// It may be disturbing to see that JWT priv key failed to load in a service that does not require it
// but I find it necessary to quickly spot potential problems.
// Any time, a svc developer can set debugLogger to nil and disable these messages
func (vc *ViperConfig) logSummary() string {
	summary := make([]string, 0)

	// JWT PUBLIC KEYS
	summary = append(summary, fmt.Sprintf("loaded %d JWT public keys", len(vc.pubKeys)))
	for i, pubKey := range vc.pubKeys {
		summary = append(summary, fmt.Sprintf("public key [%d] name: '%s'", i, pubKey.Name))
	}

	// JWT PRIVATE KEYS
	summary = append(summary, fmt.Sprintf("loaded %d JWT private keys", len(vc.privKeys)))
	for i, privKey := range vc.privKeys {
		summary = append(summary, fmt.Sprintf("private key [%d] name: '%s', enabled?: '%t'", i, privKey.Name, privKey.Enabled))
	}
	return strings.Join(summary, "; ")
}

// PrintSummary prints the summary generated by logSummary() using the debugLogger
func (vc *ViperConfig) PrintSummary() {
	vc.logger.LogDebug(vc.logSummary())
}

// WatchConfig enables a listener that will execute `update<secret-name>` automatically on every secret/config change
func (vc *ViperConfig) WatchConfig() {
	vc.v.WatchConfig()
	// Define what to do when config changes
	vc.v.OnConfigChange(func(e fsnotify.Event) {
		vc.handleConfigChange(e)
	})
}

// handleConfigChange handles configuration change triggered automatically by Viper
func (vc *ViperConfig) handleConfigChange(e ...fsnotify.Event) {
	if fatalErr, infoErrs := vc.updatePublicKeys(); fatalErr != nil {
		vc.logger.LogError(fatalErr, "fatal error")
		MetricConfigHotReloads.WithLabelValues(string(vc.name), string(vc.svcName), "JWT-pub-keys", "failed").Inc()
	} else if len(infoErrs) > 0 {
		for _, infoErr := range infoErrs {
			vc.logger.LogError(infoErr, "partial error")
		}
		MetricConfigHotReloads.WithLabelValues(string(vc.name), string(vc.svcName), "JWT-pub-keys", "ok-but-parts-failed").Inc()
	} else {
		vc.logger.LogInfo("successfully read config for JWT public keys")
		MetricConfigHotReloads.WithLabelValues(string(vc.name), string(vc.svcName), "JWT-pub-keys", "ok").Inc()
	}

	if fatalErr, infoErrs := vc.updatePrivateKeys(); fatalErr != nil {
		vc.logger.LogError(fatalErr, "fatal error")
		MetricConfigHotReloads.WithLabelValues(string(vc.name), string(vc.svcName), "JWT-priv-keys", "failed").Inc()
	} else if len(infoErrs) > 0 {
		for _, infoErr := range infoErrs {
			vc.logger.LogError(infoErr, "partial error")
			MetricConfigHotReloads.WithLabelValues(string(vc.name), string(vc.svcName), "JWT-priv-keys", "ok-but-parts-failed").Inc()
		}
	} else {
		vc.logger.LogInfo("successfully read config for JWT private keys")
		MetricConfigHotReloads.WithLabelValues(string(vc.name), string(vc.svcName), "JWT-priv-keys", "ok").Inc()
	}

	vc.logger.LogDebug(vc.logSummary())
}

// PublicKeys wraps JWTPublicKeys for backwards compatibility
// may be used in hades or handshake
func (vc *ViperConfig) PublicKeys() ([]JWTPublicKey, error) {
	return vc.JWTPublicKeys()
}

// JWTPublicKeys returns all valid public keys to the user
// external interface to be used by the user
func (vc *ViperConfig) JWTPublicKeys() ([]JWTPublicKey, error) {
	if len(vc.pubKeys) > 0 {
		return vc.pubKeys, nil
	}
	// no need to return infoErrs to the user, because the correctly parsed parts may be enough to proceed
	// if not, we will return 0 keys
	if fatalErr, _ := vc.updatePublicKeys(); fatalErr != nil {
		return vc.pubKeys, fatalErr
	}
	if len(vc.pubKeys) == 0 {
		return vc.pubKeys, fmt.Errorf("found 0 valid public keys")
	}
	return vc.pubKeys, nil
}

// updatePublicKeys is called by the watcher on each change in the configmap/secret
// this runs in the background and it should do the heavy-lifting (parsing, validating, etc.)
// it must set vc.pubKeys array
// returns fatal error when entire processing of the config has failed
// returns partial error(s) if only parts of the config could not be read
func (vc *ViperConfig) updatePublicKeys() (fatal error, infoErrs []error) {
	keys, fatalErr, infoErrs := parsePublicKeys(vc.v, vc.logger)
	if fatalErr != nil {
		MetricSecretsLoaded.WithLabelValues(string(vc.name), string(vc.svcName), "JWT-pub-keys").Set(float64(0))
		return fatalErr, infoErrs
	}
	MetricSecretsLoaded.WithLabelValues(string(vc.name), string(vc.svcName), "JWT-pub-keys").Set(float64(len(keys) - len(infoErrs)))
	vc.pubKeys = keys // store the keys to speed-up access in the future
	return nil, infoErrs
}

// parsePublicKeys un-marshals viper config and parses PEM public keys from it
// The Keys and their metadata are stored in the PublicKey struct
// returns fatal error when entire processing of the config has failed
// returns partial error(s) if only parts of the config could not be read
func parsePublicKeys(vp *viper.Viper, logger Logger) (keys []JWTPublicKey, fatal error, infoErrs []error) {
	keys = make([]JWTPublicKey, 0)
	infoErrs = make([]error, 0)

	var config JWTPublicKeyConfig
	err := vp.Unmarshal(&config)
	if err != nil {
		return keys, fmt.Errorf("unable to decode into struct: %w", err), infoErrs
	}
	logger.LogDebug("parsePublicKeys. Parsed: %s", config.String())
	for _, entry := range config.Entries {
		entryValid := true
		if entry.Key == "" {
			infoErrs = append(infoErrs, fmt.Errorf("error parsing public key entry '%s': key is empty", entry.Name))
			entryValid = false
		}
		publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(entry.Key))
		if err != nil {
			infoErrs = append(infoErrs, fmt.Errorf("error parsing public key contents '%s': %w", entry.Name, err))
			entryValid = false
		}
		if entryValid {
			logger.LogDebug("parsePublicKeys. Valid public key (name: '%s')", entry.Name)
			keys = append(keys, JWTPublicKey{Key: publicKey, Name: entry.Name, Comment: entry.Comment, NotBefore: entry.NotBefore, NotAfter: entry.NotAfter})
		}
	}
	return keys, nil, infoErrs
}

// JWTPrivateKey returns the first enabled private key from the config
// external interface to be used by the user
func (vc *ViperConfig) JWTPrivateKey() (JWTPrivateKey, error) {
	if vc.activePrivKey == nil {
		return JWTPrivateKey{}, fmt.Errorf("found 0 enabled private keys")
	}
	return *vc.activePrivKey, nil
}

// updatePrivateKeys is called by the watcher on each change in the configmap/secret
// this runs in the background and it should do the heavy-lifting (parsing, validating, etc.)
// it must set vc.activePrivKey to a valid value or leave it nil
// returns fatal error when entire processing of the config has failed
// returns partial error(s) if only parts of the config could not be read
func (vc *ViperConfig) updatePrivateKeys() (fatal error, infoErr []error) {
	keys, fatalErr, infoErrs := parsePrivateKeys(vc.v, vc.logger)
	if fatalErr != nil {
		MetricSecretsLoaded.WithLabelValues(string(vc.name), string(vc.svcName), "JWT-priv-keys").Set(0)
		return fatalErr, infoErr
	}

	vc.privKeys = keys // store the key to speed-up access in the future

	currentlyActiveKey := "nil"
	if vc.activePrivKey != nil {
		currentlyActiveKey = vc.activePrivKey.Name
	}
	vc.logger.LogDebug("pre-update active private key: '%s'", currentlyActiveKey)
	vc.logger.LogDebug("starting update of private keys")
	for i, key := range keys {
		if key.Enabled {
			vc.activePrivKey = &key
			vc.logger.LogDebug("Setting private key (index %d name '%s') as active", i, key.Name)
			break
		}
	}
	currentlyActiveKey = "nil"
	if vc.activePrivKey != nil {
		currentlyActiveKey = vc.activePrivKey.Name
	}
	vc.logger.LogDebug("post-update active private key: '%s'", currentlyActiveKey)
	MetricSecretsLoaded.WithLabelValues(string(vc.name), string(vc.svcName), "JWT-priv-keys").Set(float64(len(keys) - len(infoErrs)))
	return nil, infoErrs
}

// parsePrivateKey un-marshals viper config into JWTPrivateKeyConfig
// next, it verifies each entry in the JWTPrivateKeyConfig: does it have key? is the key parsable into PEM private key?
// all valid entries are returned as []JWTPrivateKey
// critical errors (cannot parse anything) are returned as 'fatal'
// partial errors (cannot parse one entry) are returned as 'infoErr'
func parsePrivateKeys(vp *viper.Viper, logger Logger) (keys []JWTPrivateKey, fatal error, infoErrs []error) {
	keys = make([]JWTPrivateKey, 0)
	infoErrs = make([]error, 0)
	var config JWTPrivateKeyConfig
	err := vp.Unmarshal(&config)
	if err != nil {
		return keys, fmt.Errorf("unable to decode into struct: %w", err), nil
	}
	logger.LogDebug("parsePrivateKeys. Parsed: %s", config.String())
	for _, entry := range config.Entries {
		entryValid := true
		logger.LogDebug("parsePrivateKeys. Proceesing entry: %s", entry.Name)
		if entry.Key == "" {
			// it is okay to skip parsing of one broken entry if the rest of the config is valid
			infoErrs = append(infoErrs, fmt.Errorf("error parsing private key entry '%s': key is empty", entry.Name))
			entryValid = false
		}
		privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(entry.Key))
		if err != nil {
			// it is okay to skip parsing of one broken key if the rest of the config is valid
			infoErrs = append(infoErrs, fmt.Errorf("error parsing private key contents '%s': %w", entry.Name, err))
			entryValid = false
		}
		logger.LogDebug("parsePrivateKeys. Finished processing private key '%s', isValid = %t", entry.Name, entryValid)
		if entryValid {
			keys = append(keys, JWTPrivateKey{
				Key:       privateKey,
				Name:      entry.Name,
				Comment:   entry.Comment,
				Enabled:   entry.Enabled,
				CreatedAt: entry.CreatedAt,
				Author:    entry.Author,
			})
		}
	}
	return keys, nil, infoErrs
}
