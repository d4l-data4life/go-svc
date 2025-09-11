package log_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/d4l-data4life/go-svc/pkg/log"
)

func TestAudit(t *testing.T) {
	t.Run("Audit", func(t *testing.T) {
		const (
			tenantID = "tenant-1"
		)

		ctx := context.Background()
		ctx = context.WithValue(ctx, log.UserIDContextKey, "u1")
		ctx = context.WithValue(ctx, log.TraceIDContextKey, "t1")
		ctx = context.WithValue(ctx, log.TenantIDContextKey, tenantID)

		expectedPrefix := `"log-level":"audit",` +
			`"trace-id":"t1",` +
			`"service-name":"vega",` +
			`"service-version":"v1.0.0",` +
			`"hostname":"vega-123-123",` +
			`"event-type":"audit",` +
			`"tenant-id":"` + tenantID + `",` +
			`"user-id":"u1",` +
			`"message":"permissions have changed"`

		tests := []struct {
			name          string
			ctx           context.Context
			message       string
			auditedObject interface{}
			want          string
		}{
			{
				name:          "audit: with full context and plain object",
				ctx:           ctx,
				message:       "permissions have changed",
				auditedObject: struct{ Value int }{5},
				want:          expectedPrefix + `,"object":"{\"Value\":5}"`,
			},
			{
				name:    "audit: with full context and richer object",
				ctx:     ctx,
				message: "permissions have changed",
				auditedObject: struct {
					User       string
					Permission int
				}{"henry", 0},
				want: expectedPrefix + `,"object":"{\"User\":\"henry\",\"Permission\":0}"`,
			},
			{
				name:    "audit: with full context and array object",
				ctx:     ctx,
				message: "permissions have changed",
				auditedObject: []struct {
					User       string
					Permission int
				}{
					{"henry", 0},
					{"john", 7},
				},
				want: expectedPrefix + `,"object":"[{\"User\":\"henry\",\"Permission\":0},{\"User\":\"john\",\"Permission\":7}]"`,
			},
			{
				name:          "audit: with full context and empty object",
				ctx:           ctx,
				message:       "permissions have changed",
				auditedObject: struct{}{},
				want:          expectedPrefix,
			},
			{
				name:          "audit: with full context and nil object",
				ctx:           ctx,
				message:       "permissions have changed",
				auditedObject: nil,
				want:          expectedPrefix,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				buffer := new(bytes.Buffer)
				auditLogger := log.NewLogger("vega", "v1.0.0", "vega-123-123", log.WithWriter(buffer)).Audit

				if err := auditLogger(tc.ctx, tc.message, tc.auditedObject); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				got := buffer.String()
				if !strings.Contains(got, tc.want) {
					t.Fatalf("Expected %v\n to contain: \n%v", got, tc.want)
				}
			})
		}
	})
}
