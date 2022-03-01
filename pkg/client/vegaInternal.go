package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/logging"

	"github.com/gofrs/uuid"
)

// EmailTenant is a pair of named strings
type EmailTenant struct {
	Email    string `json:"email"`
	TenantID string `json:"tenantID"`
}

type VegaInternal interface {
	// ResolveEmailV2 resolves userID into pair: (email address, tenant ID)
	ResolveEmailV2(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]EmailTenant, int, error)
}

var _ VegaInternal = (*VegaInternalAPI)(nil)
var userAgentVega = "cds-notification.client.VegaInternalAPI"

type VegaInternalAPI struct {
	svcAddr   string
	svcSecret string
	caller    *caller
}

func NewVegaInternalAPI(svcAddr, svcSecret, caller string) *VegaInternalAPI {
	if caller == "" {
		caller = "unknown"
	}
	return &VegaInternalAPI{
		svcAddr:   svcAddr,
		svcSecret: "Bearer " + svcSecret,
		caller:    NewCaller(30*time.Second, caller),
	}
}

// ResolveEmailV2 returns mapping of userID -> (email, tenantID). Uses tenantID returned from vega
func (c *VegaInternalAPI) ResolveEmailV2(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]EmailTenant, int, error) {
	contentURL := fmt.Sprintf("%s/internal/users/api/v2/email", c.svcAddr)
	emailMap := make(map[uuid.UUID]EmailTenant)
	bodyBytes, err := json.Marshal(userIDs)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "marshalling user IDs failed")
		return emailMap, 0, err
	}
	body := bytes.NewBuffer(bodyBytes)
	// additionally accept 404s as null-results
	bReply, statusCode, err := c.caller.call(ctx, contentURL, "POST", c.svcSecret, userAgentVega, body, http.StatusOK, http.StatusNotFound)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "resolving user email over APIv2 failed")
		return emailMap, statusCode, err
	}
	// no need to unmarshal on 404s
	if statusCode == 404 {
		// we expect any error on 404s
		return emailMap, statusCode, fmt.Errorf("server replied with 404 and body: %s", string(bReply))
	}
	if err := json.Unmarshal(bReply, &emailMap); err != nil {
		logging.LogErrorfCtx(ctx, err, "failed unmarshalling vega reply into EmailTenant. Reply body: %s", string(bReply))
		return emailMap, statusCode, err
	}
	logging.LogDebugfCtx(ctx, "Vega '%s' returned emailMap: %+v", contentURL, emailMap)
	return emailMap, statusCode, nil
}
