package jwt

import (
	"crypto/rsa"
	"time"

	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/gofrs/uuid"
)

type Token struct {
	AccessToken string
	TokenType   string
}

type createOptions struct {
	expiration     time.Duration
	expirationTime time.Time
	tenantID       string
	userID         uuid.UUID
	appID          uuid.UUID
	clientID       string
	scopes         Scope
	email          string
}

type TokenOption func(opt *createOptions)

func WithUserID(userID uuid.UUID) TokenOption {
	return func(opt *createOptions) {
		opt.userID = userID
	}
}
func WithAppID(appID uuid.UUID) TokenOption {
	return func(opt *createOptions) {
		opt.appID = appID
	}
}
func WithClientID(clientID string) TokenOption {
	return func(opt *createOptions) {
		opt.clientID = clientID
	}
}
func WithTenantID(tenantID string) TokenOption {
	return func(opt *createOptions) {
		opt.tenantID = tenantID
	}
}
func WithExpirationDuration(expiration time.Duration) TokenOption {
	return func(opt *createOptions) {
		opt.expiration = expiration
	}
}
func WithExpirationTime(expiration time.Time) TokenOption {
	return func(opt *createOptions) {
		opt.expirationTime = expiration
	}
}
func WithScopeStrings(scopes ...string) TokenOption {
	return func(opt *createOptions) {
		opt.scopes = Scope{Tokens: scopes}
	}
}
func WithScope(s Scope) TokenOption {
	return func(opt *createOptions) {
		opt.scopes = s
	}
}
func WithEMail(email string) TokenOption {
	return func(opt *createOptions) {
		opt.email = email
	}
}

func CreateAccessToken(privateKey *rsa.PrivateKey, opt ...TokenOption) (Token, error) {
	options := &createOptions{}
	for _, f := range opt {
		f(options)
	}

	// jwt ID
	jwtID, err := uuid.NewV4()
	if err != nil {
		return Token{}, err
	}

	// Adjustment for time misalignments between servers.
	// This value is set to the maximum expected difference between production servers.
	const skew = time.Minute
	now := time.Now()

	claims := &Claims{
		Issuer:    IssuerGesundheitscloud,
		JWTID:     jwtID,
		Subject:   Owner{ID: options.userID},
		UserID:    options.userID,
		IssuedAt:  Time(now),
		NotBefore: Time(now.Add(-skew)),
		TenantID:  options.tenantID,
		AppID:     options.appID,
		ClientID:  options.clientID,
		Email:     options.email,
		Scope:     options.scopes,
	}

	// expiration
	var nilTime = time.Time{}
	if options.expiration > 0 {
		claims.Expiration = Time(now.Add(options.expiration).Add(skew))
	} else if options.expirationTime != nilTime {
		claims.Expiration = Time(options.expirationTime.Add(skew))
	}

	// create and sign token
	t := jwtgo.NewWithClaims(jwtgo.SigningMethodRS256, claims)
	at, err := t.SignedString(privateKey)
	if err != nil {
		return Token{}, err
	}

	return Token{
		AccessToken: at,
		TokenType:   "Bearer",
	}, nil
}
