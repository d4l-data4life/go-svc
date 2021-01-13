package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
	uuid "github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

// Names of consent events as defined in models pkg of consent-management
const (
	// EventConsent is returned when user gave consent
	EventConsent = "consent"
	// EventRevoke is returned when user revoked this consent
	EventRevoke = "revoke"
	// ConsentNeverConsented is means that these users has never given a consent and never revoked it
	ConsentNeverConsented = "never-consented"
)

// Names of consent events as used in cds-notification service
// to mark two extra states that do not exist in consent-management service
const (
	// ConsentUnknown covers any other event type returned
	// (e.g., when consent-management returns 404 or there is a new event type added in consent-management, but not here)
	ConsentUnknown = "consent-unknown"
	// ConsentNotNeeded is set when a given action requires no consent to be checked
	// this is never returned from the consent-service, but used to mark that there is no need to call consent-service
	ConsentNotNeeded = "consent-not-needed"
)

var (
	// ErrConsentNotFound wraps 404 error returned from consent-management client
	ErrConsentNotFound = errors.New("consent key not found")
)

type Consent interface {
	// GetBatchConsents fetches state of user consents identified by consent key
	GetBatchConsents(ctx context.Context, consentKey string, minVersion int, subscribers ...uuid.UUID) (map[uuid.UUID]string, error)
}

var _ Consent = (*ConsentService)(nil)
var userAgentConsent = "go-svc.client.ConsentService"

// ConsentService is a client for the cds-Consent
// it implements Consent and ConsentV2 interfaces
type ConsentService struct {
	svcAddr   string
	svcSecret string
	caller    string
}

func NewConsentService(svcAddr, svcSecret, caller string) *ConsentService {
	if caller == "" {
		caller = "unknown"
	}
	return &ConsentService{
		svcAddr:   svcAddr,
		svcSecret: "Bearer " + svcSecret, // Service still requires 'Bearer ' prefix for service-secret auth
		caller:    caller,
	}
}

func (cs *ConsentService) GetBatchConsents(ctx context.Context, consentKey string, minVersion int, subscribers ...uuid.UUID) (map[uuid.UUID]string, error) {
	var contentURL string
	if minVersion > 0 {
		contentURL = fmt.Sprintf("%s/api/v1/admin/userConsents/%s/batch?version=%d", cs.svcAddr, consentKey, minVersion)
	} else {
		contentURL = fmt.Sprintf("%s/api/v1/admin/userConsents/%s/batch", cs.svcAddr, consentKey)
	}
	result := make(map[uuid.UUID]string)
	payload, err := json.Marshal(subscribers)
	if err != nil {
		return result, err
	}
	body, gotStatus, err := call(ctx, contentURL, "POST", cs.svcSecret, userAgentConsent, bytes.NewBuffer(payload), http.StatusOK)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "error calling GetBatchConsents")
		if gotStatus == 404 {
			return result, ErrConsentNotFound
		}
		return result, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		logging.LogErrorfCtx(ctx, err, "error unmarshalling reply body: '%s'", string(body))
		return result, err
	}
	return result, err
}
