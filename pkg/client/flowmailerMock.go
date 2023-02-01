package client

import (
	"context"
	"net/http"
)

var _ Flowmailer = (*FlowmailerMock)(nil)

const (
	FMStatusDelivered = "DELIVERED"
	FMStatusProcessed = "PROCESSED"
	FMStatusBounced   = "HARDBOUNCE"
)

// FlowmailerMock is a mocked client for flowmailer API
type FlowmailerMock struct {
	counter map[string]map[int]int // email, templateId -> counter
}

func NewFlowmailerMock() *FlowmailerMock {
	return &FlowmailerMock{}
}

func (f *FlowmailerMock) SendEmail(ctx context.Context, msg FMSendEmail) (int, string, error) {
	if _, ok := f.counter[msg.Recipient]; !ok {
		f.counter[msg.Recipient] = make(map[int]int)
	}
	f.counter[msg.Recipient][msg.TemplateID] = f.counter[msg.Recipient][msg.TemplateID] + 1
	return http.StatusCreated, "", nil
}

// GetMessageStatus mock return Delivered status and no timestamps
func (f *FlowmailerMock) GetMessageStatus(ctx context.Context, messageID string) (int, FMMessageStatus, error) {
	return http.StatusOK, FMMessageStatus{Status: FMStatusDelivered}, nil
}
