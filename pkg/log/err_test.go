package log_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/log"
)

func TestErrorGeneric(t *testing.T) {
	t.Run("Errors", func(t *testing.T) {
		const (
			tenantID = "tenant-1"
		)

		buf := new(bytes.Buffer)
		l := log.NewLogger("vega", "v1.0.0", "vega-123-123", log.WithWriter(buf))

		ctx := context.Background()
		ctx = context.WithValue(ctx, log.UserIDContextKey, "u1")
		ctx = context.WithValue(ctx, log.TraceIDContextKey, "t1")
		ctx = context.WithValue(ctx, log.TenantIDContextKey, tenantID)

		for _, tc := range [...]struct {
			name      string
			errLogger func(context.Context, error) error
			result    map[string]string
			ctx       context.Context
		}{
			{
				name:      "err-generic: with full context",
				ctx:       ctx,
				errLogger: l.ErrGeneric,
				result: map[string]string{
					"trace-id":        "t1",
					"log-level":       string(log.LevelError),
					"service-name":    "vega",
					"service-version": "v1.0.0",
					"hostname":        "vega-123-123",
					"user-id":         "u1",
					"tenant-id":       tenantID,
					"message":         bytes.ErrTooLarge.Error(),
				},
			},
			{
				name:      "err-internal: with full context",
				ctx:       ctx,
				errLogger: l.ErrInternal,
				result: map[string]string{
					"trace-id":        "t1",
					"log-level":       string(log.LevelError),
					"service-name":    "vega",
					"service-version": "v1.0.0",
					"hostname":        "vega-123-123",
					"event-type":      "err-internal",
					"user-id":         "u1",
					"tenant-id":       tenantID,
					"message":         bytes.ErrTooLarge.Error(),
				},
			},
			{
				name:      "err-io: with full context",
				ctx:       ctx,
				errLogger: l.ErrInputOutput,
				result: map[string]string{
					"trace-id":        "t1",
					"log-level":       string(log.LevelError),
					"service-name":    "vega",
					"service-version": "v1.0.0",
					"hostname":        "vega-123-123",
					"event-type":      "err-io",
					"user-id":         "u1",
					"tenant-id":       tenantID,
					"message":         bytes.ErrTooLarge.Error(),
				},
			},
			{
				name:      "err-user-auth: with full context",
				ctx:       ctx,
				errLogger: l.ErrUserAuth,
				result: map[string]string{
					"trace-id":        "t1",
					"log-level":       string(log.LevelError),
					"service-name":    "vega",
					"service-version": "v1.0.0",
					"hostname":        "vega-123-123",
					"event-type":      "err-user-auth",
					"user-id":         "u1",
					"tenant-id":       tenantID,
					"message":         bytes.ErrTooLarge.Error(),
				},
			},
			{
				name:      "err-input-validation: with full context",
				ctx:       ctx,
				errLogger: l.ErrInputValidation,
				result: map[string]string{
					"trace-id":        "t1",
					"log-level":       string(log.LevelError),
					"service-name":    "vega",
					"service-version": "v1.0.0",
					"hostname":        "vega-123-123",
					"event-type":      "err-input-validation",
					"user-id":         "u1",
					"tenant-id":       tenantID,
					"message":         bytes.ErrTooLarge.Error(),
				},
			},
			{
				name:      "err-oauth2client-auth: with full context",
				ctx:       ctx,
				errLogger: l.ErrOauth2ClientAuth,
				result: map[string]string{
					"trace-id":        "t1",
					"log-level":       string(log.LevelError),
					"service-name":    "vega",
					"service-version": "v1.0.0",
					"hostname":        "vega-123-123",
					"event-type":      "err-oauth2client-auth",
					"user-id":         "u1",
					"tenant-id":       tenantID,
					"message":         bytes.ErrTooLarge.Error(),
				},
			},
			{
				name:      "err-generic: with empty context",
				ctx:       context.TODO(),
				errLogger: l.ErrGeneric,
				result: map[string]string{
					"trace-id":        "",
					"log-level":       string(log.LevelError),
					"service-name":    "vega",
					"service-version": "v1.0.0",
					"hostname":        "vega-123-123",
					"user-id":         "",
					"tenant-id":       "",
					"message":         bytes.ErrTooLarge.Error(),
				},
			},
			{
				name:      "err-generic: with nil context",
				ctx:       context.TODO(),
				errLogger: l.ErrGeneric,
				result: map[string]string{
					"trace-id":        "",
					"log-level":       string(log.LevelError),
					"service-name":    "vega",
					"service-version": "v1.0.0",
					"hostname":        "vega-123-123",
					"user-id":         "",
					"tenant-id":       "",
					"message":         bytes.ErrTooLarge.Error(),
				},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {

				if err := tc.errLogger(tc.ctx, bytes.ErrTooLarge); err != nil {
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
