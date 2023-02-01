package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	uuid "github.com/gofrs/uuid"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
)

// NotificationServiceRequest is a copy of `NotificationRequest` from `cds-notification`
type NotificationServiceRequest struct {
	AccountIDs            []uuid.UUID            `json:"accountIDs"`
	ArbitraryEmailAddress string                 `json:"arbitraryEmailAddress"`
	FromName              string                 `json:"fromName"`
	FromAddress           string                 `json:"fromAddress"`
	Subject               string                 `json:"subject"`
	Message               string                 `json:"message"`
	TemplateKey           string                 `json:"templateKey"`
	LanguageSettingKey    string                 `json:"languageSettingKey"`
	ConsentGuardKey       string                 `json:"consentGuardKey"`             // optional parameter - default = ""
	MinConsentVersion     string                 `json:"minConsentVersion,omitempty"` // optional parameter - default = "0"
	TemplateLanguage      string                 `json:"templateLanguage"`
	Caller                string                 `json:"caller"`
	TemplatePayload       map[string]interface{} `json:"templatePayload,omitempty"`
}

// NotificationStatus object is returned by 'SendTemplated' and 'GetJobStatus' to the caller
type NotificationStatus struct {
	JobIDs          []uuid.UUID `json:"jobIDs"`
	StateQueue      string      `json:"stateQueue"`
	StateProcessing string      `json:"stateProcessing"`
	Result          string      `json:"result"`
	Error           string      `json:"error"`
	Caller          string      `json:"caller"`
	TraceID         string      `json:"traceID"`
}

type Notification interface {
	// SendTemplated sends a templated email and returns error
	SendTemplated(ctx context.Context,
		templateKey, language, languageSettingKey string,
		consentGuardKey string, minConsentVersion string,
		arbitraryEmailAddress string,
		payload map[string]interface{}, subscribers ...uuid.UUID) (NotificationStatus, error)
	// SendRaw sends a plain-text email and returns error
	SendRaw(ctx context.Context,
		consentGuardKey string, minConsentVersion string,
		fromName string, fromAddress string, subject string, message string,
		arbitraryEmailAddress string,
		payload map[string]interface{}, subscribers ...uuid.UUID) (NotificationStatus, error)
	// GetJobStatus returns the status of a notification job submitted asynchronously before
	GetJobStatus(ctx context.Context, jobID uuid.UUID) (NotificationStatus, error)
	// DeleteJob cancels job processing
	DeleteJob(ctx context.Context, jobID uuid.UUID) error
	// GetNotifiedUsers returns basic info about notified users and error
	GetNotifiedUsers() NotifiedUsers
}

var _ Notification = (*NotificationService)(nil)
var userAgentNotification = "go-svc.client.NotificationService"

// NotificationService is a client for the cds-notification
// it implements Notification and NotificationV2 interfaces
type NotificationService struct {
	svcAddr   string
	svcSecret string
	caller    *caller
	// counter stores state to return the NotifiedUsers information - mainly used in tests
	counter *notifiedUsersCounter
}

func NewNotificationService(svcAddr, svcSecret, caller string) *NotificationService {
	if caller == "" {
		caller = "unknown"
	}
	return &NotificationService{
		svcAddr:   svcAddr,
		svcSecret: svcSecret,
		caller:    NewCaller(30*time.Second, caller),
		counter:   newNotifiedUsersCounter(),
	}
}

func (c *NotificationService) GetNotifiedUsers() NotifiedUsers {
	return c.counter.GetStatus()
}

func (c *NotificationService) SendRaw(ctx context.Context,
	consentGuardKey string, minConsentVersion string,
	fromName string, fromAddress string, subject string, message string,
	arbitraryEmailAddress string,
	payload map[string]interface{},
	subscribers ...uuid.UUID,
) (NotificationStatus, error) {
	requestBody := NotificationServiceRequest{
		AccountIDs:            subscribers,
		ArbitraryEmailAddress: arbitraryEmailAddress,
		ConsentGuardKey:       consentGuardKey,
		MinConsentVersion:     minConsentVersion,
		FromName:              fromName,
		FromAddress:           fromAddress,
		Subject:               subject,
		Message:               message,
		Caller:                c.caller.name,
		TemplatePayload:       payload,
	}
	c.counter.Count("raw", "raw", subscribers...)
	return c.sendRawEmail(ctx, requestBody)
}

