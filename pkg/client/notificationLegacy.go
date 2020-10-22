package client

import (
	"context"

	uuid "github.com/satori/go.uuid"
)

var _ Notification = (*NotificationServiceLegacy)(nil)
var _ NotificationV2 = (*NotificationServiceLegacy)(nil)

// NotificationServiceLegacy is a client for the cds-notification
// it implements Notification and NotificationV2 interfaces
type NotificationServiceLegacy struct {
	ns *NotificationService
}

func NewNotificationServiceLegacy(svcAddr, svcSecret, caller string) *NotificationServiceLegacy {
	return &NotificationServiceLegacy{NewNotificationService(svcAddr, svcSecret, caller)}
}

func (c *NotificationServiceLegacy) GetNotifiedUsers() NotifiedUsers {
	return c.ns.GetNotifiedUsers()
}

func (c *NotificationServiceLegacy) SendTemplated(templateKey, language string,
	payload map[string]interface{},
	subscribers ...uuid.UUID,
) error {
	return c.ns.SendTemplated(context.Background(), templateKey, language, "global.language", payload, subscribers...)
}
