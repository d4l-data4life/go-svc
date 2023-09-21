package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"os"
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

// NewMigrationWithFdw returns a new migration instances for the given database connection with support for postgres_fdw via fdw.up.sql and fdw.down.sql scripts
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
func (m *Migration) MigrateDB(ctx context.Context, migrationVersion uint) error {
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

func fileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		// path exists
		return true, nil
	} else if os.IsNotExist(err) {
		// path does *not* exist
		return false, nil
	} else {
		// file may exists but os.Stat fails for other reasons (eg. permission, failing disk)
		return false, err
	}
}
