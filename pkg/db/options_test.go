package db

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gorm.io/gorm"
)

func TestWithMigrationFuncDoesNotSetVersionedMigrationFunc(t *testing.T) {
	fn := func(_ *gorm.DB) error { return nil }

	opts := NewConnection(
		WithMigrationVersion(2),
		WithMigrationFunc(fn),
	)

	require.NotNil(t, opts.MigrationFunc)
	require.Nil(t, opts.VersionedMigrationFunc)
}
