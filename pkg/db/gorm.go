package db

import (
	"context"
	"database/sql"
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

		err = runMigration(conn, opts.MigrationFunc, opts.VersionedMigrationFunc, opts.MigrationVersion, opts.MigrationStartFromZero)
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
func runMigration(
	conn *gorm.DB,
	legacyFn MigrationFunc,
	versionedFn VersionedMigrationFunc,
	migrationVersion uint,
	startFromZero bool,
) error {
	if conn == nil {
		logging.LogErrorf(ErrDBConnection, "MigrateDB() - db handle is nil")
		return ErrDBConnection
	}
	if legacyFn != nil && versionedFn != nil {
		return errors.New("both MigrationFunc (legacy) and VersionedMigrationFunc are set; please configure only one migration flow")
	}
	if migrationVersion == 0 {
		// No SQL migrations; run whichever automigration function is provided.
		if versionedFn != nil {
			return versionedFn(conn, 0)
		}
		if legacyFn != nil {
			return legacyFn(conn)
		}
		return nil
	}

	// Prefer explicit versioned flow when configured.
	if versionedFn != nil {
		return runMigrationVersioned(conn, versionedFn, migrationVersion, startFromZero)
	}

	sqlDB, err := conn.DB()
	if err != nil {
		logging.LogErrorf(err, "error getting sql DB")
		return err
	}

	return runMigrationLegacy(sqlDB, conn, legacyFn, migrationVersion, startFromZero)
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

func runMigrationVersioned(conn *gorm.DB, migFn VersionedMigrationFunc, migrationVersion uint, startFromZero bool) error {
	sqlDB, err := conn.DB()
	if err != nil {
		logging.LogErrorf(err, "error getting sql DB")
		return err
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

	return runMigrationVersions(ctx, conn, migration, migFn, migrationVersion, startFromZero)
}

func runMigrationVersions(
	ctx context.Context,
	conn *gorm.DB,
	migration *migrate.Migration,
	migFn VersionedMigrationFunc,
	migrationVersion uint,
	startFromZero bool,
) error {
	mpg, cleanup, err := migration.MigrateInstanceForVersionTracking()
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	currentVersion, dirty, needsRecordTarget, err := currentMigrationVersion(mpg, migrationVersion, startFromZero)
	if err != nil {
		return err
	}
	if dirty {
		return errors.Errorf("database migration is dirty at version %d", currentVersion)
	}

	// Legacy behavior: when the database has no version info and startFromZero is false,
	// run AutoMigrate once, then record the target version without running per-version hooks.
	if needsRecordTarget {
		if migFn != nil {
			if err := migFn(conn, migrationVersion); err != nil {
				logging.LogErrorf(err, "error running auto migration for version %d", migrationVersion)
				return err
			}
		}
		if err := migrate.SetVersion(mpg, migrationVersion); err != nil {
			logging.LogErrorf(err, "error setting migration version to %d", migrationVersion)
			return err
		}
		return nil
	}

	for version := currentVersion + 1; version <= migrationVersion; version++ {
		if err := applyMigrationVersion(ctx, conn, migration, migFn, mpg, version); err != nil {
			return err
		}
		logging.LogInfof("migration for version %d executed successfully", version)
	}

	return nil
}

func currentMigrationVersion(mpg migrate.VersionSetter, migrationVersion uint, startFromZero bool) (uint, bool, bool, error) {
	currentVersion, dirty, err := mpg.Version()
	if err == nil {
		return currentVersion, dirty, false, nil
	}
	if !stderrors.Is(err, migrate.ErrNilVersion) {
		return 0, false, false, err
	}
	if startFromZero {
		return 0, false, false, nil
	}
	// Caller should run a single AutoMigrate and then record the target version.
	return migrationVersion, false, true, nil
}

func applyMigrationVersion(
	ctx context.Context,
	conn *gorm.DB,
	migration *migrate.Migration,
	migFn VersionedMigrationFunc,
	mpg migrate.VersionSetter,
	version uint,
) error {
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
	if err := migrate.SetVersion(mpg, version); err != nil {
		logging.LogErrorf(err, "error setting migration version to %d", version)
		return err
	}

	return nil
}
