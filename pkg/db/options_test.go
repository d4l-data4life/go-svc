package db

import (
	"testing"

	"gorm.io/gorm"
)

func TestWithMigrationFuncDoesNotSetVersionedMigrationFunc(t *testing.T) {
	fn := func(_ *gorm.DB) error { return nil }

	opts := NewConnection(
		WithMigrationVersion(2),
		WithMigrationFunc(fn),
	)

	if opts.MigrationFunc == nil {
		t.Fatalf("expected MigrationFunc to be set")
	}
	if opts.VersionedMigrationFunc != nil {
		t.Fatalf("expected VersionedMigrationFunc to be nil when only WithMigrationFunc is used")
	}
}
