package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/tut"
)

func TestNotificationClient_SendTemplated(t *testing.T) {
	tests := []struct {
		name           string
		prepareSvcMock func(m *tut.ExternalServiceMock)
		wantErr        bool
		desiredErrMsg  string
	}{
		{
			name: "Can handle well-formed result",
			prepareSvcMock: func(m *tut.ExternalServiceMock) {
				m.On("/api/v1/notifications/template",
					tut.EndpointMock{
						RequestChecks: tut.CheckRequest(
							tut.ReqHasHTTPMethod(http.MethodPost),
						),
						ResponseBuilder: tut.BuildResponse(
							tut.RespWithJSONBody(NotificationStatus{
								Result: "Success",
							}),
						),
					},
				)
			},
			wantErr: false,
		},
		{
			name: "Can handle errors",
			prepareSvcMock: func(m *tut.ExternalServiceMock) {
				m.On("/api/v1/notifications/template",
					tut.EndpointMock{
						RequestChecks: tut.CheckRequest(
							tut.ReqHasHTTPMethod(http.MethodPost),
						),
						ResponseBuilder: tut.BuildResponse(
							tut.RespWithStatus(http.StatusServiceUnavailable),
						),
					},
				)
			},
			wantErr:       true,
			desiredErrMsg: "unexpected return code 503 (wanted one of: [200,202])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tut.NewExternalService()
			tt.prepareSvcMock(&mock)
			notificationSrv := httptest.NewServer(mock.Handler())

			svc := NewNotificationService(notificationSrv.URL, "secret", "testClient")
			_, err := svc.SendTemplated(context.Background(), "some-template", "some-language", "", "", "", "test@test.test", nil)

			if (err != nil) != tt.wantErr {
				t.Fatalf("WantErr: %t -- HasErr: %t", tt.wantErr, (err != nil))
			}

			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.desiredErrMsg) {
					t.Fatalf("Expected error: %v, got error: %v", tt.desiredErrMsg, err)
				}
			}
		})
	}
}

func TestNotificationClient_SendRaw(t *testing.T) {
	tests := []struct {
		name           string
		prepareSvcMock func(m *tut.ExternalServiceMock)
		wantErr        bool
		desiredErrMsg  string
	}{
		{
			name: "Can handle well-formed result",
			prepareSvcMock: func(m *tut.ExternalServiceMock) {
				m.On("/api/v1/notifications/raw",
					tut.EndpointMock{
						RequestChecks: tut.CheckRequest(
							tut.ReqHasHTTPMethod(http.MethodPost),
						),
						ResponseBuilder: tut.BuildResponse(
							tut.RespWithJSONBody(NotificationStatus{
								Result: "Success",
							}),
						),
					},
				)
			},
			wantErr: false,
		},
		{
			name: "Can handle errors",
			prepareSvcMock: func(m *tut.ExternalServiceMock) {
				m.On("/api/v1/notifications/raw",
					tut.EndpointMock{
						RequestChecks: tut.CheckRequest(
							tut.ReqHasHTTPMethod(http.MethodPost),
						),
						ResponseBuilder: tut.BuildResponse(
							tut.RespWithStatus(http.StatusServiceUnavailable),
						),
					},
				)
			},
			wantErr:       true,
			desiredErrMsg: "unexpected return code 503 (wanted one of: [200,202])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tut.NewExternalService()
			tt.prepareSvcMock(&mock)
			notificationSrv := httptest.NewServer(mock.Handler())

			svc := NewNotificationService(notificationSrv.URL, "secret", "testClient")
			_, err := svc.SendRaw(context.Background(), "", "", "", "", "", "", "", nil)

			if (err != nil) != tt.wantErr {
				t.Fatalf("WantErr: %t -- HasErr: %t", tt.wantErr, (err != nil))
			}

			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.desiredErrMsg) {
					t.Fatalf("Expected error: %v, got error: %v", tt.desiredErrMsg, err)
				}
			}
		})
	}
}
