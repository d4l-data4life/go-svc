package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/d4l-data4life/go-svc/pkg/logging"
	"github.com/d4l-data4life/go-svc/pkg/migrate"
	"gorm.io/gorm"
)

func runMigrationVersioned(ctx context.Context, conn *gorm.DB, opts *ConnectionOptions) (err error) {
	sqlDB, err := conn.DB()
	if err != nil {
		return err
	}
	store, err := migrate.OpenVersionStore(ctx, sqlDB, migrationsTable)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, store.Close()) }()
	current, fresh, err := validateMigrationStart(ctx, store, opts)
	if err != nil {
		return err
	}
	if !fresh && current == opts.MigrationVersion {
		return nil
	}

	// All work uses the same physical session as the advisory lock. A dropped
	// connection cannot release the lock while callbacks continue on another one.
	scoped := conn.Session(&gorm.Session{NewDB: true, Context: ctx})
	scoped.Statement.ConnPool = store.Conn
	migration := migrate.NewVersionedMigration(store.Conn, migrationsSource, migrationsTable, logging.Logger())
	next := current + 1
	bootstrap := fresh && !opts.MigrationStartFromZero
	if bootstrap {
		next = opts.MigrationVersion
	}
	// Persist intent before any setup, hook or GORM mutation. Interrupted steps
	// remain dirty and are never automatically replayed against a partial schema.
	if err = store.SetVersion(ctx, next, true); err != nil {
		return err
	}
	if err = prepareVersionedMigration(ctx, migration); err != nil {
		return err
	}
	defer func() {
		cleanupErr := migration.ExecuteFdwDown(ctx)
		err = errors.Join(err, cleanupErr)
		if err == nil {
			err = store.SetVersion(ctx, opts.MigrationVersion, false)
		}
	}()
	return applyMigrationSteps(ctx, scoped, migration, store, opts, next, bootstrap)
}

func validateMigrationStart(ctx context.Context, store *migrate.VersionStore, opts *ConnectionOptions) (uint, bool, error) {
	current, dirty, readErr := store.Version(ctx)
	fresh := errors.Is(readErr, migrate.ErrNilVersion)
	if readErr != nil && !fresh {
		return 0, false, readErr
	}
	if dirty {
		return 0, false, fmt.Errorf("database migration is dirty at version %d; inspect and recover before restarting", current)
	}
	if !fresh && current > opts.MigrationVersion {
		return 0, false, fmt.Errorf("database version %d is newer than supported target %d", current, opts.MigrationVersion)
	}
	if !fresh && current < opts.MinimumMigrationVersion {
		return 0, false, fmt.Errorf(
			"unsupported database migration version %d; lowest supported starting version is %d",
			current,
			opts.MinimumMigrationVersion,
		)
	}
	if fresh {
		return 0, true, validateFreshMigration(ctx, store, opts)
	}

	return current, fresh, nil
}

func applyMigrationSteps(ctx context.Context, scoped *gorm.DB, migration *migrate.Migration,
	store *migrate.VersionStore, opts *ConnectionOptions, next uint, bootstrap bool) (err error) {
	for version := next; version <= opts.MigrationVersion; version++ {
		if version != next {
			if err = store.SetVersion(ctx, version, true); err != nil {
				return err
			}
		}
		if !bootstrap {
			if _, err = migration.ExecuteBeforeUp(ctx, version); err != nil {
				return err
			}
		}
		if err = opts.VersionedMigrationFunc(scoped, version); err != nil {
			return err
		}
		if !bootstrap {
			if _, err = migration.ExecuteAfterUp(ctx, version); err != nil {
				return err
			}
		}
		// Keep the final step dirty until shared cleanup has also succeeded.
		if version < opts.MigrationVersion {
			if err = store.SetVersion(ctx, version, false); err != nil {
				return err
			}
		}
	}
	return nil
}

func prepareVersionedMigration(ctx context.Context, migration *migrate.Migration) error {
	if err := migration.ExecuteSetup(ctx); err != nil {
		return err
	}
	return migration.ExecuteFdwUp(ctx)
}
func validateFreshMigration(ctx context.Context, store *migrate.VersionStore, opts *ConnectionOptions) error {
	empty, err := store.EmptySchema(ctx)
	if err != nil {
		return err
	}
	if !empty {
		return errors.New("cannot baseline a populated database without migration metadata")
	}
	if opts.MigrationStartFromZero && opts.MinimumMigrationVersion > 0 {
		return fmt.Errorf("cannot replay from zero; lowest supported starting version is %d", opts.MinimumMigrationVersion)
	}
	return nil
}
