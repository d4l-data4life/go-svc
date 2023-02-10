package client

import (
	"context"

	uuid "github.com/gofrs/uuid"
)

var _ Consent = (*ConsentMock)(nil)
var _ Consent = (*ConsentMockDetailed)(nil)
var _ Consent = (*ConsentMockRoundRobin)(nil)

type ConsentMock struct {
	state map[uuid.UUID]string
}

// NewConsentMock returns mocked consent-management service where the information about consents is stored in a map
// Note! this mock should return only the constants that are known to consent-management service, i.e.: EventConsent, EventRevoke, ConsentNeverConsented
func NewConsentMock(state map[uuid.UUID]string) *ConsentMock {
	return &ConsentMock{
		state: state,
	}
}

func (cs *ConsentMock) GetBatchConsents(ctx context.Context, consentKey string, minVersion string, subscribers ...uuid.UUID) (map[uuid.UUID]string, error) {
	result := make(map[uuid.UUID]string)
	for _, accID := range subscribers {
		value, ok := cs.state[accID]
		if !ok {
			result[accID] = ConsentNeverConsented
		} else {
			result[accID] = value
		}
	}
	return result, nil
}

type ConsentMockDetailed struct {
	err   error
	state map[uuid.UUID]map[string]string
}

// NewConsentMockDetailed returns mocked consent-management service where the information about consents is stored in a map per consent
func NewConsentMockDetailed(state map[uuid.UUID]map[string]string, err error) *ConsentMockDetailed {
	return &ConsentMockDetailed{
		state: state,
		err:   err,
	}
}

func (cm *ConsentMockDetailed) GetBatchConsents(ctx context.Context, consentKey string, minVersion string, subscribers ...uuid.UUID) (map[uuid.UUID]string, error) {
	if cm.err != nil {
		return nil, cm.err
	}
	result := make(map[uuid.UUID]string)
	for _, accID := range subscribers {
		consents, ok := cm.state[accID]
		if !ok {
			result[accID] = ConsentNeverConsented
		} else {
			value, ok := consents[consentKey]
			if !ok {
				result[accID] = ConsentNeverConsented
			} else {
				result[accID] = value
			}
		}
	}
	return result, nil
}

// ConsentMockRoundRobin assings consent events without looking at the user IDs
// first user always gets 'consent', second, 'revoke', third 'never-consented', fourth 'consent' and so on
type ConsentMockRoundRobin struct {
}

func NewConsentMockRoundRobin() *ConsentMockRoundRobin {
	return &ConsentMockRoundRobin{}
}

func (cs *ConsentMockRoundRobin) GetBatchConsents(ctx context.Context, consentKey string, minVersion string, subscribers ...uuid.UUID) (map[uuid.UUID]string, error) {
	result := make(map[uuid.UUID]string)
	for idx, accID := range subscribers {
		switch idx % 3 {
		case 0:
			result[accID] = EventConsent
		case 1:
			result[accID] = EventRevoke
		default:
			result[accID] = ConsentNeverConsented
		}
	}
	return result, nil
}
