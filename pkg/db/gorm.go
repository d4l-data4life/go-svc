package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/d4l-data4life/go-svc/pkg/logging"
	"github.com/d4l-data4life/go-svc/pkg/migrate"
)

var (
	db *gorm.DB
)

const (
	numConnectAttempts uint   = 7 // with expTimeBackoff 2^7 = 2 minutes + eps
	migrationsTable    string = "migrations"
	migrationsSource   string = "sql"
)

// define general error messages
var (
	ErrDBConnection   = errors.New("database connection error")
	ErrDBMigration    = errors.New("database migration error")
	ErrRunCtxCanceled = errors.New("run context canceled by the user")
)

// Initialize connects to the Database and migrates the schema
// nolint: funlen
func Initialize(runCtx context.Context, opts *ConnectionOptions) <-chan struct{} {
	dbUp := make(chan struct{})
	// goroutine to establish connection including retries
	go func() {
		defer close(dbUp)
		if opts == nil {
			dbUp <- struct{}{} // nothing to be done, so show success
			return
		}
		connectString := ConnectString(opts)
		connectFn := func() (*gorm.DB, error) { return DefaultPostgresDriver(connectString, opts) }

		// retries as long as err != nil
		conn, err := retryExponential(runCtx, numConnectAttempts, 1*time.Second, connectFn)
		if err != nil {
			logging.LogErrorf(err, "Could not connect to the database")
			return
		}
		logging.LogInfof("connection to the database succeeded")

		// goroutine to close DB connection when run context is canceled
		go func() {
			<-runCtx.Done()
			logging.LogInfof("run context canceled, closing database connection")
			defer Close()
			defer logging.LogInfof("database connection closed")
		}()

		err = runMigrationWithOptions(runCtx, conn, opts)
		if err != nil {
			if opts.MigrationHaltOnError || opts.VersionedMigrationFunc != nil {
				if sqlDB, closeErr := conn.DB(); closeErr == nil {
					_ = sqlDB.Close()
				}
				logging.LogErrorf(err, "database migration failed - aborting")
				return
			}
			logging.LogWarningf(err, "database migration failed - continuing")
		}
		logging.LogInfof("database migration finished")

		db = conn
		sqlDB, err := conn.DB()
		if err != nil {
			logging.LogErrorf(err, "Could not get sql DB")
			return
		}
		sqlDB.SetConnMaxLifetime(opts.MaxConnectionLifetime)
		sqlDB.SetMaxIdleConns(opts.MaxIdleConnections)
		sqlDB.SetMaxOpenConns(opts.MaxOpenConnections)

		logging.LogInfof("database connection is up and configured")

		if opts.EnableInstrumentation {
			err = db.Use(NewInstrumenter())
			if err != nil {
				logging.LogErrorf(err, "Could not register instrumenter plugin")
				return
			}
			logging.LogInfof("database instrumenter plugin registered")
		}
		dbUp <- struct{}{} // notify that DB is up now
	}()

	return dbUp
}

// Get returns a handle to the DB object
func Get() *gorm.DB {
	if db == nil {
		logging.LogErrorf(ErrDBConnection, "Get() - db handle is nil")
	}
	return db
}

func Ping() error {
	sqlDB, err := db.DB()
	if err != nil {
		logging.LogErrorf(err, "error getting sql DB")
		return err
	}
	return sqlDB.Ping()
}

// Close closes the DB connecton
func Close() {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			logging.LogErrorf(err, "error getting sql DB")
			return
		}

		err = sqlDB.Close()
		if err != nil {
			logging.LogErrorf(err, "error closing DB")
		}
	}
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

// runMigrationWithOptions keeps legacy behavior separate from the opt-in versioned runner.
func runMigrationWithOptions(ctx context.Context, conn *gorm.DB, opts *ConnectionOptions) error {
	if conn == nil {
		return ErrDBConnection
	}
	if opts.MigrationFunc != nil && opts.VersionedMigrationFunc != nil {
		return errors.New("both MigrationFunc (legacy) and VersionedMigrationFunc are set; configure only one")
	}
	if opts.MigrationVersion == 0 {
		if opts.VersionedMigrationFunc != nil {
			return opts.VersionedMigrationFunc(conn.WithContext(ctx), 0)
		}
		if opts.MigrationFunc != nil {
			return opts.MigrationFunc(conn.WithContext(ctx))
		}
		return nil
	}
	if opts.VersionedMigrationFunc != nil {
		return runMigrationVersioned(ctx, conn, opts)
	}
	sqlDB, err := conn.DB()
	if err != nil {
		return err
	}
	return runMigrationLegacy(sqlDB, conn, opts.MigrationFunc, opts.MigrationVersion, opts.MigrationStartFromZero)
}

func runMigrationLegacy(sqlDB *sql.DB, conn *gorm.DB, legacyFn MigrationFunc, migrationVersion uint, startFromZero bool) error {
	// Preserve legacy behavior: AutoMigrate once, then SQL migrations via golang-migrate.
	if legacyFn != nil {
		if err := legacyFn(conn); err != nil {
			return err
		}
	}

	if migrationVersion == 0 {
		return nil
	}
	migration := migrate.NewMigration(sqlDB, migrationsSource, migrationsTable, logging.Logger())
	return migration.MigrateDB(context.Background(), migrationVersion, startFromZero)
}
