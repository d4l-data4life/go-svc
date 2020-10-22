package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	uuid "github.com/satori/go.uuid"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
)

// NotificationServiceRequest defines the payload of the actual send request
type NotificationServiceRequest struct {
	AccountIDs                   []uuid.UUID            `json:"accountIDs"`
	TemplateKey                  string                 `json:"templateKey"`
	TemplateLanguage             string                 `json:"templateLanguage"`
	LanguageSettingKey           string                 `json:"languageSettingKey"`
	Caller                       string                 `json:"caller"`
	TraceID                      uuid.UUID              `json:"traceID"`
	UseMailJetTemplatingLanguage bool                   `json:"useMailJetTemplatingLanguage"`
	TemplatePayload              map[string]interface{} `json:"templatePayload"`
	TemplateErrorReportingEmail  string                 `json:"templateErrorReportingEmail"`
}

type Notification interface {
	// SendTemplated sends a templated email and returns error
	SendTemplated(templateKey, language string, payload map[string]interface{}, subscribers ...uuid.UUID) error
}

// NotificationV2 is an extension of Notification(V1) interface
// It adds new method(s) and is compatible with NotificationV1
type NotificationV2 interface {
	// SendTemplated sends a templated email and returns error
	SendTemplated(templateKey, language string, payload map[string]interface{}, subscribers ...uuid.UUID) error
	// GetNotifiedUsers returns basic info about notififed users and error
	GetNotifiedUsers() NotifiedUsers
}

// NotificationV3 is an extension of Notification(V2) interface
// It changes a method signature and is not backwards-compatible with NotificationV2 and NotificationV1
// However, NotificationV2 and NotificationV1 remain compatible with cds-notification v0.6.x
type NotificationV3 interface {
	// SendTemplated sends a templated email and returns error
	SendTemplated(ctx context.Context,
		templateKey, language, languageSettingKey string, payload map[string]interface{}, subscribers ...uuid.UUID) error
	// GetNotifiedUsers returns basic info about notififed users and error
	GetNotifiedUsers() NotifiedUsers
}

var _ NotificationV3 = (*NotificationService)(nil)

// NotificationService is a client for the cds-notification
// it implements Notification and NotificationV2 interfaces
type NotificationService struct {
	svcAddr   string
	svcSecret string
	caller    string
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
		caller:    caller,
		counter:   newNotifiedUsersCounter(),
	}
}

func (c *NotificationService) GetNotifiedUsers() NotifiedUsers {
	return c.counter.GetStatus()
}

func (c *NotificationService) SendTemplated(ctx context.Context,
	templateKey, language, languageSettingKey string,
	payload map[string]interface{},
	subscribers ...uuid.UUID,
) error {
	traceID := uuid.NewV4()
	requestBody := NotificationServiceRequest{
		AccountIDs:                   subscribers,
		TemplateKey:                  templateKey,
		TemplateLanguage:             language,
		LanguageSettingKey:           languageSettingKey,
		Caller:                       c.caller,
		TraceID:                      traceID,
		UseMailJetTemplatingLanguage: payload != nil,
		TemplatePayload:              payload,
		TemplateErrorReportingEmail:  "",
	}
	// for calculation of notifiedUsersInfo
	c.counter.Count(templateKey, language, subscribers...)
	return c.sendTemplatedEmail(ctx, requestBody)
}

func (c *NotificationService) sendTemplatedEmail(ctx context.Context, requestBody NotificationServiceRequest) error {
	tracePrefix := "traceID = " + requestBody.TraceID.String() + ": "
	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error transforming notification request to JSON")
		return err
	}

	contentURL := fmt.Sprintf("%s/api/v1/notifications/template", c.svcAddr)
	request, err := http.NewRequestWithContext(ctx, "POST", contentURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error creating HTTP request")
		return err
	}
	request.Header.Add("Authorization", c.svcSecret)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "go-svc.client.NotificationService")
	request.Close = true

	client := &http.Client{}
	response, err := client.Do(request)
	if response != nil {
		defer response.Body.Close()
	}

	if err != nil {
		logging.LogErrorf(err, "%s error sending request to notification service", tracePrefix)
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	if response.StatusCode != http.StatusAccepted {
		if err == nil {
			err = fmt.Errorf(tracePrefix+"notification-svc error: %s", string(body))
		}
		logging.LogErrorf(err, "error sending request to notification service. Status: %s", http.StatusText(response.StatusCode))
		return err
	}
	return nil
}
