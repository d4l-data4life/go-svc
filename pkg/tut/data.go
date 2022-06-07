package tut

import (
	"github.com/gofrs/uuid"
)

// UniqueTestEmail creates a test email including an uuid that can be safely used for
// tests with very unlikely collision.
func UniqueTestEmail() string {
	return "test+" + uuid.Must(uuid.NewV4()).String() + "@data4life.care"
}
