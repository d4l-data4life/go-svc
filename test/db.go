package test

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/d4l-data4life/go-svc/pkg/db"
)

func connectToDB(cfg *config) (*sql.DB, error) {
	const (
		attempts    = 3
		backoffTime = 3 * time.Second
	)

	sslMode := "disable"
	if cfg.PGUseSSL {
		sslMode = db.SSLVerifyFull
	}

	sqlOpenParams := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s sslmode=%s port=%d",
		cfg.PGUser, cfg.PGPassword, cfg.PGName, cfg.PGHost, sslMode, cfg.PGPort)

	var db *sql.DB
	var err error

	for i := 0; i < attempts; i++ {
		if db, err = sql.Open("postgres", sqlOpenParams); err == nil {
			// use ping to check if the database is ready to receive queries
			if err = db.Ping(); err == nil {
				break
			}
		}
		time.Sleep(backoffTime)
	}
	if err != nil {
		return nil, errors.Wrap(err, "error opening database")
	}
	return db, nil
}

func cleanSchema(ctx context.Context, db *sql.DB, schemaName string) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;", schemaName))
	return err
}

func cleanTable(ctx context.Context, db *sql.DB, tableName string) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName))
	return err
}

func getSchemaForTable(ctx context.Context, db *sql.DB, tableName string) (string, error) {
	query := `SELECT table_schema
		FROM information_schema.tables
		WHERE table_name = $1;`
	row := db.QueryRowContext(ctx, query, tableName)

	var foundSchema string
	if err := row.Scan(&foundSchema); err != nil {
		return "", err
	}

	return foundSchema, nil
}
