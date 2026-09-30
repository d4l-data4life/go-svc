package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/d4l-data4life/go-svc/pkg/db"
	"github.com/d4l-data4life/go-svc/pkg/migrate"
	"gorm.io/gorm"
)

func TestVersionedMigrationFlow(t *testing.T) {
	cfg, err := parseEnv()
	require.NoError(t, err)

	sqlDB, err := connectToDB(cfg)
	require.NoError(t, err)
	defer sqlDB.Close()

	ctx := context.Background()
	_ = cleanTable(ctx, sqlDB, "migration_steps")
	_ = cleanTable(ctx, sqlDB, "migrations")
	defer func() {
		_ = cleanTable(ctx, sqlDB, "migration_steps")
		_ = cleanTable(ctx, sqlDB, "migrations")
	}()

	tmpDir := t.TempDir()
	sqlDir := filepath.Join(tmpDir, "sql")
	require.NoError(t, os.MkdirAll(sqlDir, 0o755))

	writeSQL(t, sqlDir, "001_init.before.sql", `
CREATE TABLE IF NOT EXISTS migration_steps (
  seq SERIAL PRIMARY KEY,
  step TEXT NOT NULL
);
INSERT INTO migration_steps (step) VALUES ('before-1');
`)
	writeSQL(t, sqlDir, "001_init.after.sql", `
INSERT INTO migration_steps (step) VALUES ('after-1');
`)
	writeSQL(t, sqlDir, "002_add.before.sql", `
INSERT INTO migration_steps (step) VALUES ('before-2');
`)
	writeSQL(t, sqlDir, "002_add.after.sql", `
INSERT INTO migration_steps (step) VALUES ('after-2');
`)
	writeSQL(t, sqlDir, "003_more.before.up.sql", `
INSERT INTO migration_steps (step) VALUES ('before-3');
`)
	writeSQL(t, sqlDir, "003_more.after.up.sql", `
INSERT INTO migration_steps (step) VALUES ('after-3');
`)

	cwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() {
		_ = os.Chdir(cwd)
	}()

	migFn := func(conn *gorm.DB, version uint) error {
		return conn.Exec(fmt.Sprintf("INSERT INTO migration_steps (step) VALUES ('auto-%d')", version)).Error
	}

	opts := db.NewConnection(
		db.WithHost(cfg.PGHost),
		db.WithPort(strconv.FormatUint(uint64(cfg.PGPort), 10)),
		db.WithDatabaseName(cfg.PGName),
		db.WithUser(cfg.PGUser),
		db.WithPassword(cfg.PGPassword),
		db.WithSSLMode("disable"),
		db.WithMigrationStartFromZero(true),
		db.WithMigrationVersion(4),
		db.WithVersionedMigrationFunc(migFn),
	)

	db.InitializeTestPostgres(opts)
	conn := db.Get()
	require.NotNil(t, conn, "db handle is nil")

	type row struct {
		Seq  int
		Step string
	}
	rows := []row{}
	require.NoError(t, conn.Raw("SELECT seq, step FROM migration_steps ORDER BY seq").Scan(&rows).Error)

	want := []string{
		"before-1",
		"auto-1",
		"after-1",
		"before-2",
		"auto-2",
		"after-2",
		"before-3",
		"auto-3",
		"after-3",
		"auto-4",
	}
	require.Len(t, rows, len(want))
	for i, w := range want {
		require.Equal(t, w, rows[i].Step, "step %d", i)
	}

	migrationPool, err := conn.DB()
	require.NoError(t, err)
	store, err := migrate.OpenVersionStore(ctx, migrationPool, "migrations")
	require.NoError(t, err)
	defer store.Close()
	version, dirty, err := store.Version(ctx)
	require.NoError(t, err)
	require.False(t, dirty)
	require.Equal(t, uint(4), version)
}

func writeSQL(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
}
