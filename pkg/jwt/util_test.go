package jwt

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

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
	testTenantID = "tenant_1"
)

type testData struct {
	name       string
	request    *http.Request
	middleware func(http.Handler) http.Handler
	checks     []checkFunc
	endHandler func(http.ResponseWriter, *http.Request)
}

////////////////////////////////////////////////////////////////////////////////
// Check Funcs
////////////////////////////////////////////////////////////////////////////////

type checkFunc func(w *httptest.ResponseRecorder) error

func checks(fns ...checkFunc) []checkFunc { return fns }

func hasStatusCode(want int) checkFunc {
	return func(r *httptest.ResponseRecorder) error {
		if have := r.Code; want != have {
			return fmt.Errorf("\nwant: %d\nhave: %d", want, have)
		}

		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// HTTP Request Builder and OKHandler
////////////////////////////////////////////////////////////////////////////////

type requestBuilder func(*http.Request) error

func buildRequest(url string, fns ...requestBuilder) *http.Request {
	r := httptest.NewRequest("", url, nil)

	for _, fn := range fns {
		_ = fn(r)
	}

	return r
}

func withAuthHeader(key *rsa.PrivateKey, owner uuid.UUID, scopes ...string) requestBuilder {
	scope := strings.Join(scopes, " ")

	return func(r *http.Request) error {
		bt, err := bearerToken(
			key, owner, scope,
		)
		if err != nil {
			return err
		}

		r.Header.Add("Authorization", bt)

		return nil
	}
}

func withOwnerPath(owner string) string {
	var builder strings.Builder

	builder.WriteString("/users/")
	builder.WriteString(owner)
	builder.WriteString("/records/456")

	return builder.String()
}

func withOwnerURL(owner string) string {
	var builder strings.Builder

	builder.WriteString("http://test.data4life.care")
	builder.WriteString(withOwnerPath(owner))

	return builder.String()
}

func okHandler(w http.ResponseWriter, r *http.Request) {}

////////////////////////////////////////////////////////////////////////////////
// Test Logger
////////////////////////////////////////////////////////////////////////////////

type testLogger struct{}

func (testLogger) ErrUserAuth(ctx context.Context, err error) error {
	fmt.Println(err)
	return nil
}
func (testLogger) InfoGeneric(ctx context.Context, msg string) error {
	fmt.Println(msg)
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// JWT Token generaration
////////////////////////////////////////////////////////////////////////////////

// bearerToken creates an Authorization Bearer header.
// Only owner and scopes are under test.
func bearerToken(
	key *rsa.PrivateKey,
	ownerID uuid.UUID,
	scope string,
) (string, error) {
	appUUID := uuid.Must(uuid.FromString(testAppID))
	jwtUUID := uuid.Must(uuid.FromString(testJWTID))
	userUUID := uuid.Must(uuid.FromString(testUserID))
	scp, err := NewScope(scope)
	if err != nil {
		return "", err
	}

	token, err := generateToken(
		key, time.Now(),
		ownerID, appUUID, jwtUUID, userUUID,
		testTenantID, testClientID, testIssuer,
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

// generateToken creates a new JWT with the given claims and signed with the given private key.
// In case of errors, it panics.
func generateToken(
	privateKey *rsa.PrivateKey,
	tm time.Time,
	ownerID, appID, jwtID, userID uuid.UUID,
	tenantID, clientID, issuer string,
	scope Scope,
) (string, error) {

	t := jwt.NewWithClaims(jwt.SigningMethodRS256, &Claims{
		Issuer:     issuer,
		Subject:    Owner{ownerID},
		Expiration: Time(tm.Add(time.Minute)),
		NotBefore:  Time(tm.Add(-skew)),
		IssuedAt:   Time(tm),
		JWTID:      jwtID,
		AppID:      appID,
		ClientID:   clientID,
		TenantID:   tenantID,
		UserID:     userID,
		Scope:      scope,
	})

	return t.SignedString(privateKey)
}
