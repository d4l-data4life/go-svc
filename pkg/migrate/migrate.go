package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	// import the file driver for reading the migration scripts from files
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pkg/errors"
)

const (
	setupScriptName   = "setup.sql"
	fdwUpScriptName   = "fdw.up.sql"
	fdwDownScriptName = "fdw.down.sql"
	beforeUpSuffix    = ".before.up.sql"
	beforeDownSuffix  = ".before.down.sql"
)

// Migration is the struct that holds the information needed for migrating a database.
type Migration struct {
	db              *sql.DB
	migrationTable  string
	foreignDatabase *ForeignDatabase
	sourceFolder    string
	log             logger
}

type ForeignDatabase struct {
	LocalUser string
	DBName    string
	Hostname  string
	Port      uint
	User      string
	Password  string
}

// NewMigration returns a new migration instances for the given database connection
func NewMigration(db *sql.DB, sourceFolder, migrationTable string, log logger) *Migration {
	return &Migration{
		db:             db,
		migrationTable: migrationTable,
		sourceFolder:   sourceFolder,
		log:            log,
	}
}

// NewMigrationWithFdw returns a new migration instances for the given database connection
// with support forpostgres_fdwvia fdw.up.sql and fdw.down.sql scripts
func NewMigrationWithFdw(db *sql.DB, sourceFolder, migrationTable string, foreignDB *ForeignDatabase, log logger) *Migration {
	return &Migration{
		db:              db,
		migrationTable:  migrationTable,
		foreignDatabase: foreignDB,
		sourceFolder:    sourceFolder,
		log:             log,
	}
}

// MigrateDB executes a DB migration.
//
// 1. It first executes the setup script (if such a script exists).
//
// 2. Execute the fdw.up.sql script (if exists) by templating via ForeignDatabase (e.g. for postgres_fdw)
//
// 3. Then it delegates the run of the numbered migration steps to golang-migrate.
//
// 4. Execute the fdw.down.sql script (if exists) by templating via ForeignDatabase (e.g. for postgres_fdw)
//
// nolint: gocyclo
func (m *Migration) MigrateDB(ctx context.Context, migrationVersion uint, startFromZero bool) error {
	if err := m.execute(ctx, setupScriptName, nil); err != nil { // execute setup
		return errors.Wrap(err, "could not run the setup script")
	}

	if err := m.execute(ctx, fdwUpScriptName, m.foreignDatabase); err != nil { // execute fdw.up
		return errors.Wrap(err, "could not run the fdw.up script")
	}

	driver, err := postgres.WithInstance(m.db, &postgres.Config{
		MigrationsTable: m.migrationTable,
	})
	if err != nil {
		return errors.Wrap(err, "error creating database driver")
	}

	mpg, err := migrate.NewWithDatabaseInstance(
		"file://"+m.sourceFolder,
		"postgres",
		driver,
	)
	if err != nil {
		return errors.Wrap(err, "error creating migrate instance")
	}

	_, _, err = mpg.Version()
	if err == migrate.ErrNilVersion && !startFromZero {
		// no migration information in the database, so it's a fresh database
		// and the data model is already the latest one set up Gorm automigrations
		// nolint: gosec
		err = mpg.Force(int(migrationVersion))
		if err != nil {
			return errors.Wrap(err, "error setting migration version")
		}
	}

	err = mpg.Migrate(migrationVersion)

	switch err {
	case nil:
		_ = m.log.InfoGeneric(ctx, fmt.Sprintf("migration to v%d succeeded", migrationVersion))
	case migrate.ErrNoChange:
		_ = m.log.InfoGeneric(ctx, fmt.Sprintf("migration to v%d skipped: no changes", migrationVersion))
	case database.ErrLocked:
		_ = m.log.InfoGeneric(ctx, fmt.Sprintf("migration to v%d skipped: database locked by another instance", migrationVersion))
	default:
		return errors.Wrap(err, fmt.Sprintf("error migrating database to v%d", migrationVersion))
	}

	if err := m.execute(ctx, fdwDownScriptName, m.foreignDatabase); err != nil { // execute fdw.down
		return errors.Wrap(err, "could not run the fdw.down script")
	}

	return nil
}