func (c *NotificationService) SendTemplated(ctx context.Context,
	templateKey, language, languageSettingKey string,
	consentGuardKey string,
	minConsentVersion string,
	arbitraryEmailAddress string,
	payload map[string]interface{},
	subscribers ...uuid.UUID,
) (NotificationStatus, error) {
	requestBody := NotificationServiceRequest{
		AccountIDs:            subscribers,
		ArbitraryEmailAddress: arbitraryEmailAddress,
		TemplateKey:           templateKey,
		TemplateLanguage:      language,
		LanguageSettingKey:    languageSettingKey,
		ConsentGuardKey:       consentGuardKey,
		MinConsentVersion:     minConsentVersion,
		Caller:                c.caller.name,
		TemplatePayload:       payload,
	}
	// for calculation of notifiedUsersInfo
	c.counter.Count(templateKey, language, subscribers...)
	return c.sendTemplatedEmail(ctx, requestBody)
}

func (c *NotificationService) GetJobStatus(ctx context.Context, jobID uuid.UUID) (NotificationStatus, error) {
	contentURL := fmt.Sprintf("%s/api/v1/jobs/%s", c.svcAddr, jobID.String())
	reply := NotificationStatus{}
	respBody, _, _, err := c.caller.call(ctx, contentURL, "GET", c.svcSecret, userAgentNotification, &bytes.Buffer{}, http.StatusOK, http.StatusAccepted)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "fetching job status failed")
		return NotificationStatus{}, err
	}

	if err := json.Unmarshal(respBody, &reply); err != nil {
		logging.LogErrorfCtx(ctx, err, "error transforming job status service reply to an object")
		return reply, err
	}
	return reply, nil
}

func (c *NotificationService) DeleteJob(ctx context.Context, jobID uuid.UUID) error {
	contentURL := fmt.Sprintf("%s/api/v1/jobs/%s", c.svcAddr, jobID.String())
	_, _, _, err := c.caller.call(ctx, contentURL, "DELETE", c.svcSecret, userAgentNotification, &bytes.Buffer{}, http.StatusNoContent)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "canceling notification job failed")
	}
	return err
}

func (c *NotificationService) sendTemplatedEmail(ctx context.Context, requestBody NotificationServiceRequest) (NotificationStatus, error) {
	contentURL := fmt.Sprintf("%s/api/v1/notifications/template", c.svcAddr)
	reply := NotificationStatus{}
	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error transforming notification request to JSON")
		return reply, err
	}
	body, code, _, err := c.caller.call(ctx, contentURL, "POST", c.svcSecret, userAgentNotification, bytes.NewBuffer(jsonBytes), http.StatusOK, http.StatusAccepted)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "sendTemplatedEmail failed, code: %d", code)
		return reply, err
	}

	err = json.Unmarshal(body, &reply)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error unmarshalling reply from notification service. Reply: %s", string(body))
	}
	return reply, err
}

func (c *NotificationService) sendRawEmail(ctx context.Context, requestBody NotificationServiceRequest) (NotificationStatus, error) {
	contentURL := fmt.Sprintf("%s/api/v1/notifications/raw", c.svcAddr)
	reply := NotificationStatus{}
	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error transforming notification request to JSON")
		return reply, err
	}
	body, code, _, err := c.caller.call(ctx, contentURL, "POST", c.svcSecret, userAgentNotification, bytes.NewBuffer(jsonBytes), http.StatusOK, http.StatusAccepted)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "sendRawEmail failed, code: %d", code)
		return reply, err
	}

	err = json.Unmarshal(body, &reply)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error unmarshalling reply from notification service. Reply: %s", string(body))
	}
	return reply, err
}
