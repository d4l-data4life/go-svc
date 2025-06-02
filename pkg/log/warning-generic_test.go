package log_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/log"
)

func TestLogger_WarnGeneric(t *testing.T) {
	t.Run("Warn", func(t *testing.T) {
		const (
			tenantID = "tenant-1"
		)
		buf := new(bytes.Buffer)
		l := log.NewLogger("vega", "v1.0.0", "vega-123-123", log.WithWriter(buf))

		ctx := context.Background()
		ctx = context.WithValue(ctx, log.TraceIDContextKey, "t1")
		ctx = context.WithValue(ctx, log.UserIDContextKey, "u1")
		ctx = context.WithValue(ctx, log.ClientIDContextKey, "c1")
		ctx = context.WithValue(ctx, log.TenantIDContextKey, tenantID)
		basicMessage := "comment to the warning"
		exampleError := bytes.ErrTooLarge

		for _, tc := range [...]struct {
			name      string
			message   string
			err       error
			errLogger func(context.Context, string, error) error
			result    map[string]string
			ctx       context.Context
		}{
			{
				name:      "warning-generic: with full context",
				message:   basicMessage,
				err:       exampleError,
				ctx:       ctx,
				errLogger: l.WarnGeneric,
				result: map[string]string{
					"trace-id":        "t1",
					"log-level":       string(log.LevelWarning),
					"service-name":    "vega",
					"service-version": "v1.0.0",
					"hostname":        "vega-123-123",
					"user-id":         "u1",
					"message":         basicMessage,
					"error":           exampleError.Error(),
					"client-id":       "c1",
					"tenant-id":       tenantID,
				},
			},
			{
				name:      "warning-generic: with empty context",
				message:   basicMessage,
				err:       exampleError,
				ctx:       context.TODO(),
				errLogger: l.WarnGeneric,
				result: map[string]string{
					"trace-id":        "",
					"log-level":       string(log.LevelWarning),
					"service-name":    "vega",
					"service-version": "v1.0.0",
					"hostname":        "vega-123-123",
					"user-id":         "",
					"message":         basicMessage,
					"error":           exampleError.Error(),
					"tenant-id":       "",
				},
			},
			{
				name:      "warning-generic: with nil context and nil err",
				message:   basicMessage,
				err:       nil,
				ctx:       nil,
				errLogger: l.WarnGeneric,
				result: map[string]string{
					"trace-id":        "",
					"log-level":       string(log.LevelWarning),
					"service-name":    "vega",
					"service-version": "v1.0.0",
					"hostname":        "vega-123-123",
					"user-id":         "",
					"message":         basicMessage,
					"error":           "",
					"tenant-id":       "",
				},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				if err := tc.errLogger(tc.ctx, tc.message, tc.err); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				m := make(map[string]string)
				if err := json.NewDecoder(buf).Decode(&m); err != nil {
					t.Fatalf("unmarshaling JSON: %v", err)
				}
				// Note that tc.result doesn't have timestamp in it as of
				// now as it is hard to compare time.Now()
				if want, have := tc.result, m; !isMapEqual(want, have) {
					t.Errorf("expected to be %q, found %q", want, have)
				}
			})
		}
	})
}
