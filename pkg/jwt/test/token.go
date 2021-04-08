package test

import (
	"crypto/rsa"
	"strings"
	"time"

	gcJWT "github.com/gesundheitscloud/go-svc/pkg/jwt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gofrs/uuid"
)

const (
	// Time diff tolerance.
	skew = time.Second

	testIssuer = "urn:ghc"

	testUserID   = "11111111-1111-1111-1111-111111111111"
	testClientID = "22222222-2222-2222-2222-222222222222#web"
	testAppID    = "33333333-3333-3333-3333-333333333333"
	testJWTID    = "44444444-4444-4444-4444-444444444444"
)

// GenerateToken creates a new JWT with the given claims and signed with the given private key.
// In case of errors, it panics.
func GenerateToken(
	privateKey *rsa.PrivateKey,
	tm time.Time,
	ownerID, appID, jwtID, userID uuid.UUID,
	clientID, issuer string,
	scope gcJWT.Scope,
) (string, error) {

	t := jwt.NewWithClaims(jwt.SigningMethodRS256, &gcJWT.Claims{
		Issuer:     issuer,
		Subject:    gcJWT.Owner{ID: ownerID},
		Expiration: gcJWT.Time(tm.Add(time.Minute)),
		NotBefore:  gcJWT.Time(tm.Add(-skew)),
		IssuedAt:   gcJWT.Time(tm),
		JWTID:      jwtID,
		AppID:      appID,
		ClientID:   clientID,
		UserID:     userID,
		Scope:      scope,
	})

	return t.SignedString(privateKey)
}

// BearerToken creates an Authorization Bearer header.
func BearerToken(
	key *rsa.PrivateKey,
	ownerID uuid.UUID,
	scope string,
) (string, error) {
	appUUID := uuid.Must(uuid.FromString(testAppID))
	jwtUUID := uuid.Must(uuid.FromString(testJWTID))
	userUUID := uuid.Must(uuid.FromString(testUserID))
	scp, err := gcJWT.NewScope(scope)
	if err != nil {
		return "", err
	}

	token, err := GenerateToken(
		key, time.Now(),
		ownerID, appUUID, jwtUUID, userUUID,
		testClientID, testIssuer,
		scp,
	)
	if err != nil {
		return "", err
	}

	var builder strings.Builder

	builder.WriteString("Bearer ")
	builder.WriteString(token)

	return builder.String(), nil
}
