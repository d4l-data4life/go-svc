package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/gesundheitscloud/go-svc/pkg/db"
	"github.com/gesundheitscloud/go-svc/pkg/gormer"
)

func MigrationFunc(conn *gorm.DB) error {
	conn.Exec("CREATE SCHEMA IF NOT EXISTS \"testing\"")
	return conn.AutoMigrate(&gormer.Example{})
}

// InitializeTestDB connects to DB with transaction driver
func InitializeTestDB(t *testing.T) {
	dbOpts := db.NewConnection(
		db.WithHost("localhost"),
		db.WithPort("5432"),
		db.WithDatabaseName("test"),
		db.WithDatabaseSchema("testing"),
		db.WithUser("user"),
		db.WithPassword("test"),
		db.WithSSLMode("disable"),
		db.WithMigrationFunc(MigrationFunc),
		db.WithDriverFunc(db.TXDBPostgresDriver),
	)
	db.InitializeTestPostgres(dbOpts)
	assert.NotNil(t, db.Get(), "DB handle is nil")
	err := db.Ping()
	require.NoError(t, err)
}
