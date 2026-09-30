package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4l-data4life/go-svc/pkg/migrate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMigrationConfiguration(t *testing.T) {
	opts := NewConnection(WithMigrationFunc(func(*gorm.DB) error { return nil }),
		WithVersionedMigrationFunc(func(*gorm.DB, uint) error { return nil }))
	require.ErrorContains(t, runMigrationWithOptions(context.Background(), &gorm.DB{}, opts), "both MigrationFunc")
	require.ErrorIs(t, runMigrationWithOptions(context.Background(), nil, opts), ErrDBConnection)
	called := false
	opts.MigrationFunc = nil
	opts.VersionedMigrationFunc = func(_ *gorm.DB, version uint) error { called = true; assert.Zero(t, version); return nil }
	require.NoError(t, runMigrationWithOptions(context.Background(), &gorm.DB{Config: &gorm.Config{}, Statement: &gorm.Statement{}}, opts))
	require.True(t, called)
}

func TestMigrationMinimumExceedsTarget(t *testing.T) {
	opts := NewConnection(WithMigrationVersion(8), WithMinimumMigrationVersion(9),
		WithVersionedMigrationFunc(func(*gorm.DB, uint) error {
			t.Fatal("invalid options must not reach the callback")
			return nil
		}))
	// A handle without a connection also proves validation precedes DB access.
	require.ErrorContains(t, runMigrationWithOptions(context.Background(), &gorm.DB{}, opts),
		"minimum migration version 9 exceeds target 8")
}

