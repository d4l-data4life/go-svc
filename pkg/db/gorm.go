package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
	"github.com/gesundheitscloud/go-svc/pkg/probe"

	// Blank import required by gorm
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	db *gorm.DB
)

const (
	numConnectAttempts uint = 7 // with expTimeBackoff 2^7 = 2 minutes + eps
)

// define general error messages
var (
	ErrDBConnection   = errors.New("database connection error")
	ErrDBMigration    = errors.New("database migration error")
	ErrRunCtxCanceled = errors.New("run context canceled by the user")
)

// Initialize connects to the Database and migrates the schema
func Initialize(runCtx context.Context, opts *ConnectionOptions) <-chan struct{} {
	dbUp := make(chan struct{})
	// goroutine to establish connection including retries
	go func() {
		defer close(dbUp)

		connectFn := func() (*gorm.DB, error) { return connect(opts) }

		// retries as long as err != nil
		conn, err := retryExponential(runCtx, numConnectAttempts, 1*time.Second, connectFn)
		if err != nil {
			logging.LogErrorf(err, "Could not connect to the database")
			return
		}
		logging.LogInfof("connection to the database succeeded")
		err = migrate(conn, opts.MigrationFunc)
		if err != nil {
			logging.LogWarningf(err, "database migration failed - continuing")
		}
		logging.LogInfof("database migration finished")
		conn.DB().SetConnMaxLifetime(opts.MaxConnectionLifetime)
		conn.DB().SetMaxIdleConns(opts.MaxIdleConnections)
		conn.DB().SetMaxOpenConns(opts.MaxOpenConnections)
		db = conn
		logging.LogInfof("database connection is up and configured")
		if opts.EnableInstrumentation {
			registerInstrumenterPlugin()
			logging.LogInfof("database instrumenter plugin registered")
		}
		dbUp <- struct{}{} // notify that DB is up now
	}()

	// goroutine to close DB connection when run context is canceled
	go func() {
		<-runCtx.Done()
		logging.LogInfof("run context canceled, closing database connection")
		defer Close(db)
		defer logging.LogInfof("database connection closed")
	}()
	return dbUp
}

// Get returns a handle to the DB object
func Get() *gorm.DB {
	if db == nil {
		logging.LogErrorf(ErrDBConnection, "Get() - db handle is nil")
		probe.Liveness().SetDead()
	}
	return db
}

// Close closes the DB connecton
func Close(conn *gorm.DB) {
	if conn != nil {
		err := conn.Close()
		if err != nil {
			logging.LogErrorf(err, "error closing DB")
		}
	}
}

// connect reads environment variables for DB configuration and attempts to open the connection
func connect(opts *ConnectionOptions) (*gorm.DB, error) {
	connectString := fmt.Sprintf("host=%s port=%s dbname=%s sslmode=%s",
		opts.Host, opts.Port, opts.DatabaseName, opts.SSLMode)

	if (opts.SSLMode == "verify-ca" || opts.SSLMode == "verify-full") && opts.SSLRootCertPath != "" {
		connectString += fmt.Sprintf(" sslrootcert=%s", opts.SSLRootCertPath)
	}

	secretString := fmt.Sprintf(" user=%s password=%s", opts.User, opts.Password)

	logging.LogDebugf("Attempting to connect to DB: %s", connectString)
	return gorm.Open("postgres", connectString+secretString)
}

// retryExponential runc function fn() as long as fn() returns no error, but maximally 'attempts' times
func retryExponential(runCtx context.Context, attempts uint, waitPeriod time.Duration, fn func() (*gorm.DB, error)) (*gorm.DB, error) {
	timeout := time.After(waitPeriod)
	logging.LogDebugf("retryExponential: timeout is %s ", waitPeriod)
	conn, err := fn()
	if err != nil {
		if attempts--; attempts > 0 {
			select {
			case <-runCtx.Done():
				return nil, ErrRunCtxCanceled
			case <-timeout:
				logging.LogDebugf("timeout event - attempts = %d ", attempts)
				return retryExponential(runCtx, attempts, 2*waitPeriod, fn)
			}
		}
		return conn, err
	}
	return conn, nil
}

// migrate Executes Migrations on the database
func migrate(conn *gorm.DB, migFn MigrationFunc) error {
	if conn == nil {
		logging.LogErrorf(ErrDBConnection, "MigrateDB() - db handle is nil")
		return ErrDBConnection
	}
	return migFn(conn)
}
