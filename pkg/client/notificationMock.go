package client

import (
	uuid "github.com/satori/go.uuid"
)

var _ Notification = (*NotificationMock)(nil)
var _ NotificationV2 = (*NotificationMock)(nil)

// NotificationMock mimics notification service - returns information about notififed users
type NotificationMock struct {
	counter *notifiedUsersCounter
}

func NewNotificationMock() *NotificationMock {
	return &NotificationMock{
		counter: newNotifiedUsersCounter(),
	}
}

func (c *NotificationMock) SendTemplated(templateKey, language string, payload map[string]interface{}, subscribers ...uuid.UUID) error {
	c.counter.Count(templateKey, language, subscribers...)
	return nil
}

func (c *NotificationMock) GetNotifiedUsers() NotifiedUsers {
	return c.counter.GetStatus()
}
