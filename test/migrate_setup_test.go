package test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/pkg/errors"

	"github.com/gesundheitscloud/go-svc/pkg/migrate"
)

func TestMigrateSetup(t *testing.T) {
	const (
		testSchema     = "test_setup"
		migrationTable = "migration"
		setupTestTable = "setuptable" // table created in the setup script
		expectedValue  = 1            // value inserted in the setup script, expected to be found in a successful test
	)

	// getTestValue attempts to retrieve and return the test value inserted from the setup script
	getTestValue := func(ctx context.Context, db *sql.DB) (int, error) {
		query := fmt.Sprintf("SELECT val FROM %s.%s LIMIT 1", testSchema, setupTestTable)

		var result int

		if err := db.QueryRowContext(ctx, query).Scan(&result); err != nil {
			return 0, err
		}

		return result, nil
	}

	cfg, err := parseEnv()
	if err != nil {
		t.Fatal(errors.Wrap(err, "could not parse the env"))
	}

	db, err := connectToDB(cfg)
	if err != nil {
		t.Fatal(errors.Wrap(err, "could not connect to the DB"))
	}

	ctx := context.Background()

	for _, tc := range []struct {
		name          string
		scriptsFolder string
		version       uint
	}{
		{
			name:          "setup script is run",
			scriptsFolder: "sql/test-setup",
			version:       1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				_ = cleanSchema(ctx, db, testSchema)
				_ = cleanTable(ctx, db, migrationTable)
			}()

			m := migrate.NewMigration(db, tc.scriptsFolder, migrationTable, &testLog{})
			if err != nil {
				t.Fatal(errors.Wrap(err, "could not create a migration instance"))
			}

			if err = m.MigrateDB(ctx, tc.version); err != nil {
				t.Fatal(errors.Wrap(err, "could not run the migration"))
			}

			if val, err := getTestValue(ctx, db); err != nil || val != expectedValue {
				t.Errorf("expected to find value %d in table %s", expectedValue, setupTestTable)
			}
		})
	}
}
