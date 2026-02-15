package db

import (
	"testing"

	"gorm.io/gorm"
)

func TestRunMigrationRejectsBothLegacyAndVersionedFuncs(t *testing.T) {
	legacyFn := func(_ *gorm.DB) error { return nil }
	versionedFn := func(_ *gorm.DB, _ uint) error { return nil }

	err := runMigration(&gorm.DB{}, legacyFn, versionedFn, 1, true)
	if err == nil {
		t.Fatalf("expected error")
	}
}
