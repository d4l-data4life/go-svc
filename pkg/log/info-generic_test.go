package log_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/d4l-data4life/go-svc/pkg/log"
)

const (
	testMessage string = "Test message"
)

func TestInfoGeneric(t *testing.T) {
	t.Run("info generic", func(t *testing.T) {
		const (
			tenantID = "tenant-1"
		)
		buf := new(bytes.Buffer)
		l := log.NewLogger("test-svc", "v1.0.0", "test-svc-123-123", log.WithWriter(buf))

		ctx := context.Background()
		ctx = context.WithValue(ctx, log.UserIDContextKey, "u1")
		ctx = context.WithValue(ctx, log.TraceIDContextKey, "t1")
		ctx = context.WithValue(ctx, log.ClientIDContextKey, "c1")
		ctx = context.WithValue(ctx, log.TenantIDContextKey, tenantID)
		if err := l.InfoGeneric(ctx, testMessage); err != nil {
			t.Fatalf("logging: %v", err)
		}
		m := make(map[string]json.RawMessage)
		if err := json.NewDecoder(buf).Decode(&m); err != nil {
			t.Fatalf("unmarshaling JSON: %v", err)
		}

		for _, tc := range [...]struct {
			key   string
			value string
		}{
			{
				key:   "trace-id",
				value: `"t1"`,
			},
			{
				key:   "client-id",
				value: `"c1"`,
			},
			{
				key:   "log-level",
				value: fmt.Sprintf(`"%s"`, string(log.LevelInfo)),
			},
			{
				key:   "user-id",
				value: `"u1"`,
			},
			{
				key:   "service-name",
				value: `"test-svc"`,
			},
			{
				key:   "service-version",
				value: `"v1.0.0"`,
			},
			{
				key:   "hostname",
				value: `"test-svc-123-123"`,
			},
			{
				key:   "message",
				value: fmt.Sprintf(`"%s"`, testMessage),
			},
			{
				key:   "tenant-id",
				value: fmt.Sprintf(`"%s"`, tenantID),
			},
		} {
			t.Run("contains "+tc.key, func(t *testing.T) {
				if want, have := []byte(tc.value), m[tc.key]; !bytes.Equal(want, have) {
					t.Errorf("expected %q to be %q, found %q", tc.key, want, have)
				}
			})
		}
	})
}
