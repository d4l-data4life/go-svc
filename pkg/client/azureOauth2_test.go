package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAzureOauth2_Authenticate(t *testing.T) {
	scope := "https://alpsgdev.onmicrosoft.com/alp-portal/.default"
	endpoint := "https://login.microsoftonline.com/f59c7915-9283-4fb5-b633-435b9145ca00/oauth2/v2.0/token"
	clientID := ""
	invalidSecret := "super-secret"

	type fields struct {
		endpoint     string
		scope        string
		clientID     string
		clientSecret string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"fake secret",
			fields{endpoint, scope, clientID, invalidSecret}, true},
		// unable to test more without mock
	}
	for _, tt := range tests {
		ctx := context.Background()
		t.Run(tt.name, func(t *testing.T) {
			ao := &AzureOauth2{
				tokenEndpoint: tt.fields.endpoint,
				scope:         tt.fields.scope,
				clientID:      tt.fields.clientID,
				clientSecret:  tt.fields.clientSecret,
			}
			got, err := ao.Authenticate(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, got)
		})
	}
}
