package log_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	. "github.com/d4l-data4life/go-svc/pkg/log"
)

func TestAuditCreate(t *testing.T) {
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
		`"audit-log-type":"change",` +
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
		value          interface{}
		additionalData interface{}
		want           string
	}{
		{
			name: "should work with string value",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			ownerID:        "owner-1",
			resourceType:   "res-type-1",
			resourceID:     "res-1",
			value:          "new value",
			additionalData: "extra info",
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"owner-id":"owner-1",` +
				`"event-type":"create",` +
				`"resource-type":"res-type-1",` +
				`"value-new":"new value",` +
				`"additional-data":"extra info",` +
				`"resource-id":"res-1"`,
		},
		{
			name: "should work with int value",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			ownerID:        "owner-1",
			resourceType:   "res-type-1",
			resourceID:     "res-1",
			value:          1,
			additionalData: "extra info",
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"owner-id":"owner-1",` +
				`"event-type":"create",` +
				`"resource-type":"res-type-1",` +
				`"value-new":1,` +
				`"additional-data":"extra info",` +
				`"resource-id":"res-1"`,
		},
		{
			name: "should work with complex value",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			subjectID:    "sub-1",
			ownerID:      "owner-1",
			resourceType: "res-type-1",
			resourceID:   "res-1",
			value: exampleStruct{
				IntField:    1,
				StringField: "a",
				BoolField:   true,
			},
			additionalData: "extra info",
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"owner-id":"owner-1",` +
				`"event-type":"create",` +
				`"resource-type":"res-type-1",` +
				`"value-new":{"int-field":1,"bool-field":true,"string-field":"a"},` +
				`"additional-data":"extra info",` +
				`"resource-id":"res-1"`,
		},
		{
			name: "should work without additional data",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			ownerID:      "owner-1",
			resourceType: "res-type-1",
			resourceID:   "res-1",
			value: exampleStruct{
				IntField:    1,
				StringField: "a",
				BoolField:   true,
			},
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"owner-id":"owner-1",` +
				`"event-type":"create",` +
				`"resource-type":"res-type-1",` +
				`"value-new":{"int-field":1,"bool-field":true,"string-field":"a"},` +
				`"resource-id":"res-1"`,
		},
		{
			name: "extra subject ID should override the default one from context",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			subjectID:    "sub-2",
			ownerID:      "owner-1",
			resourceType: "res-type-1",
			resourceID:   "res-1",
			value: exampleStruct{
				IntField:    1,
				StringField: "a",
				BoolField:   true,
			},
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-2",` +
				`"owner-id":"owner-1",` +
				`"event-type":"create",` +
				`"resource-type":"res-type-1",` +
				`"value-new":{"int-field":1,"bool-field":true,"string-field":"a"},` +
				`"resource-id":"res-1"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buffer := new(bytes.Buffer)
			logger := NewLogger(
				"svc",
				"v1.0.0",
				"host-1",
				WithWriter(buffer),
				WithPodName(podName),
				WithEnv(environment),
			)

			extras := make([]ExtraAuditInfoProvider, 0)
			if tc.additionalData != nil {
				extras = append(extras, AdditionalData(tc.additionalData))
			}
			if tc.subjectID != "" {
				extras = append(extras, SubjectID(tc.subjectID))
			}

			if err := logger.AuditCreate(
				tc.ctx,
				tc.ownerID,
				tc.resourceType,
				tc.resourceID,
				tc.value,
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

func TestAuditUpdate(t *testing.T) {
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
		`"audit-log-type":"change",` +
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
		value          interface{}
		oldValue       interface{}
		additionalData interface{}
		want           string
	}{
		{
			name: "should work with complex values",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			ownerID:      "owner-1",
			resourceType: "res-type-1",
			resourceID:   "res-1",
			value: exampleStruct{
				IntField:    1,
				StringField: "a1",
				BoolField:   true,
			},
			oldValue: exampleStruct{
				IntField:    -1,
				StringField: "a0",
				BoolField:   false,
			},
			additionalData: nil,
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"owner-id":"owner-1",` +
				`"event-type":"update",` +
				`"resource-type":"res-type-1",` +
				`"value-old":{"int-field":-1,"bool-field":false,"string-field":"a0"},` +
				`"value-new":{"int-field":1,"bool-field":true,"string-field":"a1"},` +
				`"resource-id":"res-1"`,
		},
		{
			name: "should work with nil values",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			ownerID:        "owner-1",
			resourceType:   "res-type-1",
			resourceID:     "res-1",
			value:          nil,
			oldValue:       nil,
			additionalData: nil,
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"owner-id":"owner-1",` +
				`"event-type":"update",` +
				`"resource-type":"res-type-1",` +
				`"resource-id":"res-1"`,
		},
		{
			name: "subject ID as extra should override the context one",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			subjectID:      "sub-2",
			ownerID:        "owner-1",
			resourceType:   "res-type-1",
			resourceID:     "res-1",
			value:          nil,
			oldValue:       nil,
			additionalData: nil,
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-2",` +
				`"owner-id":"owner-1",` +
				`"event-type":"update",` +
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
				extras = append(extras, SubjectID(tc.subjectID))
			}
			if tc.oldValue != nil {
				extras = append(extras, OldValue(tc.oldValue))
			}

			if err := logger.AuditUpdate(
				tc.ctx,
				tc.ownerID,
				tc.resourceType,
				tc.resourceID,
				tc.value,
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

func TestAuditDelete(t *testing.T) {
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
		`"audit-log-type":"change",` +
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
		oldValue       interface{}
		additionalData interface{}
		want           string
	}{
		{
			name: "should work with complex old value",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			ownerID:      "owner-1",
			resourceType: "res-type-1",
			resourceID:   "res-1",
			oldValue: exampleStruct{
				IntField:    -1,
				StringField: "a0",
				BoolField:   false,
			},
			additionalData: nil,
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"owner-id":"owner-1",` +
				`"event-type":"delete",` +
				`"resource-type":"res-type-1",` +
				`"value-old":{"int-field":-1,"bool-field":false,"string-field":"a0"},` +
				`"resource-id":"res-1"`,
		},
		{
			name: "should work without an old values",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			ownerID:        "owner-1",
			resourceType:   "res-type-1",
			resourceID:     "res-1",
			oldValue:       nil,
			additionalData: nil,
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"owner-id":"owner-1",` +
				`"event-type":"delete",` +
				`"resource-type":"res-type-1",` +
				`"resource-id":"res-1"`,
		},
		{
			name: "subject ID as extra should override the context one",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			subjectID:      "sub-2",
			ownerID:        "owner-1",
			resourceType:   "res-type-1",
			resourceID:     "res-1",
			oldValue:       nil,
			additionalData: nil,
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-2",` +
				`"owner-id":"owner-1",` +
				`"event-type":"delete",` +
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
				extras = append(extras, SubjectID(tc.subjectID))
			}
			if tc.oldValue != nil {
				extras = append(extras, OldValue(tc.oldValue))
			}

			if err := logger.AuditDelete(
				tc.ctx,
				tc.ownerID,
				tc.resourceType,
				tc.resourceID,
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

func TestAuditBulkDelete(t *testing.T) {
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
		`"audit-log-type":"change",` +
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
		resourceIDs    []string
		oldValue       interface{}
		additionalData interface{}
		want           string
	}{
		{
			name: "should work with one resource ID",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			ownerID:        "owner-1",
			resourceType:   "res-type-1",
			resourceIDs:    []string{"res-1"},
			oldValue:       nil,
			additionalData: nil,
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"owner-id":"owner-1",` +
				`"event-type":"delete",` +
				`"resource-type":"res-type-1",` +
				`"resource-ids":["res-1"]`,
		},
		{
			name: "should work with multiple resource IDs",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			ownerID:        "owner-1",
			resourceType:   "res-type-1",
			resourceIDs:    []string{"res-1", "res-2"},
			oldValue:       nil,
			additionalData: nil,
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"owner-id":"owner-1",` +
				`"event-type":"delete",` +
				`"resource-type":"res-type-1",` +
				`"resource-ids":["res-1","res-2"]`,
		},
		{
			name: "should work with complex old value",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			ownerID:      "owner-1",
			resourceType: "res-type-1",
			resourceIDs:  []string{"res-1"},
			oldValue: exampleStruct{
				IntField:    -1,
				StringField: "a0",
				BoolField:   false,
			},
			additionalData: nil,
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"owner-id":"owner-1",` +
				`"event-type":"delete",` +
				`"resource-type":"res-type-1",` +
				`"value-old":{"int-field":-1,"bool-field":false,"string-field":"a0"},` +
				`"resource-ids":["res-1"]`,
		},
		{
			name: "should work without an old values",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			ownerID:        "owner-1",
			resourceType:   "res-type-1",
			resourceIDs:    []string{"res-1"},
			oldValue:       nil,
			additionalData: nil,
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-1",` +
				`"owner-id":"owner-1",` +
				`"event-type":"delete",` +
				`"resource-type":"res-type-1",` +
				`"resource-ids":["res-1"]`,
		},
		{
			name: "subject ID as extra should override the context one",
			ctx: context.WithValue(
				context.WithValue(ctx, CallerIPContextKey, "0.0.0.1"),
				UserIDContextKey, "sub-1",
			),
			subjectID:      "sub-2",
			ownerID:        "owner-1",
			resourceType:   "res-type-1",
			resourceIDs:    []string{"res-1"},
			oldValue:       nil,
			additionalData: nil,
			want: expectedPrefix +
				`"caller-ip":"0.0.0.1",` +
				`"subject-id":"sub-2",` +
				`"owner-id":"owner-1",` +
				`"event-type":"delete",` +
				`"resource-type":"res-type-1",` +
				`"resource-ids":["res-1"]`,
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
			if tc.oldValue != nil {
				extras = append(extras, OldValue(tc.oldValue))
			}

			if err := logger.AuditBulkDelete(
				tc.ctx,
				tc.ownerID,
				tc.resourceType,
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
