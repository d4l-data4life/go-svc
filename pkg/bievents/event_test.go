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

	"github.com/gesundheitscloud/go-svc/pkg/jwt"
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
				ActivityType:       "login",
				UserID:             "def",
				TenantID:           tenantID,
				ConsentDocumentKey: "consent",
				State:              Success,
				Data: UserRegisterData{
					CucID:       "cuc_1",
					SourceURL:   "https://abc.com",
					AccountType: Internal,
					ClientID:    "client_1",
				},
			},
			result: map[string]interface{}{
				"service-name":         serviceName,
				"service-version":      version,
				"hostname":             hostname,
				"tenant-id":            tenantID,
				"event-type":           "bi-event",
				"state":                "success",
				"activity-type":        "login",
				"user-id":              "def",
				"consent-document-key": "consent",
				"event-source":         "",
				"session-id":           "",
				"data": map[string]interface{}{
					"cuc-id":       "cuc_1",
					"account-type": "internal",
					"source-url":   "https://abc.com",
					"client-id":    "client_1",
				},
			},
		},
		{
			name: "test an empty tenant id",
			value: Event{
				ActivityType: "login",
				UserID:       "def",
				Data: UserRegisterData{
					CucID:       "cuc_2",
					SourceURL:   "https://abc.com",
					AccountType: Internal,
					ClientID:    "client_1",
				},
			},
			result: map[string]interface{}{
				"service-name":         serviceName,
				"service-version":      version,
				"hostname":             hostname,
				"tenant-id":            "",
				"event-type":           "bi-event",
				"activity-type":        "login",
				"user-id":              "def",
				"consent-document-key": "",
				"event-source":         "",
				"session-id":           "",
				"data": map[string]interface{}{
					"cuc-id":       "cuc_2",
					"account-type": "internal",
					"source-url":   "https://abc.com",
					"client-id":    "client_1",
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
	appID := uuid.Must(uuid.NewV4())

	e := NewEventEmitter(serviceName, version, hostname, WithWriter(buf))

	err := e.LogCtx(jwt.NewContext(context.Background(), &jwt.Claims{UserID: userID, AppID: appID}), Event{})
	assert.NoError(t, err)

	var res map[string]interface{}
	err = json.NewDecoder(buf).Decode(&res)
	assert.NoError(t, err)

	assert.True(t, strings.Contains(fmt.Sprintf("%s", res["event-source"]), "pkg/bievents/event_test.go"))

	if v, ok := res["session-id"]; ok {
		assert.Equal(t, 64, len(fmt.Sprintf("%s", v)))
	}

	if v, ok := res["user-id"]; ok {
		assert.Equal(t, userID.String(), fmt.Sprintf("%s", v))
	}
}
