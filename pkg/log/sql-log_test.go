package log_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/log"
)

func TestSqlLog(t *testing.T) {
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

	for _, tc := range [...]struct {
		name    string
		logData log.SqlLogData
		result  map[string]string
	}{
		{
			name: "Log sql Exec no Args",
			logData: log.SqlLogData{
				Action:   "exec",
				Duration: time.Duration(5),
				Error:    errors.New(""),
				Sql:      "SELECT * FROM keymgmt.pki",
			},
			result: map[string]string{
				"action":   `"exec"`,
				"duration": fmt.Sprintf("%d", time.Duration(5).Milliseconds()),
				"sql":      `"SELECT * FROM keymgmt.pki"`,
			},
		},
		{
			name: "Log sql Open",
			logData: log.SqlLogData{
				Action:   "open",
				Duration: time.Duration(3),
				Error:    errors.New(""),
			},
			result: map[string]string{
				"action":   `"open"`,
				"duration": fmt.Sprintf("%d", time.Duration.Milliseconds(3)),
			},
		},
		{
			name: "Log sql Exec with Args",
			logData: log.SqlLogData{
				Action:   "exec",
				Duration: time.Duration(6),
				Error:    errors.New(""),
				Sql:      "SELECT * FROM keymgmt.pki WHERE USER_ID = $1",
				Args:     "u1",
			},
			result: map[string]string{
				"action":   `"exec"`,
				"duration": fmt.Sprintf("%d", time.Duration.Milliseconds(6)),
				"sql":      `"SELECT * FROM keymgmt.pki WHERE USER_ID = $1"`,
				"args":     `"u1"`,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := l.SqlLog(ctx, tc.logData); err != nil {
				t.Fatalf("logging: %v", err)
			}
			m := make(map[string]json.RawMessage)
			if err := json.NewDecoder(buf).Decode(&m); err != nil {
				t.Fatalf("unmarshaling JSON: %v", err)
			}

			if want, have := []byte(tc.result["action"]), m["action"]; !bytes.Equal(want, have) {
				t.Errorf("expected %q to be %q, found %q", "action", want, have)
			}

			if want, have := []byte(tc.result["duration"]), m["duration"]; !bytes.Equal(want, have) {
				t.Errorf("expected %q to be %q, found %q", "duration", want, have)
			}

			if want, have := []byte(tc.result["error"]), m["error"]; !bytes.Equal(want, have) {
				t.Errorf("expected %q to be %q, found %q", "error", want, have)
			}

			if want, have := []byte(tc.result["sql"]), m["sql"]; !bytes.Equal(want, have) {
				t.Errorf("expected %q to be %q, found %q", "sql", want, have)
			}

			if want, have := []byte(tc.result["args"]), m["args"]; !bytes.Equal(want, have) {
				t.Errorf("expected %q to be %q, found %q", "args", want, have)
			}
		})
	}
}

func TestSqlLogContextData(t *testing.T) {
	t.Run("Sql Logging Test Context Data", func(t *testing.T) {
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

		logData := log.SqlLogData{
			Action:   "exec",
			Duration: time.Duration(5),
			Error:    errors.New(""),
			Sql:      "SELECT * FROM keymgmt.PKI",
		}

		if err := l.SqlLog(ctx, logData); err != nil {
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
				key:   "event-type",
				value: `"sql-log"`,
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
