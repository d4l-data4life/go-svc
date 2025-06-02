package test

import (
	"context"
	"testing"

	"github.com/pkg/errors"

	"github.com/gesundheitscloud/go-svc/pkg/migrate"
)

func TestMigrateDown(t *testing.T) {
	const (
		testSchema     = "test"
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
		name           string
		scriptsFolder  string
		initialVersion uint
		targetVersion  uint
	}{
		{
			name:           "should be able to downgrade partially",
			scriptsFolder:  "sql/down",
			initialVersion: 3,
			targetVersion:  2,
		},
		{
			name:           "should be able to downgrade completely",
			scriptsFolder:  "sql/down",
			initialVersion: 3,
			targetVersion:  1,
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

			// first migrate to the initial version
			if err = m.MigrateDB(ctx, tc.initialVersion); err != nil {
				t.Fatal(errors.Wrap(err, "could not reach the initial version"))
			}

			if err = m.MigrateDB(ctx, tc.targetVersion); err != nil {
				t.Error(errors.Wrap(err, "could not run the migration"))
			}

			// no setup script is setting the search path, so expect the migration table to be found in the public schema
			sch, err := getSchemaForTable(ctx, db, migrationTable)
			if err != nil {
				t.Fatal(errors.Wrap(err, "could not run the migration table check"))
			}
			if sch != "public" {
				t.Errorf("expected to find migration table %s in schema %s. Found it in %s", migrationTable, "public", sch)
			}
		})
	}
}
