package test

import (
	"context"
	"testing"

	"github.com/pkg/errors"

	"github.com/d4l-data4life/go-svc/pkg/migrate"
)

func TestMigrateUp(t *testing.T) {
	const (
		migrationTable = "migration"
	)

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
		name             string
		scriptsFolder    string
		version          uint
		searchPathSchema string // schema set in the search path by the setup script (if any); where the migration table is expected to be found
		createdSchema    string // schema created by the scripts, which needs to be cleaned up
	}{
		{
			name:             "works with a setup script that creates a schema",
			scriptsFolder:    "sql/up-with-setup",
			version:          1,
			searchPathSchema: "test_up",
			createdSchema:    "test_up",
		},
		{
			name:             "works without a setup script",
			scriptsFolder:    "sql/up-without-setup",
			version:          2,
			searchPathSchema: "",
			createdSchema:    "test_no_setup",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if tc.createdSchema != "" {
					_ = cleanSchema(ctx, db, tc.createdSchema)
				}
				_ = cleanTable(ctx, db, migrationTable)
			}()

			m := migrate.NewMigration(db, tc.scriptsFolder, migrationTable, &testLog{})
			if err != nil {
				t.Fatal(errors.Wrap(err, "could not create a migration instance"))
			}

			if err = m.MigrateDB(ctx, tc.version); err != nil {
				t.Error(errors.Wrap(err, "could not run the migration"))
			}

			migrationTableSchema, err := getSchemaForTable(ctx, db, migrationTable)
			if err != nil {
				t.Fatal(errors.Wrap(err, "could not run the migration table check"))
			}

			var expectedMigrationSchema string
			if tc.searchPathSchema == "" {
				expectedMigrationSchema = "public"
			} else {
				expectedMigrationSchema = tc.searchPathSchema
			}

			if migrationTableSchema != expectedMigrationSchema {
				t.Errorf(
					"expected to find migration table %s in schema %s. Found it in schema %s",
					migrationTable,
					expectedMigrationSchema,
					migrationTableSchema,
				)
			}
		})
	}
}
