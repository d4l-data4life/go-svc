package jwt

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

// ErrInvalidIssuer is the error when the JWT does not contain a known issuer
var (
	ErrInvalidIssuer   = errors.New("invalid issuer")
	ErrExpirationUnset = errors.New("expiration unset")
	ErrExpired         = errors.New("expired token")
	ErrNotBeforeUnset  = errors.New("not before unset")
	ErrNotValidYet     = errors.New("token not valid yet")
)

const (
	// IssuerGesundheitscloud is the issuer value, a constant set in conjunction with "iss"
	IssuerGesundheitscloud = "urn:ghc"
)

// Claims is the struct that represent JWT claims according to https://tools.ietf.org/html/rfc7519#section-4
// Parsed and filled by jwt-go. Valid satisfies necessary interface.
type Claims struct {
	// Subject is the subject according to https://tools.ietf.org/html/rfc7519#section-4.1.2
	Subject Owner `json:"sub"`

	// Issuer is the issuer claim according to https://tools.ietf.org/html/rfc7519#section-4.1.1
	Issuer string `json:"iss"`

	// Expiration is the expiration claim according to https://tools.ietf.org/html/rfc7519#section-4.1.4
	Expiration Time `json:"exp"`

	// NotBefore is "not before" claim according to https://tools.ietf.org/html/rfc7519#section-4.1.5
	NotBefore Time `json:"nbf"`

	// IssuedAt is the "issued at" claim according to https://tools.ietf.org/html/rfc7519#section-4.1.6
	IssuedAt Time `json:"iat"`

	// JWTID is the JWT ID according to https://tools.ietf.org/html/rfc7519#section-4.1.7
	JWTID uuid.UUID `json:"jti"`

	// AppID is the app ID claim (gesundheitscloud private claim)
	AppID uuid.UUID `json:"ghc:aid"`

	// ClientID is the Client ID claim (gesundheitscloud private claim)
	// TODO make it a type like Owner and Time
	ClientID string `json:"ghc:cid"`

	// UserID is the claim that encodes the user who originally requested the JWT (gesundheitscloud private claim)
	UserID uuid.UUID `json:"ghc:uid"`

	// TenantID is the claim that encodes the tenant ID to which the user belongs
	TenantID string `json:"ghc:tid"`

	// Scope is the claim that encodes granted permissions, aka the "scope" of this JWT (gesundheitscloud private claim)
	Scope Scope `json:"ghc:scope"`

	// Email is the email of the subject
	Email string `json:"email"`
}

func (c *Claims) valid(now time.Time) error {
	// According to https://github.com/gesundheitscloud/dev-docs/blob/master/design-documents/login-management.md#registered-claims,
	// the following invariants must be met:

	switch {

	// iss: MUST equal urn:ghc, else the request MUST be rejected.
	case c.Issuer != IssuerGesundheitscloud:
		return ErrInvalidIssuer

	// exp: If the current time is after or equal the expiration time,
	// the request MUST be rejected.
	// For ease of implementation, clock skew issues may be omitted.

	case time.Time(c.Expiration).IsZero():
		return ErrExpirationUnset

	case now.After(time.Time(c.Expiration)):
		return ErrExpired

	// nbf: If the current time is before or equal the "not before" time, the request MUST be rejected.
	// For ease of implementation, clock skew issues may be omitted.

	case time.Time(c.NotBefore).IsZero():
		return ErrNotBeforeUnset

	case time.Time(c.NotBefore).After(now):
		return ErrNotValidYet

	}

	// Note: iat: Currently, this is informative only and doesn't need to be checked.

	return nil
}

// Valid checks if the claims invariants ar met
func (c *Claims) Valid() error {
	return c.valid(time.Now())
}
