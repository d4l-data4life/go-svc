package client

import (
	"context"
	"strings"
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/log"
	"github.com/go-test/deep"
	uuid "github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTestNotfifcationMock_GetNotifiedUsers(t *testing.T) {
	// for the sake of the assertions, lets make the user IDs sorted
	user1 := uuid.FromStringOrNil("22173f04-d443-4545-a883-680afd305141")
	user2 := uuid.FromStringOrNil("461603ab-b71d-472f-8c1e-b965defdc6c7")
	user3 := uuid.FromStringOrNil("519d75e2-61a5-44c2-93cd-476886cd5091")
	// non-subscriber, just has entry in user-preferences
	user4 := uuid.FromStringOrNil("879847b7-3907-4f6a-8ae1-301b65582ebf")
	upCli := NewUserPreferencesMockWithState(map[uuid.UUID]struct{ K, V string }{
		user1: {"global.language", "en"},
		user2: {"global.language", "en"},
		user3: {"global.language", "de"},
		user4: {"global.language", "fr"},
	})
	csCli := NewConsentMock(map[uuid.UUID]string{
		user1: EventConsent,
		user2: EventRevoke,
		user3: EventConsent,
		user4: ConsentUnknown,
	})

	type call struct {
		templateKey        string
		language           string
		languageSettingKey string
		consentGuardKey    string
		subscribers        []uuid.UUID
	}
	tests := []struct {
		name  string
		calls []call
		want  NotifiedUsers
	}{
		{
			"Initial empty state",
			[]call{},
			map[string]map[string][]uuid.UUID{},
		},
		{
			"Single key two languages",
			[]call{
				{"key1", "en", "global.language", "consentKey", []uuid.UUID{user1, user2}},
				{"key1", "de", "global.language", "consentKey", []uuid.UUID{user3}},
			},
			map[string]map[string][]uuid.UUID{
				"key1": {
					"en": []uuid.UUID{user1, user2},
					"de": []uuid.UUID{user3},
				},
			},
		},
		{
			"Two keys two languages",
			[]call{
				{"key1", "en", "global.language", "consentKey", []uuid.UUID{user1, user2}},
				{"key1", "de", "global.language", "consentKey", []uuid.UUID{user3}},
			},
			map[string]map[string][]uuid.UUID{
				"key1": {
					"en": []uuid.UUID{user1, user2},
					"de": []uuid.UUID{user3},
				},
			},
		},
		{
			"Null conditions",
			[]call{{"key1", "", "global.language", "consentKey", []uuid.UUID{user1, user2}}},
			map[string]map[string][]uuid.UUID{
				"key1": {
					"en": []uuid.UUID{user1, user2},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := NewNotificationMockWithClients(upCli, csCli)
			for i := 0; i < len(tt.calls); i++ {
				_, err := c.SendTemplated(
					context.Background(),
					tt.calls[i].templateKey,
					tt.calls[i].language,
					tt.calls[i].languageSettingKey,
					tt.calls[i].consentGuardKey,
					"",
					"",
					nil,
					tt.calls[i].subscribers...)
				assert.NoError(t, err)
			}
			got := c.GetNotifiedUsers()
			assert.EqualValuesf(t, tt.want, got, "Notified users should match")
		})
	}
}

func TestTestNotfifcationMock_LegacyGetNotifiedUsers(t *testing.T) {
	// for the sake of the assertions, lets make the user IDs sorted
	user1 := uuid.FromStringOrNil("22173f04-d443-4545-a883-680afd305141")
	user2 := uuid.FromStringOrNil("461603ab-b71d-472f-8c1e-b965defdc6c7")
	user3 := uuid.FromStringOrNil("519d75e2-61a5-44c2-93cd-476886cd5091")

	type call struct {
		templateKey string
		language    string
		subscribers []uuid.UUID
	}
	tests := []struct {
		name  string
		calls []call
		want  NotifiedUsers
	}{
		{
			"Initial empty state",
			[]call{},
			map[string]map[string][]uuid.UUID{},
		},
		{
			"Single key two languages",
			[]call{
				{"key1", "en", []uuid.UUID{user1, user2}},
				{"key1", "de", []uuid.UUID{user3}},
			},
			map[string]map[string][]uuid.UUID{
				"key1": {
					"en": []uuid.UUID{user1, user2},
					"de": []uuid.UUID{user3},
				},
			},
		},
		{
			"Two keys two languages",
			[]call{
				{"key1", "en", []uuid.UUID{user1, user2}},
				{"key1", "de", []uuid.UUID{user3}},
			},
			map[string]map[string][]uuid.UUID{
				"key1": {
					"en": []uuid.UUID{user1, user2},
					"de": []uuid.UUID{user3},
				},
			},
		},
		{
			"Null conditions",
			[]call{{"key1", "", []uuid.UUID{}}},
			map[string]map[string][]uuid.UUID{
				"key1": {
					"": []uuid.UUID{},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := NewNotificationMockLegacy()
			for i := 0; i < len(tt.calls); i++ {
				_ = c.SendTemplated(tt.calls[i].templateKey, tt.calls[i].language, nil, tt.calls[i].subscribers...)
			}
			got := c.GetNotifiedUsers()
			assert.EqualValuesf(t, tt.want, got, "Notified users should match")
		})
	}
}

func TestNotificationMock_SendTemplated(t *testing.T) {
	// for the sake of the assertions, lets make the user IDs sorted
	user1 := uuid.FromStringOrNil("22173f04-d443-4545-a883-680afd305141")
	user2 := uuid.FromStringOrNil("461603ab-b71d-472f-8c1e-b965defdc6c7")
	user3 := uuid.FromStringOrNil("519d75e2-61a5-44c2-93cd-476886cd5091")

	type args struct {
		ctx                context.Context
		templateKey        string
		language           string
		languageSettingKey string
		consentGuardKey    string
		minConsentVersion  string
		payload            map[string]interface{}
		subscribers        []uuid.UUID
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantStatus NotificationStatus
	}{
		{
			name: "Checking consent using mock",
			args: args{
				ctx:                context.Background(),
				templateKey:        "dummy",
				language:           "de",
				languageSettingKey: "dummy",
				consentGuardKey:    "dummy",
				minConsentVersion:  "",
				payload:            map[string]interface{}{"field": "value"},
				subscribers:        []uuid.UUID{user1, user2, user3},
			},
			wantErr: false,
			wantStatus: NotificationStatus{
				StateProcessing: "not ready yet",
				StateQueue:      "not in queue",
				ConsentStats: map[string]int{
					EventConsent:          1,
					EventRevoke:           1,
					ConsentNeverConsented: 1,
					ConsentUnknown:        0,
					ConsentNotNeeded:      0,
				},
			},
		},
		{
			name: "No need to check consent",
			args: args{
				ctx:                context.Background(),
				templateKey:        "dummy",
				language:           "de",
				languageSettingKey: "dummy",
				consentGuardKey:    "",
				minConsentVersion:  "",
				payload:            map[string]interface{}{"field": "value"},
				subscribers:        []uuid.UUID{user1, user2, user3},
			},
			wantErr: false,
			wantStatus: NotificationStatus{
				StateProcessing: "not ready yet",
				StateQueue:      "not in queue",
				ConsentStats: map[string]int{
					EventConsent:          0,
					EventRevoke:           0,
					ConsentNeverConsented: 0,
					ConsentUnknown:        0,
					ConsentNotNeeded:      3,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := NewNotificationMock()
			genTraceID := uuid.Must(uuid.NewV4())
			tt.args.ctx = context.WithValue(tt.args.ctx, log.TraceIDContextKey, genTraceID)
			got, err := c.SendTemplated(tt.args.ctx, tt.args.templateKey, tt.args.language, tt.args.languageSettingKey, tt.args.consentGuardKey, tt.args.minConsentVersion, "", tt.args.payload, tt.args.subscribers...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// do not comapre jobIDs - they are automatically-generated
				assert.Equal(t, got.StateQueue, tt.wantStatus.StateQueue)
				assert.Equal(t, got.StateProcessing, tt.wantStatus.StateProcessing)
				assert.Equal(t, got.Result, tt.wantStatus.Result)
				assert.Equal(t, got.Error, tt.wantStatus.Error)
				assert.Equal(t, got.Caller, tt.wantStatus.Caller)
				assert.Equal(t, got.TraceID, tt.wantStatus.TraceID)
				if diff := deep.Equal(tt.wantStatus.ConsentStats, got.ConsentStats); diff != nil {
					t.Errorf("Arrays are not equal:\n%v", strings.Join(diff, "\n"))
					t.Logf("Want: %+v", tt.wantStatus.ConsentStats)
					t.Logf("Got: %+v", got.ConsentStats)
				}
			}
		})
	}
}
