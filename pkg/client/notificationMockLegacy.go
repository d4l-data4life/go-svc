package client

import (
	"context"

	uuid "github.com/gofrs/uuid"
)

var _ NotificationV4 = (*NotificationMockLegacyV4)(nil)
var _ NotificationV3 = (*NotificationMockLegacyV3)(nil)
var _ NotificationV2 = (*NotificationMockLegacy)(nil)
var _ Notification = (*NotificationMockLegacy)(nil)

// NotificationMockLegacyV4 mimics notification service < v0.6.0
// returns information about notified users, accpets languageSettingKey
type NotificationMockLegacyV4 struct {
	counter *notifiedUsersCounter
}

func NewNotificationMockLegacyV4() *NotificationMockLegacyV4 {
	return &NotificationMockLegacyV4{
		counter: newNotifiedUsersCounter(),
	}
}

func (c *NotificationMockLegacyV4) SendTemplated(ctx context.Context,
	templateKey, language, languageSettingKey string,
	consentGuardKey string,
	minConsentVersion int,
	payload map[string]interface{},
	subscribers ...uuid.UUID,
) (NotificationStatus, error) {
	c.counter.Count(templateKey, language, subscribers...)
	return NotificationStatus{
		JobIDs: []uuid.UUID{uuid.Must(uuid.NewV4())},
		Result: "Mock-ok",
	}, nil
}

func (c *NotificationMockLegacyV4) GetNotifiedUsers() NotifiedUsers {
	return c.counter.GetStatus()
}

func (c *NotificationMockLegacyV4) GetJobStatus(ctx context.Context, jobID uuid.UUID) (NotificationStatus, error) {
	reply := NotificationStatus{
		JobIDs: []uuid.UUID{jobID},
		Result: "Mock-ok",
	}
	return reply, nil
}

func (c *NotificationMockLegacyV4) DeleteJob(ctx context.Context, jobID uuid.UUID) error {
	return nil
}

// NotificationMockLegacyV3 mimics notification service < v0.6.0
// returns information about notified users, accpets languageSettingKey
type NotificationMockLegacyV3 struct {
	counter *notifiedUsersCounter
}

func NewNotificationMockLegacyV3() *NotificationMockLegacyV3 {
	return &NotificationMockLegacyV3{
		counter: newNotifiedUsersCounter(),
	}
}

func (c *NotificationMockLegacyV3) SendTemplated(ctx context.Context,
	templateKey, language, languageSettingKey string, payload map[string]interface{}, subscribers ...uuid.UUID) error {
	c.counter.Count(templateKey, language, subscribers...)
	return nil
}

func (c *NotificationMockLegacyV3) GetNotifiedUsers() NotifiedUsers {
	return c.counter.GetStatus()
}

// NotificationMockLegacy mimics notification service < v0.6.0
// returns information about notified users
type NotificationMockLegacy struct {
	counter *notifiedUsersCounter
}

func NewNotificationMockLegacy() *NotificationMockLegacy {
	return &NotificationMockLegacy{
		counter: newNotifiedUsersCounter(),
	}
}

func (c *NotificationMockLegacy) SendTemplated(templateKey, language string, payload map[string]interface{}, subscribers ...uuid.UUID) error {
	c.counter.Count(templateKey, language, subscribers...)
	return nil
}

func (c *NotificationMockLegacy) GetNotifiedUsers() NotifiedUsers {
	return c.counter.GetStatus()
}