// Explicit local-only opt-in: the test never reads an arbitrary database URL.
func localMigrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	port := os.Getenv("GO_SVC_MIGRATION_TEST_PORT")
	if port == "" {
		t.Skip("set GO_SVC_MIGRATION_TEST_PORT for a disposable localhost database")
	}
	opts := NewConnection(
		WithHost("127.0.0.1"),
		WithPort(port),
		WithDatabaseName("migration_test"),
		WithUser("postgres"),
		WithPassword("postgres"),
		WithSSLMode("disable"),
	)
	conn, err := DefaultPostgresDriver(ConnectString(opts), opts)
	require.NoError(t, err)
	pool, err := conn.DB()
	require.NoError(t, err)
	pool.SetMaxOpenConns(1) // Detect accidental use of a second connection while holding the lock.
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	return conn
}
func cleanMigrationDB(t *testing.T, conn *gorm.DB) {
	t.Helper()
	require.NoError(t, conn.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public").Error)
	t.Chdir(t.TempDir())
	require.NoError(t, os.Mkdir("sql", 0700))
}
func hook(t *testing.T, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join("sql", name), []byte(content), 0600))
}
func migrationState(t *testing.T, conn *gorm.DB, version uint, dirty bool) {
	t.Helper()
	var got struct {
		Version uint
		Dirty   bool
	}
	require.NoError(t, conn.Raw("SELECT version,dirty FROM migrations").Scan(&got).Error)
	require.Equal(t, version, got.Version)
	require.Equal(t, dirty, got.Dirty)
}
func TestVersionedMigrationPostgres(t *testing.T) {
	conn := localMigrationDB(t)
	t.Run("bootstrap_and_restart", func(t *testing.T) {
		cleanMigrationDB(t, conn)
		calls := 0
		opts := NewConnection(
			WithMigrationVersion(8),
			WithMinimumMigrationVersion(7),
			WithVersionedMigrationFunc(func(tx *gorm.DB, v uint) error {
				calls++
				require.Equal(t, uint(8), v)
				return tx.Exec("CREATE TABLE application_data (id integer PRIMARY KEY)").Error
			}),
		)
		hook(t, "8_skip.before.sql", "SELECT no_such_bootstrap_function();")
		require.NoError(t, runMigrationWithOptions(context.Background(), conn, opts))
		migrationState(t, conn, 8, false)
		require.NoError(t, runMigrationWithOptions(context.Background(), conn, opts))
		require.Equal(t, 1, calls)
		pool, _ := conn.DB()
		require.Zero(t, pool.Stats().InUse)
		require.NoError(t, pool.Ping())
	})
	t.Run("ordered_steps", func(t *testing.T) {
		cleanMigrationDB(t, conn)
		hook(t, "1_init.before.sql", "CREATE TABLE steps (seq serial, step text); INSERT INTO steps(step) VALUES ('before1');")
		hook(t, "1_init.after.sql", "INSERT INTO steps(step) VALUES ('after1');")
		hook(t, "2_more.before.up.sql", "INSERT INTO steps(step) VALUES ('before2');")
		hook(t, "2_more.after.up.sql", "INSERT INTO steps(step) VALUES ('after2');")
		opts := NewConnection(
			WithMigrationVersion(2),
			WithMigrationStartFromZero(true),
			WithVersionedMigrationFunc(func(tx *gorm.DB, v uint) error {
				return tx.Exec("INSERT INTO steps(step) VALUES (?)", fmt.Sprintf("auto%d", v)).Error
			}),
		)
		require.NoError(t, runMigrationWithOptions(context.Background(), conn, opts))
		var steps []string
		require.NoError(t, conn.Raw("SELECT step FROM steps ORDER BY seq").Scan(&steps).Error)
		require.Equal(t, []string{"before1", "auto1", "after1", "before2", "auto2", "after2"}, steps)
		migrationState(t, conn, 2, false)
	})
	for _, stage := range []string{"setup", "before", "auto", "after", "record", "cleanup", "bootstrap"} {
		t.Run("failure_"+stage, func(t *testing.T) {
			cleanMigrationDB(t, conn)
			calls := 0
			if stage == "setup" {
				hook(t, "setup.sql", "SELECT missing_function();")
			}
			if stage == "before" {
				hook(t, "1_fail.before.sql", "BEGIN; CREATE TABLE partial (id int); SELECT missing_function(); COMMIT;")
			}
			if stage == "after" {
				hook(t, "1_fail.after.sql", "SELECT missing_function();")
			}
			if stage == "cleanup" {
				hook(t, "fdw.down.sql", "SELECT missing_function();")
			}
			if stage == "record" {
				hook(
					t,
					"1_fail.after.sql",
					`CREATE FUNCTION reject_record() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'record failure'; END $$;
CREATE TRIGGER reject_record BEFORE DELETE ON migrations FOR EACH ROW EXECUTE FUNCTION reject_record();`,
				)
			}
			opts := NewConnection(
				WithMigrationVersion(1),
				WithMigrationStartFromZero(stage != "bootstrap"),
				WithVersionedMigrationFunc(func(tx *gorm.DB, _ uint) error {
					calls++
					if stage == "auto" || stage == "bootstrap" {
						return errors.New("injected callback failure")
					}
					return tx.Exec("CREATE TABLE completed_callback (id int)").Error
				}),
			)
			require.Error(t, runMigrationWithOptions(context.Background(), conn, opts))
			migrationState(t, conn, 1, true)
			previous := calls
			require.ErrorContains(t, runMigrationWithOptions(context.Background(), conn, opts), "dirty")
			require.Equal(t, previous, calls, "dirty steps must not replay")
		})
	}
	for _, tc := range []struct {
		name    string
		version uint
		dirty   bool
		message string
	}{
		{"older", 6, false, "lowest supported"}, {"newer", 9, false, "newer than"}, {"dirty", 7, true, "dirty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cleanMigrationDB(t, conn)
			require.NoError(t, conn.Exec("CREATE TABLE migrations (version bigint PRIMARY KEY,dirty boolean NOT NULL)").Error)
			require.NoError(t, conn.Exec("INSERT INTO migrations VALUES (?,?)", tc.version, tc.dirty).Error)
			opts := NewConnection(
				WithMigrationVersion(8),
				WithMinimumMigrationVersion(7),
				WithVersionedMigrationFunc(func(*gorm.DB, uint) error { t.Fatal("callback must not run"); return nil }),
			)
			require.ErrorContains(t, runMigrationWithOptions(context.Background(), conn, opts), tc.message)
			migrationState(t, conn, tc.version, tc.dirty)
		})
	}
	t.Run("unversioned_populated", func(t *testing.T) {
		cleanMigrationDB(t, conn)
		require.NoError(t, conn.Exec("CREATE TABLE existing_data (id int); INSERT INTO existing_data VALUES (42)").Error)
		opts := NewConnection(
			WithMigrationVersion(8),
			WithVersionedMigrationFunc(func(*gorm.DB, uint) error { t.Fatal("callback must not run"); return nil }),
		)
		require.ErrorContains(t, runMigrationWithOptions(context.Background(), conn, opts), "populated")
		var id int
		require.NoError(t, conn.Raw("SELECT id FROM existing_data").Scan(&id).Error)
		require.Equal(t, 42, id)
	})
	t.Run("zero_replay_below_minimum", func(t *testing.T) {
		cleanMigrationDB(t, conn)
		opts := NewConnection(
			WithMigrationVersion(8),
			WithMinimumMigrationVersion(7),
			WithMigrationStartFromZero(true),
			WithVersionedMigrationFunc(func(*gorm.DB, uint) error { t.Fatal("callback must not run"); return nil }),
		)
		require.ErrorContains(t, runMigrationWithOptions(context.Background(), conn, opts), "cannot replay from zero")
	})
	t.Run("concurrent_startups_wait_and_reread", func(t *testing.T) {
		cleanMigrationDB(t, conn)
		second := localMigrationDB(t)
		entered := make(chan struct{})
		release := make(chan struct{})
		defer func() {
			select {
			case <-release:
			default:
				close(release)
			}
		}()
		var calls atomic.Int32
		opts := NewConnection(WithMigrationVersion(8), WithVersionedMigrationFunc(func(tx *gorm.DB, _ uint) error {
			if calls.Add(1) == 1 {
				close(entered)
				<-release
			}
			return tx.Exec("CREATE TABLE one_execution (id int)").Error
		}))
		firstDone := make(chan error, 1)
		secondDone := make(chan error, 1)
		go func() { firstDone <- runMigrationWithOptions(context.Background(), conn, opts) }()
		select {
		case <-entered:
		case <-time.After(5 * time.Second):
			t.Fatal("first runner did not enter")
		}
		go func() { secondDone <- runMigrationWithOptions(context.Background(), second, opts) }()
		select {
		case err := <-secondDone:
			t.Fatalf("waiter returned before migration finished: %v", err)
		case <-time.After(250 * time.Millisecond):
		}
		close(release)
		require.NoError(t, <-firstDone)
		require.NoError(t, <-secondDone)
		require.Equal(t, int32(1), calls.Load())
		migrationState(t, conn, 8, false)
	})
	t.Run("cancel_waiter", func(t *testing.T) {
		cleanMigrationDB(t, conn)
		second := localMigrationDB(t)
		entered := make(chan struct{})
		release := make(chan struct{})
		defer func() {
			select {
			case <-release:
			default:
				close(release)
			}
		}()
		opts := NewConnection(
			WithMigrationVersion(8),
			WithVersionedMigrationFunc(func(*gorm.DB, uint) error { close(entered); <-release; return nil }),
		)
		done := make(chan error, 1)
		go func() { done <- runMigrationWithOptions(context.Background(), conn, opts) }()
		select {
		case <-entered:
		case <-time.After(5 * time.Second):
			t.Fatal("first runner did not enter")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()
		require.ErrorIs(t, runMigrationWithOptions(ctx, second, opts), context.DeadlineExceeded)
		close(release)
		require.NoError(t, <-done)
	})
}

