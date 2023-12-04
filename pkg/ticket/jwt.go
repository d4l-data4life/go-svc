package ticket

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

// Ticketer creates and verifies tickets (jwts) used to authorize downloads from data-receiver
type Ticketer struct {
	key      []byte
	Validity time.Duration
}

func NewTicketer(key []byte, validity time.Duration) *Ticketer {
	return &Ticketer{key, validity}
}

func (t Ticketer) Create(studyID string, subjectIDs []uuid.UUID) (string, error) {
	ticket := jwt.NewWithClaims(jwt.SigningMethodHS256,
		Claims{
			Expiration: time.Now().Add(t.Validity).Unix(),
			StudyID:    studyID,
			SubjectIDs: subjectIDs,
		})
	return ticket.SignedString(t.key)
}

// Verify a Ticket (jwt) using a given key
// If the token is valid, it returns the TicketClaims object
func (t Ticketer) Verify(token string) (*Claims, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &Claims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return t.key, nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed verifying token")
	}

	claims, ok := parsedToken.Claims.(*Claims)
	if !ok {
		return nil, errors.New("failed getting ticket claims")
	}
	return claims, nil
}
