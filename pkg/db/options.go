package db

import (
	"database/sql"
	"sort"
	"time"

	"github.com/DATA-DOG/go-txdb"
	"github.com/jinzhu/gorm"
)

type MigrationFunc func(do *gorm.DB) error
type DriverFunc func(connectString string) (*gorm.DB, error)

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
	User                  string
	Password              string
	SSLMode               string
	MigrationFunc         MigrationFunc
	DriverFunc            DriverFunc
	EnableInstrumentation bool
}

type ConnectionOption func(*ConnectionOptions)

func WithDefaults() ConnectionOption {
	return func(c *ConnectionOptions) {
		c.Debug = false
		c.MaxConnectionLifetime = 5 * time.Minute
		c.MaxIdleConnections = 3
		c.MaxOpenConnections = 6
		c.Host = "localhost"
		c.Port = "5432"
		c.SSLMode = "disable"
		c.EnableInstrumentation = true
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
func WithMigrationFunc(fn MigrationFunc) ConnectionOption {
	return func(c *ConnectionOptions) {
		c.MigrationFunc = fn
	}
}

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

// DefaultPostgresDriver defines the default DB driver
// To be used in options as argument for db.WithDriverFunc()
// Is default, hence will be used when db.WithDriverFunc() is not used in options
// Example: db.InitializeTestPostgres(db.NewConnection(db.WithDriverFunc(db.DefaultPostgresDriver)))
func DefaultPostgresDriver(connectString string) (*gorm.DB, error) {
	return gorm.Open("postgres", connectString)
}

// TXDBPostgresDriver defines the TXDB DB driver which treats every operation as transaction, and does rollback on disconnect.
// To be used in options as argument for db.WithDriverFunc()
// Example: db.InitializeTestPostgres(db.NewConnection(db.WithDriverFunc(db.TXDBPostgresDriver)))
func TXDBPostgresDriver(connectString string) (*gorm.DB, error) {
	drivers := sql.Drivers()
	i := sort.SearchStrings(drivers, "txdb")
	if i >= len(drivers) || drivers[i] != "txdb" {
		txdb.Register("txdb", "postgres", connectString)
	}
	return gorm.Open("postgres", "txdb", "tx_1")
}
