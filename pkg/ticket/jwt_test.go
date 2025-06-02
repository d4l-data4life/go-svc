package ticket_test

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gesundheitscloud/go-svc/pkg/ticket"
)

func GetTestKey(t *testing.T) []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)
	return key
}

func TestTicketProcess(t *testing.T) {
	key := GetTestKey(t)
	studyID := "test"
	subjectIDs := []uuid.UUID{
		uuid.Must(uuid.NewV4()),
		uuid.Must(uuid.NewV4()),
	}
	ticketer := ticket.NewTicketer(key, 30*time.Minute)
	ticketJWT, err := ticketer.Create(studyID, subjectIDs, true, true)
	require.NoError(t, err, "failed creating jwt")

	claims, err := ticketer.Verify(ticketJWT)
	require.NoError(t, err, "failed verifying jwt")
	assert.Equal(t, studyID, claims.StudyID)
	assert.Equal(t, subjectIDs, claims.SubjectIDs)

	// Should not succeed with wrong key
	wrongKey := GetTestKey(t)
	ticketer = ticket.NewTicketer(wrongKey, 30*time.Minute)
	emptyClaims, err := ticketer.Verify(ticketJWT)
	require.Error(t, err, "should fail with wrong key")
	require.ErrorIs(t, err, jwt.ErrSignatureInvalid)
	assert.Nil(t, emptyClaims)
}

func TestExpiredTicket(t *testing.T) {
	key := GetTestKey(t)
	ticketer := ticket.NewTicketer(key, -30*time.Minute)
	ticketJWT, err := ticketer.Create("test", nil, false, false)
	require.NoError(t, err, "failed creating ticket")

	claims, err := ticketer.Verify(ticketJWT)
	require.Error(t, err)
	require.ErrorIs(t, err, jwt.ErrTokenExpired)
	assert.Nil(t, claims)
}

func TestInvalidClaimsTicket(t *testing.T) {
	key := GetTestKey(t)
	ticketer := ticket.NewTicketer(key, 30*time.Minute)
	ticketJWT, err := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"exp":        time.Now().Add(30 * time.Minute).Unix(),
			"studyID":    3,
			"subjectIds": 3,
		}).SignedString(key)
	require.NoError(t, err, "failed creating ticket")

	claims, err := ticketer.Verify(ticketJWT)
	require.Error(t, err)
	require.ErrorIs(t, err, jwt.ErrTokenMalformed)
	assert.Nil(t, claims)
}
