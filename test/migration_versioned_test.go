package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/d4l-data4life/go-svc/pkg/db"
	"github.com/d4l-data4life/go-svc/pkg/migrate"
	"gorm.io/gorm"
)

func TestVersionedMigrationFlow(t *testing.T) {
	cfg, err := parseEnv()
	if err != nil {
		t.Fatal(err)
	}

	sqlDB, err := connectToDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
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
	if err := os.MkdirAll(sqlDir, 0o755); err != nil {
		t.Fatal(err)
	}

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
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
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
	if conn == nil {
		t.Fatal("db handle is nil")
	}

	type row struct {
		Seq  int
		Step string
	}
	rows := []row{}
	if err := conn.Raw("SELECT seq, step FROM migration_steps ORDER BY seq").Scan(&rows).Error; err != nil {
		t.Fatalf("query steps: %v", err)
	}

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
	if len(rows) != len(want) {
		t.Fatalf("got %d steps, want %d", len(rows), len(want))
	}
	for i, w := range want {
		if rows[i].Step != w {
			t.Fatalf("step %d: got %q, want %q", i, rows[i].Step, w)
		}
	}

	migration := migrate.NewMigration(sqlDB, sqlDir, "migrations", &testLog{})
	mpg, cleanup, err := migration.MigrateInstanceForVersionTracking()
	if err != nil {
		t.Fatalf("current version: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}
	version, dirty, err := mpg.Version()
	if err != nil {
		t.Fatalf("current version: %v", err)
	}
	if dirty {
		t.Fatalf("expected clean migrations table")
	}
	if version != 4 {
		t.Fatalf("expected version 4, got %d", version)
	}
}

func writeSQL(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
