package client

import (
	"context"
	"errors"
	"net/http"

	"github.com/gofrs/uuid"
)

var _ VegaInternal = (*VegaInternalMock)(nil)

type VegaInternalMock struct {
	statusCode   int
	emailTenants map[uuid.UUID]EmailTenant
}

func NewVegaInternalMock(emailTenants map[uuid.UUID]EmailTenant, statusCode int) *VegaInternalMock {
	return &VegaInternalMock{
		statusCode:   statusCode,
		emailTenants: emailTenants,
	}
}

// ResolveEmailV2 returns mapping of userID -> (email, tenantID). Uses tenantID returned from vega
func (v *VegaInternalMock) ResolveEmailV2(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]EmailTenant, int, error) {
	emailMap := make(map[uuid.UUID]EmailTenant)

	if v.statusCode != http.StatusOK {
		return emailMap, v.statusCode, errors.New("failed resolving emails")
	}

	for _, userID := range userIDs {
		if email, ok := v.emailTenants[userID]; ok {
			emailMap[userID] = email
		}
	}

	if len(emailMap) == 0 {
		return emailMap, http.StatusNotFound, errors.New("no users found")
	}

	return emailMap, http.StatusOK, nil
}
