package migrate

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4/database"
	"github.com/lib/pq"
)

// VersionStore owns one connection and the golang-migrate-compatible advisory lock.
// Hooks, GORM and version writes must all use Conn while this store is open.
// Closing the store releases only its connection, never the application's pool.
type VersionStore struct {
	Conn      *sql.Conn
	schema    string
	table     string
	qualified string
	lockID    string
}

// OpenVersionStore waits at most 30 seconds for the migration lock. The caller's
// context can cancel sooner. The version must be read only after this returns.
func OpenVersionStore(ctx context.Context, db *sql.DB, table string) (_ *VersionStore, err error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	store := &VersionStore{Conn: conn, table: table}
	defer func() {
		if err != nil {
			_ = store.Close()
		}
	}()
	var name string
	if err = conn.QueryRowContext(ctx, "SELECT current_database(), current_schema()").Scan(&name, &store.schema); err != nil {
		return nil, err
	}
	store.qualified = pq.QuoteIdentifier(store.schema) + "." + pq.QuoteIdentifier(table)
	id, err := database.GenerateAdvisoryLockId(name, store.schema, table)
	if err != nil {
		return nil, err
	}
	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	for {
		var acquired bool
		if err = conn.QueryRowContext(waitCtx, "SELECT pg_try_advisory_lock($1)", id).Scan(&acquired); err != nil {
			return nil, err
		}
		if acquired {
			store.lockID = id
			break
		}
		timer := time.NewTimer(100 * time.Millisecond)
		select {
		case <-waitCtx.Done():
			timer.Stop()
			return nil, fmt.Errorf("waiting for migration lock: %w", waitCtx.Err())
		case <-timer.C:
		}
	}
	query := "CREATE TABLE IF NOT EXISTS " + store.qualified + " (version bigint NOT NULL PRIMARY KEY, dirty boolean NOT NULL)"
	if _, err = conn.ExecContext(ctx, query); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *VersionStore) Version(ctx context.Context) (uint, bool, error) {
	var version int64
	var dirty bool
	err := s.Conn.QueryRowContext(ctx, "SELECT version, dirty FROM "+s.qualified).Scan(&version, &dirty)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, ErrNilVersion
	}
	if err != nil {
		return 0, false, err
	}
	if version < 0 {
		return 0, dirty, fmt.Errorf("invalid recorded migration version %d", version)
	}
	return uint(version), dirty, nil
}

// SetVersion atomically replaces the version record. It does not reacquire the lock.
func (s *VersionStore) SetVersion(ctx context.Context, version uint, dirty bool) error {
	tx, err := s.Conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	// Identifier is quoted when the store is opened; values remain parameters.
	//nolint:gosec
	if _, err = tx.ExecContext(ctx, "DELETE FROM "+s.qualified); err != nil {
		return err
	}
	//nolint:gosec // s.qualified contains only quoted identifiers.
	if _, err = tx.ExecContext(ctx, "INSERT INTO "+s.qualified+" (version,dirty) VALUES ($1,$2)", version, dirty); err != nil {
		return err
	}
	return tx.Commit()
}

// EmptySchema prevents silently baselining an existing, unversioned application.
func (s *VersionStore) EmptySchema(ctx context.Context) (bool, error) {
	var empty bool
	err := s.Conn.QueryRowContext(ctx, `SELECT NOT EXISTS (
 SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace
 WHERE n.nspname=$1 AND c.relname<>$2 AND c.relkind IN ('r','p','v','m','f','S')
 )`, s.schema, s.table).Scan(&empty)
	return empty, err
}

func (s *VersionStore) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// A failed hook may leave an explicit transaction open or aborted. Roll it
	// back before unlocking, and never return a possibly locked session to the pool.
	_, rollbackErr := s.Conn.ExecContext(ctx, "ROLLBACK")
	var unlockErr error
	if s.lockID != "" {
		var unlocked bool
		unlockErr = s.Conn.QueryRowContext(ctx, "SELECT pg_advisory_unlock($1)", s.lockID).Scan(&unlocked)
		if unlockErr == nil && !unlocked {
			unlockErr = errors.New("migration advisory lock was not held")
		}
	}
	if rollbackErr != nil || unlockErr != nil {
		_ = s.Conn.Raw(func(any) error { return driver.ErrBadConn })
	}
	return errors.Join(rollbackErr, unlockErr, s.Conn.Close())
}
