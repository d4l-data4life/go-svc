package db

import (
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/DATA-DOG/go-txdb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
)

type MigrationFunc func(do *gorm.DB) error
type DriverFunc func(connectString string, opts *ConnectionOptions) (*gorm.DB, error)

func NewConnection(opts ...ConnectionOption) *ConnectionOptions {
	o := &ConnectionOptions{}
	// apply defaults
	WithDefaults()(o)
	// apply user-provided options
	for _, opt := range opts {
		opt(o)
	}
	return o
}

type ConnectionOptions struct {
	Debug                 bool
	MaxConnectionLifetime time.Duration
	MaxIdleConnections    int
	MaxOpenConnections    int
	Host                  string
	Port                  string
	DatabaseName          string
	DatabaseSchema        string
	User                  string
	Password              string
	SSLMode               string
	MigrationVersion      uint
	MigrationHaltOnError  bool
	// SSLRootCertPath represents path to a file containing the root-CA used for Postgres server identity validation
	// The cert is provided by Jenkins on build under default path "/root.ca.pem"
	SSLRootCertPath        string
	MigrationFunc          MigrationFunc
	DriverFunc             DriverFunc
	EnableInstrumentation  bool
	LoggerConfig           logger.Config
	SkipDefaultTransaction bool
}

type ConnectionOption func(*ConnectionOptions)

func WithDefaults() ConnectionOption {
	return func(c *ConnectionOptions) {
		c.Debug = false
		c.MaxConnectionLifetime = 5 * time.Minute
		c.MaxIdleConnections = 3
		c.MaxOpenConnections = 6
		c.DatabaseSchema = "public"
		c.Host = "localhost"
		c.Port = "5432"
		c.SSLMode = "verify-full"
		c.SSLRootCertPath = "/root.ca.pem"
		c.EnableInstrumentation = true
		c.LoggerConfig = logger.Config{
			SlowThreshold:             500 * time.Millisecond,
			IgnoreRecordNotFoundError: true,
			LogLevel:                  logger.Silent,
		}
		c.SkipDefaultTransaction = false
	}
}

func WithDebug(value bool) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.Debug = value
	}
}

func WithMaxConnectionLifetime(value time.Duration) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.MaxConnectionLifetime = value
	}
}
func WithMaxIdleConnections(value int) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.MaxIdleConnections = value
	}
}
func WithMaxOpenConnections(value int) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.MaxOpenConnections = value
	}
}
func WithHost(value string) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.Host = value
	}
}
func WithPort(value string) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.Port = value
	}
}
func WithDatabaseName(value string) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.DatabaseName = value
	}
}
func WithDatabaseSchema(value string) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.DatabaseSchema = value
	}
}
func WithUser(value string) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.User = value
	}
}
func WithPassword(value string) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.Password = value
	}
}
func WithSSLMode(value string) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.SSLMode = value
	}
}
func WithSSLRootCertPath(value string) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.SSLRootCertPath = value
	}
}
func WithMigrationFunc(fn MigrationFunc) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.MigrationFunc = fn
	}
}

func WithMigrationVersion(version uint) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.MigrationVersion = version
	}
}

func WithMigrationHaltOnError(haltOnError bool) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.MigrationHaltOnError = haltOnError
	}
}

// WithDriverFunc is used to overwrite the DB driver for testing
func WithDriverFunc(fn DriverFunc) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.DriverFunc = fn
	}
}

func WithEnableInstrumentation(value bool) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.EnableInstrumentation = value
	}
}

func WithLoggerConfig(conf logger.Config) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.LoggerConfig = conf
	}
}

func WithLogLevel(logLevel LogLevel) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.LoggerConfig.LogLevel = logger.LogLevel(logLevel)
	}
}

func WithSkipDefaultTransaction(value bool) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.SkipDefaultTransaction = value
	}
}

// ConnectString reads connect options and compiles them to string form
func ConnectString(opts *ConnectionOptions) string {
	connectString := fmt.Sprintf("host=%s port=%s dbname=%s sslmode=%s",
		opts.Host, opts.Port, opts.DatabaseName, opts.SSLMode)

	if (opts.SSLMode == "verify-ca" || opts.SSLMode == "verify-full") && opts.SSLRootCertPath != "" {
		connectString += fmt.Sprintf(" sslrootcert=%s", opts.SSLRootCertPath)
	}

	secretString := fmt.Sprintf(" user=%s password=%s", opts.User, opts.Password)

	logging.LogDebugf("Attempting to connect to DB: %s", connectString)

	return connectString + secretString
}

// DefaultPostgresDriver defines the default DB driver
// To be used in options as argument for db.WithDriverFunc()
// Is default, hence will be used when db.WithDriverFunc() is not used in options
// Example: db.InitializeTestPostgres(db.NewConnection(db.WithDriverFunc(db.DefaultPostgresDriver)))
func DefaultPostgresDriver(connectString string, opts *ConnectionOptions) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(connectString), &gorm.Config{
		Logger:                 NewLogger(opts.LoggerConfig),
		SkipDefaultTransaction: opts.SkipDefaultTransaction,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   fmt.Sprintf("%s.", opts.DatabaseSchema),
			SingularTable: false,
		},
	})
}

// TXDBPostgresDriver defines the TXDB DB driver which treats every operation as transaction, and does rollback on disconnect.
// To be used in options as argument for db.WithDriverFunc()
// Example: db.InitializeTestPostgres(db.NewConnection(db.WithDriverFunc(db.TXDBPostgresDriver)))
func TXDBPostgresDriver(connectString string, opts *ConnectionOptions) (*gorm.DB, error) {
	drivers := sql.Drivers()
	i := sort.SearchStrings(drivers, "txdb")
	if i >= len(drivers) || drivers[i] != "txdb" {
		txdb.Register("txdb", "pgx", connectString)
	}
	return gorm.Open(postgres.New(postgres.Config{DriverName: "txdb", DSN: connectString}), &gorm.Config{
		Logger:                 NewLogger(opts.LoggerConfig),
		SkipDefaultTransaction: opts.SkipDefaultTransaction,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   fmt.Sprintf("%s.", opts.DatabaseSchema),
			SingularTable: false,
		},
	})
}

func TXDBPostgresDriverWithoutSavepoint(connectString string, opts *ConnectionOptions) (*gorm.DB, error) {
	drivers := sql.Drivers()
	i := sort.SearchStrings(drivers, "txdb")
	if i >= len(drivers) || drivers[i] != "txdb" {
		txdb.Register("txdb", "pgx", connectString, txdb.SavePointOption(nil))
	}
	return gorm.Open(postgres.New(postgres.Config{DriverName: "txdb", DSN: connectString}), &gorm.Config{
		Logger:                 NewLogger(opts.LoggerConfig),
		SkipDefaultTransaction: opts.SkipDefaultTransaction,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   fmt.Sprintf("%s.", opts.DatabaseSchema),
			SingularTable: false,
		},
	})
}
