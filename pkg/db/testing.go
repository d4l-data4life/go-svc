package db

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/d4l-data4life/go-svc/pkg/logging"
)

func TestConnectString(opts *ConnectionOptions) string {
	core := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s", opts.Host, opts.Port, opts.DatabaseName, opts.User, opts.Password)
	addons := ""
	if opts.SSLMode != "" {
		addons = fmt.Sprintf("%s sslmode=%s", addons, opts.SSLMode)
	}
	if (opts.SSLMode == "verify-ca" || opts.SSLMode == SSLVerifyFull) && opts.SSLRootCertPath != "" {
		addons = fmt.Sprintf("%s sslrootcert=%s", addons, opts.SSLRootCertPath)
	}
	return core + addons
}

// InitializeTestPostgres connects to a postgess db
func InitializeTestPostgres(opts *ConnectionOptions) {
	connectString := TestConnectString(opts)
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
	if opts.VersionedMigrationFunc != nil {
		if err = runMigrationWithOptions(context.Background(), conn, opts); err != nil {
			logging.LogErrorf(err, "test DB migration error")
			Close()
			db = nil
		}
	} else if opts.MigrationFunc != nil {
		// Preserve the legacy helper: replay only when a callback is configured,
		// and retain the connected handle when migration fails.
		testOpts := *opts
		testOpts.MigrationStartFromZero = true
		if err = runMigrationWithOptions(context.Background(), conn, &testOpts); err != nil {
			logging.LogErrorf(err, "test DB migration error")
		}
	}
}
