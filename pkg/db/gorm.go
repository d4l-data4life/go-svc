package db

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/pkg/errors"

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

		err = runMigration(conn, opts.VersionedMigrationFunc, opts.MigrationVersion, opts.MigrationStartFromZero)
		if err != nil {
			if opts.MigrationHaltOnError {
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

// runMigration Executes Migrations on the database
func runMigration(conn *gorm.DB, migFn VersionedMigrationFunc, migrationVersion uint, startFromZero bool) error {
	if conn == nil {
		logging.LogErrorf(ErrDBConnection, "MigrateDB() - db handle is nil")
		return ErrDBConnection
	}
	sqlDB, err := conn.DB()
	if err != nil {
		logging.LogErrorf(err, "error getting sql DB")
		return err
	}
	if migrationVersion == 0 {
		if migFn != nil {
			return migFn(conn, 0)
		}
		return nil
	}

	ctx := context.Background()
	migration := migrate.NewMigration(sqlDB, migrationsSource, migrationsTable, logging.Logger())

	if err := migration.ExecuteSetup(ctx); err != nil {
		return err
	}
	if err := migration.ExecuteFdwUp(ctx); err != nil {
		return err
	}
	defer func() {
		if err := migration.ExecuteFdwDown(ctx); err != nil {
			logging.LogErrorf(err, "error executing fdw down script")
		}
	}()

	mpg, err := migration.MigrateInstance()
	if err != nil {
		return err
	}

	currentVersionRaw, dirty, err := mpg.Version()
	if err != nil {
		if stderrors.Is(err, migrate.ErrNilVersion) {
			if startFromZero {
				currentVersionRaw = 0
			} else {
				// no migration info and startFromZero disabled -> treat as already at target
				// nolint: gosec
				if err := migrate.SetVersion(mpg, migrationVersion); err != nil {
					return err
				}
				return nil
			}
		} else {
			return err
		}
	}
	currentVersion := uint(currentVersionRaw)
	if dirty {
		return errors.Errorf("database migration is dirty at version %d", currentVersion)
	}

	for version := currentVersion + 1; version <= migrationVersion; version++ {
		if _, err := migration.ExecuteBeforeUp(ctx, version); err != nil {
			logging.LogErrorf(err, "error running before migration for version %d", version)
			return err
		}

		if migFn != nil {
			if err := migFn(conn, version); err != nil {
				logging.LogErrorf(err, "error running auto migration for version %d", version)
				return err
			}
		}

		if _, err := migration.ExecuteAfterUp(ctx, version); err != nil {
			logging.LogErrorf(err, "error running after migration for version %d", version)
			return err
		}

		// Record version after full before/auto/after sequence.
		// nolint: gosec
		if err := migrate.SetVersion(mpg, version); err != nil {
			logging.LogErrorf(err, "error setting migration version to %d", version)
			return err
		}

		logging.LogInfof("migration for version %d executed successfully", version)
	}

	return nil
}
