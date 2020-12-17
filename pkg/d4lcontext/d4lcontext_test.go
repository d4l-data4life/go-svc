package d4lcontext

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetUserID(t *testing.T) {
	none := "none"
	id := "123e4567-e89b-12d3-a456-426655440000"
	tests := []struct {
		name   string
		userID string
		want   string
	}{
		{"Successful", id, id},
		{"Broken user id", "random", ""},
		{"No user id", none, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "", nil)
			if tt.userID != none {
				reqID, err := uuid.FromString(tt.userID)
				ctx := req.Context()
				if err == nil {
					ctx = context.WithValue(ctx, UserIDContextKey, reqID)
				} else {
					ctx = context.WithValue(ctx, UserIDContextKey, tt.userID)
				}
				req = req.WithContext(ctx)
			}
			if got := GetUserID(req); got != tt.want {
				t.Errorf("GetUserID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetClientID(t *testing.T) {
	none := "none"
	tests := []struct {
		name     string
		clientID string
		want     string
	}{
		{"Successful", "mobile123", "mobile123"},
		{"Empty client id", "", ""},
		{"No client id", none, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "", nil)
			if tt.clientID != none {
				ctx := context.WithValue(req.Context(), ClientIDContextKey, tt.clientID)
				req = req.WithContext(ctx)
			}
			if got := GetClientID(req); got != tt.want {
				t.Errorf("GetClientID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTenantID(t *testing.T) {
	none := "none"
	tests := []struct {
		name     string
		tenantID string
		want     string
	}{
		{"Successful", "charite", "charite"},
		{"Empty tenant id", "", "d4l"},
		{"No tenant id", none, "d4l"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "", nil)
			if tt.tenantID != none {
				ctx := context.WithValue(req.Context(), TenantIDContextKey, tt.tenantID)
				req = req.WithContext(ctx)
			}
			if got := GetTenantID(req); got != tt.want {
				t.Errorf("GetTenantID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRequesterID(t *testing.T) {
	id := "123e4567-e89b-12d3-a456-426655440000"
	reqID, _ := uuid.FromString(id)
	tests := []struct {
		name        string
		requesterID interface{}
		expectedID  uuid.UUID
		expectedErr string
	}{
		{"RequesterID as string", id, reqID, ""},
		{"RequesterID as uuid", reqID, reqID, ""},
		{"No requestid", nil, uuid.Nil, "missing account id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requesterID, err := ParseRequesterID(w, r)
				if err != nil {
					assert.Equal(t, tt.expectedErr, err.Error())
				}
				assert.Equal(t, tt.expectedID, requesterID)
			})

			req, _ := http.NewRequest(http.MethodGet, "/route", nil)
			ctx := context.WithValue(req.Context(), UserIDContextKey, tt.requesterID)
			req = req.WithContext(ctx)

			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)
		})
	}
}
