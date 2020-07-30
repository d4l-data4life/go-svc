package db

import (
	"fmt"

	"github.com/jinzhu/gorm"
	// Blank import required by gorm
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// InitializeTest connects to an inmemory sqlite for testing
func InitializeTest(migFn MigrationFunc) {
	conn, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		fmt.Printf("Test DB connection error: %s", err.Error())
	}

	db = conn
	// 'foreign_keys = off' is default setting in SQLite.
	db.Exec("PRAGMA foreign_keys = ON")
	err = migrate(conn, migFn)
	if err != nil {
		fmt.Printf("Test DB migration error: %s", err.Error())
	}
}
