package db

import (
	"time"

	"github.com/jinzhu/gorm"
)

type MigrationFunc func(do *gorm.DB) error

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
