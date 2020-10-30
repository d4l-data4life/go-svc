package client

import (
	"context"

	uuid "github.com/satori/go.uuid"
)

var _ NotificationV4 = (*NotificationMock)(nil)

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

func (c *NotificationMock) SendTemplated(ctx context.Context,
	templateKey, language, languageSettingKey string,
	consentGuardKey string,
	minConsentVersion int,
	payload map[string]interface{},
	subscribers ...uuid.UUID,
) (NotificationStatus, error) {
	if len(language) == 0 {
		if len(languageSettingKey) == 0 {
			languageSettingKey = "global.language"
		}
		settings, err := c.upCli.GetKeySettings(ctx, languageSettingKey)
		if err != nil {
			return NotificationStatus{
				Error: "error calling user-preferences service",
			}, err
		}
		for accID, lang := range settings {
			c.counter.Count(templateKey, lang, accID)
		}
	} else {
		c.counter.Count(templateKey, language, subscribers...)
	}

	caller, _ := payload["caller"].(string)
	traceID, _ := ctx.Value(TraceIDContextKey).(string)

	userConsents, _ := c.csCli.GetBatchConsents(ctx, consentGuardKey, minConsentVersion, subscribers...)
	stats := make(map[string]int)
	stats[EventConsent] = 0
	stats[EventRevoke] = 0
	stats[ConsentNeverConsented] = 0
	stats[ConsentUnknown] = 0
	stats[ConsentNotNeeded] = 0
	if len(consentGuardKey) == 0 {
		stats[ConsentNotNeeded] = len(subscribers)
	} else {
		for _, accID := range subscribers {
			event, ok := userConsents[accID]
			if ok {
				stats[event]++
			} else {
				stats[ConsentUnknown]++
			}
		}
	}
	ns := NotificationStatus{
		JobIDs:          []uuid.UUID{uuid.NewV4()},
		Error:           "",
		Result:          "",
		Caller:          caller,
		StateQueue:      "not in queue",
		StateProcessing: "not ready yet",
		TraceID:         traceID,
		ConsentStats:    stats,
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
