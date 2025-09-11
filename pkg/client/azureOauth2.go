package client

import (
	"context"
	"errors"

	"golang.org/x/oauth2/clientcredentials"

	"github.com/d4l-data4life/go-svc/pkg/logging"
)

// define error messages to display to users
var (
	ErrALPAuthClientMisconfigured = errors.New("azure-oauth2-client: unable to work")
)

type ALPAuth interface {
	Authenticate(ctx context.Context) (string, error)
}

// API v2 client

type AzureOauth2 struct {
	tokenEndpoint string
	scope         string
	clientID      string
	clientSecret  string
}

func NewAzureOauth2(tokenEndpoint, scope, clientID, clientSecret string) ALPAuth {
	return &AzureOauth2{
		tokenEndpoint: tokenEndpoint,
		scope:         scope,
		clientID:      clientID,
		clientSecret:  clientSecret,
	}
}

func (ao *AzureOauth2) Authenticate(ctx context.Context) (string, error) {
	conf := &clientcredentials.Config{
		ClientID:     ao.clientID,
		ClientSecret: ao.clientSecret,
		Scopes:       []string{ao.scope},
		TokenURL:     ao.tokenEndpoint,
	}
	token, err := conf.Token(ctx)
	if err != nil {
		logging.LogErrorfCtx(ctx, err, "Azure oauth2 authentication error")
		return "", err
	}
	return token.AccessToken, nil
}
