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
	_, err := c.ns.SendTemplated(context.Background(), templateKey, language, "global.language", "", 0, payload, subscribers...)
	return err
}

// NotificationServiceLegacyV3 is a client for the cds-notification
// it implements NotificationV3 interface
type NotificationServiceLegacyV3 struct {
	ns *NotificationService
}

func NewNotificationServiceLegacyV3(svcAddr, svcSecret, caller string) *NotificationServiceLegacyV3 {
	return &NotificationServiceLegacyV3{NewNotificationService(svcAddr, svcSecret, caller)}
}

func (c *NotificationServiceLegacyV3) GetNotifiedUsers() NotifiedUsers {
	return c.ns.GetNotifiedUsers()
}

func (c *NotificationServiceLegacyV3) SendTemplated(ctx context.Context,
	templateKey, language, languageSettingKey string,
	payload map[string]interface{},
	subscribers ...uuid.UUID,
) error {
	_, err := c.ns.SendTemplated(context.Background(), templateKey, language, languageSettingKey, "", 0, payload, subscribers...)
	return err
}