func (m *Migration) parseFile(ctx context.Context, filename string, templateData interface{}) (string, error) {
	path := m.sourceFolder + "/" + filename

	exists, err := fileExists(path)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("could not access the file on path %s", path))
	}

	if !exists {
		_ = m.log.InfoGeneric(ctx, fmt.Sprintf("sql file %q does not exist - skipped execution", path))
		return "", nil
	}

	c, err := os.ReadFile(path)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("could not open the file on path %s", path))
	}

	sql := string(c)

	if templateData != nil {
		tmpl, err := template.New("sqlTemplate").Parse(sql)
		if err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("unable to parse template on path %s", path))
		}
		parsed := &strings.Builder{}
		if err := tmpl.Execute(parsed, templateData); err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("unable to execute template on path %s", path))
		}
		sql = parsed.String()
	}

	return sql, nil
}

func (m *Migration) execute(ctx context.Context, filename string, templateData interface{}) error {
	sql, err := m.parseFile(ctx, filename, templateData)
	if err != nil {
		return fmt.Errorf("could not parse the %q script: %w", filename, err)
	}
	if sql == "" {
		_ = m.log.InfoGeneric(ctx, fmt.Sprintf("nothing to execute for script %q", filename))
		return nil
	}
	_, err = m.db.ExecContext(ctx, sql)
	if err == nil {
		_ = m.log.InfoGeneric(ctx, fmt.Sprintf("successfully executed script %q", filename))
	}
	return err
}

// ExecuteTargetBeforeUp runs a target-version before migration if present.
// The script is expected to be idempotent because it is not tracked.
func (m *Migration) ExecuteTargetBeforeUp(ctx context.Context, migrationVersion uint) (bool, error) {
	filename, err := findBeforeUpFile(m.sourceFolder, migrationVersion)
	if err != nil {
		return false, errors.Wrap(err, "could not scan for before migration")
	}
	if filename == "" {
		_ = m.log.InfoGeneric(ctx, fmt.Sprintf("no before migration found for version %d - skipped", migrationVersion))
		return false, nil
	}
	if err := m.execute(ctx, filename, nil); err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("could not run before migration %q", filename))
	}
	return true, nil
}

// CreateAfterSourceFolder returns a temp folder containing only non-before migrations.
func CreateAfterSourceFolder(sourceFolder string) (string, func(), error) {
	entries, err := os.ReadDir(sourceFolder)
	if err != nil {
		return "", nil, errors.Wrap(err, "could not read migrations folder")
	}
	tempDir, err := os.MkdirTemp("", "migrate-after-*")
	if err != nil {
		return "", nil, errors.Wrap(err, "could not create temp folder")
	}
	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if isBeforeMigrationFile(name) {
			continue
		}
		if err := copyFile(filepath.Join(sourceFolder, name), filepath.Join(tempDir, name)); err != nil {
			cleanup()
			return "", nil, err
		}
	}
	return tempDir, cleanup, nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		// path exists
		return true, nil
	} else if os.IsNotExist(err) {
		// path does *not* exist
		return false, nil
	}
	// file may exists but os.Stat fails for other reasons (eg. permission, failing disk)
	return false, err
}

func isBeforeMigrationFile(filename string) bool {
	return strings.HasSuffix(filename, beforeUpSuffix) || strings.HasSuffix(filename, beforeDownSuffix)
}

func findBeforeUpFile(sourceFolder string, migrationVersion uint) (string, error) {
	entries, err := os.ReadDir(sourceFolder)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, beforeUpSuffix) {
			continue
		}
		version, ok := parseMigrationVersion(name)
		if !ok {
			continue
		}
		if version == migrationVersion {
			return name, nil
		}
	}
	return "", nil
}

func parseMigrationVersion(filename string) (uint, bool) {
	base := filepath.Base(filename)
	sep := strings.Index(base, "_")
	if sep <= 0 {
		return 0, false
	}
	versionStr := base[:sep]
	if len(versionStr) == 0 {
		return 0, false
	}
	parsed, err := strconv.ParseUint(versionStr, 10, 32)
	if err != nil {
		return 0, false
	}
	return uint(parsed), true
}

func copyFile(src, dest string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not read %q", src))
	}
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not write %q", dest))
	}
	return nil
}
