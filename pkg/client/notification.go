package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	uuid "github.com/gofrs/uuid"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
	"github.com/gesundheitscloud/go-svc/pkg/middlewares"
)

// NotificationServiceRequest defines the payload (input) to be sent in 'SendTemplated'
type NotificationServiceRequest struct {
	AccountIDs                   []uuid.UUID            `json:"accountIDs"`
	TemplateKey                  string                 `json:"templateKey"`
	TemplateLanguage             string                 `json:"templateLanguage"`
	LanguageSettingKey           string                 `json:"languageSettingKey"`
	ConsentGuardKey              string                 `json:"consentGuardKey"`             // optional parameter - default = ""
	MinConsentVersion            int                    `json:"minConsentVersion,omitempty"` // optional parameter - default = "0"
	Caller                       string                 `json:"caller"`
	TraceID                      uuid.UUID              `json:"traceID"`
	UseMailJetTemplatingLanguage bool                   `json:"useMailJetTemplatingLanguage"`
	TemplatePayload              map[string]interface{} `json:"templatePayload"`
	TemplateErrorReportingEmail  string                 `json:"templateErrorReportingEmail"`
}

// NotificationStatus object is returned by 'SendTemplated' and 'GetJobStatus' to the caller
type NotificationStatus struct {
	JobIDs          []uuid.UUID    `json:"jobIDs"`
	StateQueue      string         `json:"stateQueue"`
	StateProcessing string         `json:"stateProcessing"`
	Result          string         `json:"result"`
	Error           string         `json:"error"`
	Caller          string         `json:"caller"`
	TraceID         string         `json:"traceID"`
	ConsentStats    map[string]int `json:"consentStats"`
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
	// GetNotifiedUsers returns basic info about notified users and error
	GetNotifiedUsers() NotifiedUsers
}

// NotificationV3 is an extension of Notification(V2) interface
// It changes a method signature and is not backwards-compatible with NotificationV2 and NotificationV1
// However, NotificationV2 and NotificationV1 remain compatible with cds-notification v0.6.x
type NotificationV3 interface {
	// SendTemplated sends a templated email and returns error
	SendTemplated(ctx context.Context,
		templateKey, language, languageSettingKey string, payload map[string]interface{}, subscribers ...uuid.UUID) error
	// GetNotifiedUsers returns basic info about notified users and error
	GetNotifiedUsers() NotifiedUsers
}

// NotificationV4 is an extension of Notification(V3) interface
// It changes the signature of 'SendTemplated' and is not backwards-compatible with previous 'Notification' interfaces
// However, NotificationV3, NotificationV2 and NotificationV1 remain compatible with cds-notification v0.6.x
type NotificationV4 interface {
	// SendTemplated sends a templated email and returns error
	SendTemplated(ctx context.Context,
		templateKey, language, languageSettingKey string,
		consentGuardKey string, minConsentVersion int,
		payload map[string]interface{}, subscribers ...uuid.UUID) (NotificationStatus, error)
	// GetJobStatus returns the status of a notification job submitted asynchronously before
	GetJobStatus(ctx context.Context, jobID uuid.UUID) (NotificationStatus, error)
	// DeleteJob cancels job processing
	DeleteJob(ctx context.Context, jobID uuid.UUID) error
	// GetNotifiedUsers returns basic info about notified users and error
	GetNotifiedUsers() NotifiedUsers
}

var _ NotificationV4 = (*NotificationService)(nil)
var userAgentNotification = "go-svc.client.NotificationService"

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
	consentGuardKey string,
	minConsentVersion int,
	payload map[string]interface{},
	subscribers ...uuid.UUID,
) (NotificationStatus, error) {
	traceID := uuid.Must(uuid.NewV4())
	requestBody := NotificationServiceRequest{
		AccountIDs:                   subscribers,
		TemplateKey:                  templateKey,
		TemplateLanguage:             language,
		LanguageSettingKey:           languageSettingKey,
		ConsentGuardKey:              consentGuardKey,
		MinConsentVersion:            minConsentVersion,
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

func (c *NotificationService) GetJobStatus(ctx context.Context, jobID uuid.UUID) (NotificationStatus, error) {
	contentURL := fmt.Sprintf("%s/api/v1/jobs/%s", c.svcAddr, jobID.String())
	reply := NotificationStatus{}
	byteSettings, _, err := call(ctx, contentURL, "GET", c.svcSecret, userAgentNotification, &bytes.Buffer{}, http.StatusOK, http.StatusAccepted)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "fetching job status failed")
		return NotificationStatus{}, err
	}
	if err := json.Unmarshal(byteSettings, &reply); err != nil {
		logging.LogErrorfCtx(ctx, err, "error transforming job status service reply to an object")
		return reply, err
	}
	return reply, nil
}

func (c *NotificationService) DeleteJob(ctx context.Context, jobID uuid.UUID) error {
	contentURL := fmt.Sprintf("%s/api/v1/jobs/%s", c.svcAddr, jobID.String())
	_, _, err := call(ctx, contentURL, "DELETE", c.svcSecret, userAgentNotification, &bytes.Buffer{}, http.StatusNoContent)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "canceling notification job failed")
	}
	return err
}

func (c *NotificationService) sendTemplatedEmail(ctx context.Context, requestBody NotificationServiceRequest) (NotificationStatus, error) {
	ns := NotificationStatus{}
	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error transforming notification request to JSON")
		return ns, err
	}

	contentURL := fmt.Sprintf("%s/api/v1/notifications/template", c.svcAddr)
	request, err := http.NewRequestWithContext(ctx, "POST", contentURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error creating HTTP request")
		return ns, err
	}
	request.Header.Add("Authorization", c.svcSecret)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "go-svc.client.NotificationService")
	request.Close = true

	client := &http.Client{Transport: &middlewares.TraceTransport{}}
	response, err := client.Do(request)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error sending request to notification service")
		return ns, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if response.StatusCode != http.StatusAccepted {
		if err == nil {
			err = fmt.Errorf("notification-svc error: %s", string(body))
		}
		logging.LogErrorfCtx(ctx, err, "error sending request to notification service. Status: %s", http.StatusText(response.StatusCode))
		return ns, err
	}

	err = json.Unmarshal(body, &ns)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error processing reply from notification service")
		return ns, err
	}
	return ns, nil
}
