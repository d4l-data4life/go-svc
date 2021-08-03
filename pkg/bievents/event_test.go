package bievents

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
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
				Data: OnboardingData{
					CUC:         "cuc",
					SourceURL:   "https://abc.com",
					AccountType: Internal,
					Source:      "some-source",
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
				"data": map[string]interface{}{
					"cuc":          "cuc",
					"account-type": "internal",
					"source-url":   "https://abc.com",
					"source":       "some-source",
				},
			},
		},
		{
			name: "test an empty tenant id",
			value: Event{
				ActivityType: "login",
				UserID:       "def",
				Data: OnboardingData{
					CUC:         "cuc",
					SourceURL:   "https://abc.com",
					AccountType: Internal,
					Source:      "some-source",
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
				"data": map[string]interface{}{
					"cuc":          "cuc",
					"account-type": "internal",
					"source-url":   "https://abc.com",
					"source":       "some-source",
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

			if want, have := tc.result, res; !reflect.DeepEqual(want, have) {
				t.Errorf("expected values to be %q, got %q", want, have)
			}

		})
	}
}
