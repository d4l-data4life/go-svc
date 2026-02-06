package db

import (
	"testing"

	"gorm.io/gorm"
)

func TestWithMigrationFuncWrapsVersioned(t *testing.T) {
	calls := 0
	fn := func(_ *gorm.DB) error {
		calls++
		return nil
	}

	opts := NewConnection(
		WithMigrationVersion(2),
		WithMigrationFunc(fn),
	)
	if opts.VersionedMigrationFunc == nil {
		t.Fatalf("expected VersionedMigrationFunc to be set")
	}

	_ = opts.VersionedMigrationFunc(nil, 1)
	if calls != 0 {
		t.Fatalf("expected no calls for version 1, got %d", calls)
	}

	_ = opts.VersionedMigrationFunc(nil, 2)
	if calls != 1 {
		t.Fatalf("expected one call for version 2, got %d", calls)
	}
}
