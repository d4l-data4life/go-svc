package testutils

import (
	"crypto/rsa"

	"github.com/gesundheitscloud/go-svc/pkg/dynamic"
)

// DummyKeyProvider will be used when the deprecated constructor `New` is used
// We want to prevent setting Authenticator.keyProvider to nil, so instead this struct will be used
type DummyKeyProvider struct {
	Key *rsa.PublicKey
}

func (dkp *DummyKeyProvider) JWTPublicKeys() ([]dynamic.JWTPublicKey, error) {
	jwtpk := dynamic.JWTPublicKey{Key: dkp.Key, Name: "arbitrary", Comment: "generated in code jwt.New()"}
	return []dynamic.JWTPublicKey{jwtpk}, nil
}
