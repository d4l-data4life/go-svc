package client

import (
	"context"
	"strconv"

	uuid "github.com/gofrs/uuid"
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
	_, err := c.ns.SendTemplated(context.Background(), templateKey, language, "global.language", "", "", "", payload, subscribers...)
	return err
}

var _ NotificationV3 = (*NotificationServiceLegacyV3)(nil)

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
	_, err := c.ns.SendTemplated(context.Background(), templateKey, language, languageSettingKey, "", "", "", payload, subscribers...)
	return err
}

var _ NotificationV4 = (*NotificationServiceLegacyV4)(nil)

// NotificationServiceLegacyV4 is a client for the cds-notification
// it implements NotificationV4 interface
type NotificationServiceLegacyV4 struct {
	ns *NotificationService
}

func NewNotificationServiceLegacyV4(svcAddr, svcSecret, caller string) *NotificationServiceLegacyV4 {
	return &NotificationServiceLegacyV4{NewNotificationService(svcAddr, svcSecret, caller)}
}

func (c *NotificationServiceLegacyV4) GetNotifiedUsers() NotifiedUsers {
	return c.ns.GetNotifiedUsers()
}

func (c *NotificationServiceLegacyV4) SendTemplated(ctx context.Context,
	templateKey, language, languageSettingKey string,
	consentGuardKey string, minConsentVersion int,
	payload map[string]interface{}, subscribers ...uuid.UUID,
) (NotificationStatus, error) {
	return c.ns.SendTemplated(context.Background(), templateKey, language, languageSettingKey, "", "", "", payload, subscribers...)
}

func (c *NotificationServiceLegacyV4) DeleteJob(ctx context.Context, jobID uuid.UUID) error {
	return c.ns.DeleteJob(ctx, jobID)
}

func (c *NotificationServiceLegacyV4) GetJobStatus(ctx context.Context, jobID uuid.UUID) (NotificationStatus, error) {
	return c.ns.GetJobStatus(ctx, jobID)
}

var _ NotificationV5 = (*NotificationServiceLegacyV5)(nil)

// NotificationServiceLegacyV5 is a client for the cds-notification
// it implements NotificationV5 interface
type NotificationServiceLegacyV5 struct {
	ns *NotificationService
}

func NewNotificationServiceLegacyV5(svcAddr, svcSecret, caller string) *NotificationServiceLegacyV5 {
	return &NotificationServiceLegacyV5{NewNotificationService(svcAddr, svcSecret, caller)}
}

func (c *NotificationServiceLegacyV5) GetNotifiedUsers() NotifiedUsers {
	return c.ns.GetNotifiedUsers()
}

func (c *NotificationServiceLegacyV5) SendTemplated(ctx context.Context,
	templateKey, language, languageSettingKey string,
	consentGuardKey string, minConsentVersion int, arbitraryEmailAddress string,
	payload map[string]interface{}, subscribers ...uuid.UUID,
) (NotificationStatus, error) {
	return c.ns.SendTemplated(context.Background(), templateKey, language, languageSettingKey, "", strconv.Itoa(minConsentVersion), "", payload, subscribers...)
}

func (c *NotificationServiceLegacyV5) DeleteJob(ctx context.Context, jobID uuid.UUID) error {
	return c.ns.DeleteJob(ctx, jobID)
}

func (c *NotificationServiceLegacyV5) GetJobStatus(ctx context.Context, jobID uuid.UUID) (NotificationStatus, error) {
	return c.ns.GetJobStatus(ctx, jobID)
}
