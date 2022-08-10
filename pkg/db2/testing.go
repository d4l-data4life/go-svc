package db2

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
)

// InitializeTest calls InitializeTestSqlite3 (duplicate func for compatibility purposes)
func InitializeTest(migFn MigrationFunc) {
	InitializeTestSqlite3(migFn)
}

// InitializeTestSqlite3 connects to an inmemory sqlite for testing
func InitializeTestSqlite3(migFn MigrationFunc) {
	conn, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"))
	if err != nil {
		fmt.Printf("Test DB connection error: %s\n", err.Error())
	}

	db = conn
	// 'foreign_keys = off' is default setting in SQLite.
	db.Exec("PRAGMA foreign_keys = ON")
	if migFn != nil {
		if err = runMigration(conn, migFn, 0); err != nil {
			logging.LogErrorf(err, "test DB migration error")
		}
	}
}

func ConnectString(opts *ConnectionOptions) string {
	core := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s", opts.Host, opts.Port, opts.DatabaseName, opts.User, opts.Password)
	addons := ""
	if opts.SSLMode != "" {
		addons = fmt.Sprintf("%s sslmode=%s", addons, opts.SSLMode)
	}
	if (opts.SSLMode == "verify-ca" || opts.SSLMode == "verify-full") && opts.SSLRootCertPath != "" {
		addons = fmt.Sprintf("%s sslrootcert=%s", addons, opts.SSLRootCertPath)
	}
	return core + addons
}

// InitializeTestPostgres connects to a postgess db
func InitializeTestPostgres(opts *ConnectionOptions) {
	connectString := ConnectString(opts)
	logging.LogDebugf("Attempting to connect to DB using: %s", connectString)

	Close()
	var conn *gorm.DB

	if opts.DriverFunc == nil {
		opts.DriverFunc = DefaultPostgresDriver
	}
	conn, err := opts.DriverFunc(connectString, opts)

	db = conn
	if err != nil {
		logging.LogErrorf(err, "error connecting to testing postgres")
		db = nil
	}
	if opts.MigrationFunc != nil {
		if err = runMigration(conn, opts.MigrationFunc, opts.MigrationVersion); err != nil {
			logging.LogErrorf(err, "test DB migration error")
		}
	}
}
