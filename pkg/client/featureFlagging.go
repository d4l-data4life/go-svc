package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
)

type FeatureFlagging interface {
	// Get fetches a given feature flag
	Get(ctx context.Context, key string) (bool, error)
	GetForUser(ctx context.Context, key string, authorization string) (bool, error)
}

var _ FeatureFlagging = (*FeatureFlaggingService)(nil)
var userAgentFeatures = "go-svc.client.FeatureFlaggingService"

// FeatureFlaggingService is a client for the cds-notification
// it implements FeatureFlagging and FeatureFlaggingV2 interfaces
type FeatureFlaggingService struct {
	svcAddr   string
	svcSecret string
	caller    *Caller
}

func NewFeatureFlaggingService(svcAddr, svcSecret, caller string) *FeatureFlaggingService {
	if caller == "" {
		caller = "unknown"
	}
	return &FeatureFlaggingService{
		svcAddr:   svcAddr,
		svcSecret: svcSecret,
		caller:    NewCaller(30*time.Second, caller),
	}
}

// Get fetches a single setting
func (c *FeatureFlaggingService) Get(ctx context.Context, key string) (bool, error) {
	contentURL := fmt.Sprintf("%s/api/v1/services/%s", c.svcAddr, key)
	respBody, _, _, err := c.caller.call(ctx, contentURL, "GET", c.svcSecret, userAgentFeatures, &bytes.Buffer{}, http.StatusOK)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "fetching feature flag failed")
		return false, err
	}

	var boolValue bool
	err = json.Unmarshal(respBody, &boolValue)

	if err != nil {
		logging.LogErrorfCtx(ctx, err, "failed parsing response to boolean")
		return false, err
	}

	return boolValue, nil
}

// GetForUser fetches a single setting for a user
func (c *FeatureFlaggingService) GetForUser(ctx context.Context, key string, authorization string) (bool, error) {
	contentURL := fmt.Sprintf("%s/api/v1/features/%s", c.svcAddr, key)
	respBody, _, _, err := c.caller.call(ctx, contentURL, "GET", authorization, userAgentFeatures, &bytes.Buffer{}, http.StatusOK)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "fetching feature flag failed")
		return false, err
	}

	var boolValue bool
	err = json.Unmarshal(respBody, &boolValue)

	if err != nil {
		logging.LogErrorfCtx(ctx, err, "failed parsing response to boolean")
		return false, err
	}

	return boolValue, nil
}
