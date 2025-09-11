package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAzureOauth2_Authenticate(t *testing.T) {
	t.Skip("networked Azure test disabled in OSS; requires external tenant setup")
	scope := "https://example.onmicrosoft.com/app/.default"
	endpoint := "https://login.microsoftonline.com/00000000-0000-0000-0000-000000000000/oauth2/v2.0/token"
	clientID := ""
	invalidSecret := "test-secret"

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
