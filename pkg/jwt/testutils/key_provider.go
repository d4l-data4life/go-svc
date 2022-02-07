package testutils

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/dynamic"
	"github.com/stretchr/testify/assert"
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

func GeneratePEMPublicKey(t *testing.T, pk *rsa.PublicKey, indent int) []byte {
	pKey, err := x509.MarshalPKIXPublicKey(pk)
	assert.NoError(t, err)
	pubkeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pKey,
		},
	)
	if indent > 0 {
		pubkeyPEM = bytes.ReplaceAll(pubkeyPEM, []byte("\n"), []byte("\n"+strings.Repeat(" ", indent))) // magic to match the yaml indent
	}
	return pubkeyPEM
}
