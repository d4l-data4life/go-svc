package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
)

// define errors
var (
	ErrMarshalError  = errors.New("cannot marshal Tx request message")
	ErrResponseNotOK = errors.New("Listmonk reponded with an error")
)

const userAgentListmonk = "cds-notification.client.Listmonk"

type Headers []map[string]string

// ListmonkTxMessage represents an e-mail campaign.
type ListmonkTxMessage struct {
	SubscriberEmail string                 `json:"subscriber_email,omitempty"`
	SubscriberID    int                    `json:"subscriber_id,omitempty"`
	TemplateID      int                    `json:"template_id"`
	Data            map[string]interface{} `json:"data,omitempty"`
	FromName        string                 `json:"from_name,omitempty"`
	FromEmail       string                 `json:"from_email,omitempty"`
	Headers         Headers                `json:"headers,omitempty"`
	ContentType     string                 `json:"content_type,omitempty"`
	Messenger       string                 `json:"messenger,omitempty"`
}

type okRespBool struct {
	Data bool `json:"data"`
}

type Listmonk interface {
	// Send a transactional message to a subscriber using a predefined transactional template.
	Tx(ctx context.Context, msg ListmonkTxMessage) (int, error)

	// TxSync sends a synchronous (circumventing listmonk's internal queue and workers) transactional
	// message to a subscriber using a predefined transactional template.
	TxSync(ctx context.Context, msg ListmonkTxMessage) (int, error)
}

var _ Listmonk = (*ListmonkApi)(nil)

// ListmonkApi is a client for Listmonk
// it implements the Listmonk API interface
type ListmonkApi struct {
	apiURL    string
	apiSecret string
	caller    *caller
}

func NewListmonkApi(apiURL, apiUser, apiPassword, caller string) *ListmonkApi {
	if caller == "" {
		caller = "unknown"
	}
	userPass := fmt.Sprintf("%s:%s", apiUser, apiPassword)
	secret := base64.StdEncoding.EncodeToString([]byte(userPass))
	return &ListmonkApi{
		apiURL:    apiURL,
		apiSecret: fmt.Sprintf("Basic %s", secret),
		caller:    NewCaller(30*time.Second, caller),
	}
}

// sendTx sends a transactional message to a subscriber using a predefined transactional template.
func (c *ListmonkApi) sendTx(ctx context.Context, reqURL string, msg ListmonkTxMessage) (int, error) {
	if msg.FromName != "" {
		msg.FromEmail = fmt.Sprintf("%s <%s>", msg.FromName, msg.FromEmail)
		msg.FromName = ""
	}
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, ErrMarshalError.Error())
		return 0, ErrMarshalError
	}
	byteSettings, status, err := c.caller.call(ctx, reqURL, "POST", c.apiSecret, userAgentListmonk, bytes.NewBuffer(jsonBytes), http.StatusOK)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error sending a transactional message")
		return status, err
	}
	ok := okRespBool{}
	if err := json.Unmarshal(byteSettings, &ok); err != nil {
		logging.LogErrorfCtx(ctx, err, "error transforming Listmonk response to an object")
		return status, err
	}
	if !ok.Data {
		logging.LogErrorfCtx(ctx, ErrResponseNotOK, "")
		return status, ErrResponseNotOK
	}
	return status, nil
}

// Tx sends a transactional message to a subscriber using a predefined transactional template.
func (c *ListmonkApi) Tx(ctx context.Context, msg ListmonkTxMessage) (int, error) {
	reqURL := fmt.Sprintf("%s/api/tx", c.apiURL)
	return c.sendTx(ctx, reqURL, msg)
}

// TxSync sends a synchronous (circumventing listmonk's internal queue and workers) transactional
// message to a subscriber using a predefined transactional template.
func (c *ListmonkApi) TxSync(ctx context.Context, msg ListmonkTxMessage) (int, error) {
	reqURL := fmt.Sprintf("%s/api/custom/txsync", c.apiURL)
	return c.sendTx(ctx, reqURL, msg)
}
