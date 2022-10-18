package client

import (
	"context"
	"net/http"

	"golang.org/x/exp/maps"
)

var _ Listmonk = (*ListmonkMock)(nil)

// ListmonkMock is a mocked client for the user-preferences service
type ListmonkMock struct {
	counter map[string]map[int]int // email, templateId -> counter
}

func NewListmonkMock() *ListmonkMock {
	return &ListmonkMock{}
}

// Get fetches a single setting for a single user
func (lm *ListmonkMock) Tx(ctx context.Context, msg ListmonkTxMessage) (int, error) {
	if _, ok := lm.counter[msg.SubscriberEmail]; !ok {
		lm.counter[msg.SubscriberEmail] = make(map[int]int)
	}
	lm.counter[msg.SubscriberEmail][msg.TemplateID] = lm.counter[msg.SubscriberEmail][msg.TemplateID] + 1
	return http.StatusOK, nil
}

func (lm *ListmonkMock) GetNotifiedUsers() []string {
	return maps.Keys(lm.counter)
}
