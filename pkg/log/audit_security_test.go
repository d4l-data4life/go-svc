package log_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	. "github.com/gesundheitscloud/go-svc/pkg/log"
)

func TestAuditSecurity(t *testing.T) {
	const (
		podName     = "pod-1"
		environment = "test-env"
		tenantID    = "tenant-1"
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, TraceIDContextKey, "t1")
	ctx = context.WithValue(ctx, ClientIDContextKey, "client-1")
	ctx = context.WithValue(ctx, TenantIDContextKey, tenantID)

	expectedPrefix := `"log-type":"audit",` +
		`"audit-log-type":"security",` +
		`"trace-id":"t1",` +
		`"service-name":"svc",` +
		`"service-version":"v1.0.0",` +
		`"hostname":"host-1",` +
		`"pod-name":"pod-1",` +
		`"environment":"test-env",` +
		`"client-id":"client-1",` +
		`"tenant-id":"` + tenantID + `",`

	type exampleStruct struct {
		IntField    int    `json:"int-field"`
		BoolField   bool   `json:"bool-field"`
		StringField string `json:"string-field"`
	}

	tests := []struct {
		name           string
		ctx            context.Context
		subjectID      string
		securityEvent  string
		successful     bool
		message        string
		additionalData interface{}
		want           string
	}{
		{
			name: "should work with complex additional data",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			securityEvent: "log-in",
			successful:    true,
			message:       "user sub-1 logged in",
			additionalData: exampleStruct{
				IntField:    -1,
				StringField: "a0",
				BoolField:   false,
			},
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"security-event":"log-in",` +
				`"successful":true,` +
				`"message":"user sub-1 logged in",` +
				`"additional-data":{"int-field":-1,"bool-field":false,"string-field":"a0"}`,
		},
		{
			name: "should work without additional data",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			securityEvent:  "log-in",
			successful:     true,
			message:        "user sub-1 logged in",
			additionalData: nil,
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"security-event":"log-in",` +
				`"successful":true,` +
				`"message":"user sub-1 logged in"`,
		},
		{
			name: "should work with a message",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			securityEvent: "log-in",
			successful:    true,
			message:       "user sub-1 logged in",
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"security-event":"log-in",` +
				`"successful":true,` +
				`"message":"user sub-1 logged in"`,
		},
		{
			name: "should work without a message",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			securityEvent:  "log-in",
			successful:     true,
			message:        "",
			additionalData: nil,
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"security-event":"log-in",` +
				`"successful":true`,
		},
		{
			name: "should work with overridden subject ID",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			subjectID:     "sub-2",
			securityEvent: "log-in",
			successful:    true,
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-2",` +
				`"security-event":"log-in",` +
				`"successful":true`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buffer := new(bytes.Buffer)
			logger := NewLogger("svc", "v1.0.0", "host-1", WithWriter(buffer), WithPodName(podName), WithEnv(environment))

			extras := make([]ExtraAuditInfoProvider, 0)
			if tc.additionalData != nil {
				extras = append(extras, AdditionalData(tc.additionalData))
			}
			if tc.subjectID != "" {
				extras = append(extras, SubjectID(tc.subjectID))
			}
			if tc.message != "" {
				extras = append(extras, Message(tc.message))
			}

			if err := logger.AuditSecurity(
				tc.ctx,
				tc.securityEvent,
				tc.successful,
				extras...,
			); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := buffer.String()
			if !strings.Contains(got, tc.want) {
				t.Fatalf("Expected %v\n to contain: \n%v", got, tc.want)
			}
		})
	}
}
