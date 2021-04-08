package log_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	. "github.com/gesundheitscloud/go-svc/pkg/log"
)

func TestAuditRead(t *testing.T) {
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
		`"audit-log-type":"access",` +
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
		ownerID        string
		resourceType   string
		resourceID     string
		additionalData interface{}
		want           string
	}{
		{
			name: "should work with complex additional data",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			ownerID:      "owner-1",
			resourceType: "res-type-1",
			resourceID:   "res-1",
			additionalData: exampleStruct{
				IntField:    -1,
				StringField: "a0",
				BoolField:   false,
			},
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"owner-id":"owner-1",` +
				`"event-type":"read",` +
				`"resource-type":"res-type-1",` +
				`"additional-data":{"int-field":-1,"bool-field":false,"string-field":"a0"},` +
				`"resource-id":"res-1"`,
		},
		{
			name: "should work without additional data",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			ownerID:        "owner-1",
			resourceType:   "res-type-1",
			resourceID:     "res-1",
			additionalData: nil,
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"owner-id":"owner-1",` +
				`"event-type":"read",` +
				`"resource-type":"res-type-1",` +
				`"resource-id":"res-1"`,
		},
		{
			name: "should work with overridden subject ID",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			subjectID:    "sub-2",
			ownerID:      "owner-1",
			resourceType: "res-type-1",
			resourceID:   "res-1",
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-2",` +
				`"owner-id":"owner-1",` +
				`"event-type":"read",` +
				`"resource-type":"res-type-1",` +
				`"resource-id":"res-1"`,
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
				extras = append(extras, SubjectID(testStringer(tc.subjectID)))
			}

			if err := logger.AuditRead(
				tc.ctx,
				testStringer(tc.ownerID),
				testStringer(tc.resourceType),
				testStringer(tc.resourceID),
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

func TestAuditBulkRead(t *testing.T) {
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
		`"audit-log-type":"access",` +
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
		ownerID        string
		resourceType   string
		resourceIDs    []fmt.Stringer
		additionalData interface{}
		want           string
	}{
		{
			name: "should work with multiple resources",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			ownerID:      "owner-1",
			resourceType: "res-type-1",
			resourceIDs:  []fmt.Stringer{testStringer("res-1"), testStringer("res-2")},
			additionalData: exampleStruct{
				IntField:    -1,
				StringField: "a0",
				BoolField:   false,
			},
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"owner-id":"owner-1",` +
				`"event-type":"read",` +
				`"resource-type":"res-type-1",` +
				`"additional-data":{"int-field":-1,"bool-field":false,"string-field":"a0"},` +
				`"resource-ids":["res-1","res-2"]`,
		},
		{
			name: "should work without additional data",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			ownerID:        "owner-1",
			resourceType:   "res-type-1",
			resourceIDs:    []fmt.Stringer{testStringer("res-1"), testStringer("res-2")},
			additionalData: nil,
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"owner-id":"owner-1",` +
				`"event-type":"read",` +
				`"resource-type":"res-type-1",` +
				`"resource-ids":["res-1","res-2"]`,
		},
		{
			name: "should work with overridden subject ID",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			subjectID:    "sub-2",
			ownerID:      "owner-1",
			resourceType: "res-type-1",
			resourceIDs:  []fmt.Stringer{testStringer("res-1"), testStringer("res-2")},
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-2",` +
				`"owner-id":"owner-1",` +
				`"event-type":"read",` +
				`"resource-type":"res-type-1",` +
				`"resource-ids":["res-1","res-2"]`,
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
				extras = append(extras, SubjectID(testStringer(tc.subjectID)))
			}

			if err := logger.AuditBulkRead(
				tc.ctx,
				testStringer(tc.ownerID),
				testStringer(tc.resourceType),
				tc.resourceIDs,
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
