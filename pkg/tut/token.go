package tut

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/jwt"

	"github.com/gofrs/uuid"
)

// GenerateAccessToken creates a new JWT with the given claims and signed with the given private key.
// In case of errors, it panics.
func GenerateAccessToken(userID uuid.UUID, privateKey *rsa.PrivateKey, claimsOptions ...jwt.TokenOption) string {
	options := []jwt.TokenOption{
		jwt.WithUserID(userID),
		jwt.WithAppID(uuid.Must(uuid.NewV4())),
		jwt.WithClientID(uuid.Must(uuid.NewV4()).String()),
		jwt.WithExpirationDuration(time.Minute),
	}
	options = append(options, claimsOptions...)

	t, err := jwt.CreateAccessToken(privateKey, options...)
	if err != nil {
		panic(fmt.Errorf("could not sign the token: %w", err))
	}

	return t.AccessToken
}

// MakeAuthHeader generates an access token with the given claims and signed with the given private key.
// It then returns an authorization header with the new JWT (format `Bearer <JWT>`)
// In case of errors, it panics.
func MakeAuthHeader(userID uuid.UUID, privateKey *rsa.PrivateKey, claimsOptions ...jwt.TokenOption) string {
	return fmt.Sprintf("Bearer %s", GenerateAccessToken(userID, privateKey, claimsOptions...))
}
