package d4lcontext

import (
	"context"
	"net/http"
	"testing"

	uuid "github.com/gofrs/uuid"
)

func TestGetUserID(t *testing.T) {
	someID := uuid.Must(uuid.NewV4())

	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "works with UUID",
			ctx:  context.WithValue(context.TODO(), UserIDContextKey, someID),
			want: someID.String(),
		},
		{
			name: "works with string",
			ctx:  context.WithValue(context.TODO(), UserIDContextKey, someID.String()),
			want: someID.String(),
		},
		{
			name: "works with empty user ID",
			ctx:  context.WithValue(context.TODO(), UserIDContextKey, ""),
			want: "",
		},
		{
			name: "works with missing user ID",
			ctx:  context.TODO(),
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "", nil)
			if got := GetUserID(req.WithContext(tt.ctx)); got != tt.want {
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

func TestGetTennatID(t *testing.T) {
	someID := uuid.Must(uuid.NewV4())

	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "works with an existing value",
			ctx:  context.WithValue(context.TODO(), TenantIDContextKey, someID.String()),
			want: someID.String(),
		},
		{
			name: "works with empty value",
			ctx:  context.WithValue(context.TODO(), TenantIDContextKey, ""),
			want: "",
		},
		{
			name: "works with missing key",
			ctx:  context.TODO(),
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "", nil)
			if got := GetTenantID(req.WithContext(tt.ctx)); got != tt.want {
				t.Errorf("GetTenantID() = %v, want %v", got, tt.want)
			}
		})
	}
}
