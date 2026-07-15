package bievents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/d4l-data4life/go-svc/pkg/log"
)

func TestLog(t *testing.T) {
	buf := new(bytes.Buffer)
	serviceName := "vega"
	version := "v1.0.0"
	hostname := "vega-123-123"
	tenantID := "some-tenant-id"

	e := NewEventEmitter(serviceName, version, hostname, WithWriter(buf))

	for _, tc := range [...]struct {
		name   string
		value  Event
		result map[string]interface{}
	}{
		{
			name: "test a specific type",
			value: Event{
				ActivityType:       LoginEmail,
				UserID:             "def",
				TenantID:           tenantID,
				ConsentDocumentKey: "consent",
				State:              Success,
				Data: LoginEmailData{
					ClientID:  "client_1",
					SourceURL: "https://abc.com",
				},
			},
			result: map[string]interface{}{
				"service-name":         serviceName,
				"service-version":      version,
				"hostname":             hostname,
				"tenant-id":            tenantID,
				"event-type":           "bi-event",
				"state":                "success",
				"activity-type":        "login-email",
				"user-id":              "def",
				"consent-document-key": "consent",
				"event-source":         "",
				"session-id":           "",
				"data": map[string]interface{}{
					"client-id":  "client_1",
					"source-url": "https://abc.com",
				},
			},
		},
		{
			name: "test an empty tenant id",
			value: Event{
				ActivityType: LoginEmail,
				UserID:       "def",
				Data: LoginEmailData{
					ClientID:  "client_1",
					SourceURL: "https://abc.com",
				},
			},
			result: map[string]interface{}{
				"service-name":         serviceName,
				"service-version":      version,
				"hostname":             hostname,
				"tenant-id":            "",
				"event-type":           "bi-event",
				"activity-type":        "login-email",
				"user-id":              "def",
				"consent-document-key": "",
				"event-source":         "",
				"session-id":           "",
				"data": map[string]interface{}{
					"client-id":  "client_1",
					"source-url": "https://abc.com",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := e.Log(tc.value)
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			var res map[string]interface{}

			err = json.NewDecoder(buf).Decode(&res)
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			// Ignore timestamp for comparison.
			delete(res, "timestamp")
			delete(res, "event-id")

			assert.True(t, strings.Contains(fmt.Sprintf("%s", res["event-source"]), "pkg/bievents/event_test.go"))
			res["event-source"] = ""

			if want, have := tc.result, res; !reflect.DeepEqual(want, have) {
				t.Errorf("expected values to be %q, got %q", want, have)
			}

		})
	}
}

func TestGetEventSource(t *testing.T) {
	eventSource := GetEventSource(0)
	parts := strings.Split(eventSource, ":")
	assert.True(t, strings.HasSuffix(parts[0], "pkg/bievents/event.go"))
	_, err := strconv.Atoi(parts[1])
	assert.NoError(t, err)
}

func Test_LogCtx(t *testing.T) {
	buf := new(bytes.Buffer)
	serviceName := "vega"
	version := "v1.0.0"
	hostname := "vega-123-123"
	userID := uuid.Must(uuid.NewV4())

	e := NewEventEmitter(serviceName, version, hostname, WithWriter(buf))

	ctx := context.WithValue(context.Background(), log.UserIDContextKey, userID.String())
	err := e.LogCtx(ctx, Event{})
	assert.NoError(t, err)

	var res map[string]interface{}
	err = json.NewDecoder(buf).Decode(&res)
	assert.NoError(t, err)

	assert.True(t, strings.Contains(fmt.Sprintf("%s", res["event-source"]), "pkg/bievents/event_test.go"))

	// UserID is filled from the context when the caller left it empty.
	if v, ok := res["user-id"]; ok {
		assert.Equal(t, userID.String(), fmt.Sprintf("%s", v))
	}

	// SessionID is no longer derived automatically.
	assert.Empty(t, res["session-id"])
}