// An unfinished transaction must not be committed by a later version write.
func TestVersionedMigrationRejectsOpenTransactions(t *testing.T) {
	conn := localMigrationDB(t)
	for _, phase := range []string{"before", "after", "callback", "setup", "fdw.up", "fdw.down"} {
		t.Run(phase, func(t *testing.T) {
			cleanMigrationDB(t, conn)
			script := "BEGIN; CREATE TABLE uncommitted_hook (id int); INSERT INTO uncommitted_hook VALUES (42);"
			switch phase {
			case "before", "after":
				hook(t, "1_open."+phase+".sql", script)
			case "setup", "fdw.up", "fdw.down":
				hook(t, phase+".sql", script)
			}
			opts := NewConnection(WithMigrationVersion(1), WithMigrationStartFromZero(true),
				WithVersionedMigrationFunc(func(tx *gorm.DB, _ uint) error {
					if phase == "callback" {
						return tx.Exec(script).Error
					}
					return nil
				}))
			require.ErrorContains(t, runMigrationWithOptions(context.Background(), conn, opts), "non-idle transaction")
			observer := localMigrationDB(t)
			migrationState(t, observer, 1, true)
			var absent bool
			require.NoError(t, observer.Raw("SELECT to_regclass('public.uncommitted_hook') IS NULL").Scan(&absent).Error)
			require.True(t, absent, "uncommitted hook/callback changes must be rolled back")
			require.ErrorContains(t, runMigrationWithOptions(context.Background(), conn, opts), "dirty")
		})
	}
}

func TestLegacyTestHelperCompatibility(t *testing.T) {
	for _, withCallback := range []bool{false, true} {
		t.Run(fmt.Sprintf("callback_%t", withCallback), func(t *testing.T) {
			conn := localMigrationDB(t)
			cleanMigrationDB(t, conn)
			opts := NewConnection(WithMigrationVersion(1),
				WithDriverFunc(func(string, *ConnectionOptions) (*gorm.DB, error) { return conn, nil }))
			called := false
			if withCallback {
				opts.MigrationFunc = func(*gorm.DB) error {
					called = true
					return errors.New("legacy helper failure")
				}
			}
			InitializeTestPostgres(opts)
			t.Cleanup(Close)
			require.Same(t, conn, Get(), "legacy helper retains handle even after migration error")
			require.Equal(t, withCallback, called)
			var count int
			query := "SELECT count(*) FROM pg_tables WHERE schemaname='public' AND tablename='migrations'"
			require.NoError(t, conn.Raw(query).Scan(&count).Error)
			require.Zero(t, count, "version alone must not execute legacy SQL migrations")
		})
	}
}

func TestVersionedMigrationRejectsUnsupportedDriver(t *testing.T) {
	conn := localMigrationDB(t)
	cleanMigrationDB(t, conn)
	dsn := fmt.Sprintf("host=127.0.0.1 port=%s dbname=migration_test user=postgres password=postgres sslmode=disable",
		os.Getenv("GO_SVC_MIGRATION_TEST_PORT"))
	pool, err := sql.Open("postgres", dsn) // lib/pq, deliberately not pgx.
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	store, err := migrate.OpenVersionStore(context.Background(), pool, "migrations")
	require.ErrorContains(t, err, "unsupported driver")
	require.Nil(t, store)
	var absent bool
	require.NoError(t, conn.Raw("SELECT to_regclass('public.migrations') IS NULL").Scan(&absent).Error)
	require.True(t, absent, "unsupported driver must fail before metadata creation")
}
