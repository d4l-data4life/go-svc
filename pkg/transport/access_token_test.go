package transport_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/d4lcontext"
	"github.com/gesundheitscloud/go-svc/pkg/transport"
)

func TestAccessToken(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		want    string
		wantErr bool
	}{
		{
			name: "success",
			ctx: context.WithValue(
				context.Background(),
				d4lcontext.AccessTokenContextKey,
				"some-access-token",
			),
			want:    "some-access-token",
			wantErr: false,
		},
		{
			name: "empty",
			ctx: context.WithValue(
				context.Background(),
				d4lcontext.AccessTokenContextKey,
				"",
			),
			want:    "",
			wantErr: false,
		},
		{
			name: "wrong type",
			ctx: context.WithValue(
				context.Background(),
				d4lcontext.AccessTokenContextKey,
				42,
			),
			want:    "",
			wantErr: true,
		},
		{
			name:    "nil",
			ctx:     context.Background(),
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got := r.Header.Get("Authorization")
				if got != tt.want {
					t.Fatalf("invalid access token %q, want %q", got, tt.want)
				}
			}))

			req, err := http.NewRequestWithContext(tt.ctx, http.MethodGet, server.URL, nil)
			if err != nil {
				t.Fatalf("error creating http request: %v", err)
			}

			cli := server.Client()
			cli.Transport = transport.AccessToken(cli.Transport)

			_, err = cli.Do(req)
			if err != nil && !tt.wantErr {
				t.Fatalf("round trip failed: %v", err)
			}
		})
	}
}
