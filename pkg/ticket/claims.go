package ticket

import (
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrNoClaims = errors.New("missing claims")
)

// Claims represents the claims used in the ticket to authenticate and authorize data download requests
// between research-pillars and data-receiver
type Claims struct {
	// Expiration is the expiration claim according to https://tools.ietf.org/html/rfc7519#section-4.1.4
	Expiration int64       `json:"exp"`
	StudyID    string      `json:"studyID"`
	SubjectIDs []uuid.UUID `json:"subjectIds"`
}

func (t Claims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(t.Expiration, 0)), nil
}

// Fulfill the https://pkg.go.dev/github.com/golang-jwt/jwt/v5#Claims interface even though unused
func (t Claims) GetIssuedAt() (*jwt.NumericDate, error) {
	return nil, nil
}
func (t Claims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}
func (t Claims) GetIssuer() (string, error) {
	return "", nil
}
func (t Claims) GetSubject() (string, error) {
	return "", nil
}
func (t Claims) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}
