package client

import (
	uuid "github.com/satori/go.uuid"
)

var _ Notification = (*NotificationMock)(nil)

// NotificationMock mimics notification service - implements Notification interface
type NotificationMock struct{}

func NewNotificationMock() *NotificationMock {
	return &NotificationMock{}
}

func (c *NotificationMock) SendTemplated(templateKey, language string, payload map[string]interface{}, subscribers ...uuid.UUID) error {
	return nil
}
