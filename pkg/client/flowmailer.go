package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
)

// define errors
var (
	ErrFMMarshalError   = errors.New("cannot marshal Flowmailer submit message")
	ErrFMAuthentication = errors.New("failed authentication to Flowmailer API")
)

const (
	FMOAuthURL          = "https://login.flowmailer.net/oauth/token"
	FMBaseURL           = "https://api.flowmailer.net"
	userAgentFlowmailer = "go-svc.client.Flowmailer"
)

// FMSendEmail is the data format for an email request to our client
type FMSendEmail struct {
	TemplateID  int
	FromName    string
	FromAddress string
	Recipient   string
	Data        map[string]any
}

// FMMessageStatus returns the status of a flowmailer message (a sent email)
// Includes timestamps as unix timestamps
type FMMessageStatus struct {
	Status       string `json:"status,omitempty"`
	Submitted    int64  `json:"submitted,omitempty"`
	BackendStart int64  `json:"backendStart,omitempty"`
	BackendDone  int64  `json:"backendDone,omitempty"`
}

func (f FMMessageStatus) String() string {
	return fmt.Sprintf("Status: %s - Submitted %v - BackendStart %v - BackendDone %v",
		f.Status, time.UnixMilli(f.Submitted), time.UnixMilli(f.BackendStart), time.UnixMilli(f.BackendDone))
}

// FMSubmitMessage is the data format for an email request to the Flowmailer API
// MessageType is always EMAIL and the FlowSelector is derived from the templateID of the request
type FMSubmitMessage struct {
	MessageType      string         `json:"messageType,omitempty"`
	FlowSelector     string         `json:"flowSelector,omitempty"`
	RecipientAddress string         `json:"recipientAddress,omitempty"`
	SenderAddress    string         `json:"senderAddress,omitempty"`
	HeaderFromName   string         `json:"headerFromName,omitempty"`
	Data             map[string]any `json:"data,omitempty"`
}

type Flowmailer interface {
	// SendEmail submits a message via Flowmailer API
	SendEmail(ctx context.Context, msg FMSendEmail) (int, string, error)

	// GetMessageStatus requests the status of the sent email (processing, delivered, bounced, ...)
	GetMessageStatus(ctx context.Context, messageID string) (int, FMMessageStatus, error)
}

var _ Flowmailer = (*FlowmailerApi)(nil)

// FlowmailerApi is a client for Flowmailer
// it implements the Flowmailer API interface
type FlowmailerApi struct {
	accountID      string
	clientID       string
	clientSecret   string
	caller         *caller
	authHeader     string
	authExpiration time.Time
}

func NewFlowmailerApi(accountID, clientID, clientSecret, caller string) *FlowmailerApi {
	if caller == "" {
		caller = "unknown"
	}
	return &FlowmailerApi{
		accountID:    accountID,
		clientID:     clientID,
		clientSecret: clientSecret,
		caller:       NewCaller(30*time.Second, caller),
	}
}

// ensureAuthentication checks if the stored auth header is still valid and otherwise requests a new one
func (f *FlowmailerApi) ensureAuthentication(ctx context.Context) error {
	if time.Now().Before(f.authExpiration) {
		return nil
	}
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	data := url.Values{}
	data.Set("client_id", f.clientID)
	data.Set("client_secret", f.clientSecret)
	data.Set("grant_type", "client_credentials")
	encodedData := data.Encode()
	req, err := http.NewRequestWithContext(ctx, "POST", FMOAuthURL, strings.NewReader(encodedData))
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error creating HTTP request")
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Got error %s", err.Error())
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		logging.LogErrorfCtx(ctx, err, ErrFMAuthentication.Error())
		return ErrFMAuthentication
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error reading response body")
		return err
	}
	oauthResponse := struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}{}
	if err := json.Unmarshal(body, &oauthResponse); err != nil {
		logging.LogErrorfCtx(ctx, err, "error parsing OAuth response")
		return err
	}
	f.authHeader = fmt.Sprintf("Bearer %s", oauthResponse.AccessToken)
	f.authExpiration = time.Now().Add(time.Second * time.Duration(oauthResponse.ExpiresIn))
	return nil
}

// SendEmail uses Flowmailers message/submit API to send an email.
// Returns the status of the request and on success the ID of the created message.
func (f *FlowmailerApi) SendEmail(ctx context.Context, msg FMSendEmail) (int, string, error) {
	if err := f.ensureAuthentication(ctx); err != nil {
		logging.LogErrorfCtx(ctx, err, "error authenticating to Flowmailer API")
		return 0, "", err
	}

	// construct flowmailer payload
	payload := FMSubmitMessage{
		MessageType:      "EMAIL",
		FlowSelector:     fmt.Sprintf("template-%d", msg.TemplateID),
		RecipientAddress: msg.Recipient,
		SenderAddress:    msg.FromAddress,
		HeaderFromName:   msg.FromName,
		Data:             msg.Data,
	}

	// Send message
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, ErrFMMarshalError.Error())
		return 0, "", ErrFMMarshalError
	}
	reqURL := fmt.Sprintf("%s/%s/messages/submit", FMBaseURL, f.accountID)
	_, status, header, err := f.caller.call(ctx, reqURL, "POST", f.authHeader, userAgentFlowmailer, bytes.NewBuffer(jsonBytes), http.StatusCreated)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error submitting message to flowmailer")
		return status, "", err
	}

	// Get message id - FM returns only the location header with the url to the message so we extract it
	messageID := strings.Split(header.Get("location"), "messages/")[1]

	return status, messageID, nil
}

func (f *FlowmailerApi) GetMessageStatus(ctx context.Context, messageID string) (int, FMMessageStatus, error) {
	msgStatus := FMMessageStatus{}
	if err := f.ensureAuthentication(ctx); err != nil {
		logging.LogErrorfCtx(ctx, err, "error authenticating to Flowmailer API")
		return 0, msgStatus, err
	}

	reqURL := fmt.Sprintf("%s/%s/messages/%s", FMBaseURL, f.accountID, messageID)
	respBody, status, _, err := f.caller.call(ctx, reqURL, "GET", f.authHeader, userAgentFlowmailer, &bytes.Buffer{}, http.StatusOK)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error submitting message to flowmailer")
		return status, msgStatus, err
	}
	if err := json.Unmarshal(respBody, &msgStatus); err != nil {
		logging.LogErrorfCtx(ctx, err, "error parsing flowmailer message status from respone")
		return status, msgStatus, err
	}
	return status, msgStatus, nil
}
