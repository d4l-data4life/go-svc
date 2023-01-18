package client

import (
	"context"

	uuid "github.com/gofrs/uuid"

	"github.com/gesundheitscloud/go-svc/pkg/log"
)

var _ Notification = (*NotificationMock)(nil)

// NotificationMock mimics notification service >= v0.6.0  - returns information about notified users
type NotificationMock struct {
	counter *notifiedUsersCounter
	upCli   UserPreferences
	csCli   Consent
}

// NewNotificationMock creates new NotificationMock using default mock-clients for UserPreferences and ConsentManagement
func NewNotificationMock() *NotificationMock {
	return NewNotificationMockWithClients(nil, nil)
}

// NewNotificationMockWithClients creates new NotificationMock while allowing to specify clients for UserPreferences and ConsentManagement
func NewNotificationMockWithClients(upCli UserPreferences, csCli Consent) *NotificationMock {
	if upCli == nil {
		upCli = NewUserPreferencesMock()
	}
	if csCli == nil {
		csCli = NewConsentMockRoundRobin()
	}
	return &NotificationMock{
		counter: newNotifiedUsersCounter(),
		upCli:   upCli,
		csCli:   csCli,
	}
}

func (c *NotificationMock) SendRaw(ctx context.Context,
	consentGuardKey string, minConsentVersion string,
	fromName string, fromAddress string, subject string, message string,
	arbitraryEmailAddress string,
	payload map[string]interface{},
	subscribers ...uuid.UUID,
) (NotificationStatus, error) {
	c.counter.Count("raw", "raw", subscribers...)
	caller, _ := payload["caller"].(string)
	traceID, _ := ctx.Value(log.TraceIDContextKey).(string)

	ns := NotificationStatus{
		JobIDs:          []uuid.UUID{uuid.Must(uuid.NewV4())},
		Error:           "",
		Result:          "",
		Caller:          caller,
		StateQueue:      "not in queue",
		StateProcessing: "not ready yet",
		TraceID:         traceID,
	}
	return ns, nil
}

func (c *NotificationMock) SendTemplated(ctx context.Context,
	templateKey, language, languageSettingKey string,
	consentGuardKey string,
	minConsentVersion string,
	arbitraryEmailAddress string,
	payload map[string]interface{},
	subscribers ...uuid.UUID,
) (NotificationStatus, error) {
	if len(language) == 0 {
		if len(languageSettingKey) == 0 {
			languageSettingKey = "global.language"
		}
		langSettings, err := c.upCli.GetKeySettings(ctx, languageSettingKey)
		if err != nil {
			return NotificationStatus{
				Error: "error calling user-preferences service",
			}, err
		}
		for _, subsID := range subscribers {
			if prefLang, ok := langSettings[subsID].(string); ok {
				c.counter.Count(templateKey, prefLang, subsID)
			} else {
				// simulate template-default-language which is always 'en' in mock
				c.counter.Count(templateKey, "en", subsID)
			}
		}
	} else {
		c.counter.Count(templateKey, language, subscribers...)
	}

	caller, _ := payload["caller"].(string)
	traceID, _ := ctx.Value(log.TraceIDContextKey).(string)

	ns := NotificationStatus{
		JobIDs:          []uuid.UUID{uuid.Must(uuid.NewV4())},
		Error:           "",
		Result:          "",
		Caller:          caller,
		StateQueue:      "not in queue",
		StateProcessing: "not ready yet",
		TraceID:         traceID,
	}
	return ns, nil
}

func (c *NotificationMock) GetNotifiedUsers() NotifiedUsers {
	return c.counter.GetStatus()
}

func (c *NotificationMock) GetJobStatus(ctx context.Context, jobID uuid.UUID) (NotificationStatus, error) {
	reply := NotificationStatus{
		JobIDs: []uuid.UUID{jobID},
		Result: "Mock-ok",
	}
	return reply, nil
}

func (c *NotificationMock) DeleteJob(ctx context.Context, jobID uuid.UUID) error {
	return nil
}
