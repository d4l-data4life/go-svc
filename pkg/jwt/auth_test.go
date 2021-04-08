package jwt

import (
	"crypto/rsa"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
)

const (
	ownerUUID = "926c97ac-fabf-4e3d-b9ca-0930b3bb7c3c"
	otherUUID = "91424b78-d372-452e-9ff3-81aba040735e"
)

func TestGorillaAuthenticator(t *testing.T) {
	// Prepare data
	read := rand.New(rand.NewSource(time.Now().Unix()))
	priv, err := rsa.GenerateKey(read, 1024)
	if err != nil {
		t.Fatal(err)
	}
	auth := New(&priv.PublicKey, &testLogger{})
	ownerFlag := "owner"

	for _, tc := range [...]testData{
		{
			name: "should succeed with right owner and right scopes",
			middleware: auth.Verify(
				WithGorillaOwner(ownerFlag),
				WithScopes(
					TokenAttachmentsWrite,
				),
			),
			request: buildRequest(
				withOwnerURL(ownerUUID),
				withAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					TokenAttachmentsWrite,
				),
			),
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},

		{
			name: "should fail if non sense is given as vars key",
			middleware: auth.Verify(
				WithGorillaOwner("GG"),
				WithScopes(
					TokenAttachmentsWrite,
				),
			),
			request: buildRequest(
				withOwnerURL(ownerUUID),
				withAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					TokenAttachmentsWrite,
				),
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},

		{
			name: "should fail on uuid.Nil in request path",
			middleware: auth.Verify(
				WithGorillaOwner(ownerFlag),
				WithScopes(
					TokenAttachmentsWrite,
				),
			),
			request: buildRequest(
				withOwnerURL(uuid.Nil.String()),
				withAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					TokenAttachmentsWrite,
				),
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},

		{
			name: "should fail on wrong user ID in request path",
			middleware: auth.Verify(
				WithGorillaOwner(ownerFlag),
				WithScopes(
					TokenAttachmentsWrite,
				),
			),
			request: buildRequest(
				withOwnerURL(otherUUID),
				withAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					TokenAttachmentsWrite,
				),
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
	} {
		tc := tc // Pin Variable

		t.Run(tc.name, func(t *testing.T) {
			handler := tc.middleware(http.HandlerFunc(okHandler))
			res := httptest.NewRecorder()

			router := mux.NewRouter()
			router.Handle(
				withOwnerPath("{"+ownerFlag+"}"),
				handler,
			)

			router.ServeHTTP(res, tc.request)

			for _, check := range tc.checks {
				if err := check(res); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
}

// TODO add tests for all other cases (e.g. expired, wrong issuer, ...)
func TestAuthenticator(t *testing.T) {
	// Prepare data
	read := rand.New(rand.NewSource(time.Now().Unix()))
	priv, err := rsa.GenerateKey(read, 1024)
	if err != nil {
		t.Fatal(err)
	}
	auth := New(&priv.PublicKey, &testLogger{})

	for _, tc := range [...]testData{
		{
			name: "should succeed with right owner and right scopes",
			middleware: auth.Verify(
				WithOwner(func(r *http.Request) uuid.UUID {
					return uuid.Must(uuid.FromString(ownerUUID))
				}),
				WithScopes(
					TokenAttachmentsWrite,
				),
			),
			request: buildRequest(
				withOwnerURL(ownerUUID),
				withAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					TokenAttachmentsWrite,
				),
			),
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},

		{
			name: "should succeed with right owner and any scope (:w)",
			middleware: auth.Verify(
				WithOwner(func(r *http.Request) uuid.UUID {
					return uuid.Must(uuid.FromString(ownerUUID))
				}),
				WithAnyScope(
					TokenAttachmentsRead,
					TokenAttachmentsWrite,
				),
			),
			request: buildRequest(
				withOwnerURL(ownerUUID),
				withAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					TokenAttachmentsWrite,
				),
			),
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},

		{
			name: "should respond with '401 Hasta La Vista' on wrong owner",
			middleware: auth.Verify(
				WithOwner(func(r *http.Request) uuid.UUID {
					return uuid.Must(uuid.FromString(ownerUUID))
				}),
				WithScopes(
					TokenAttachmentsWrite,
				),
			),
			request: buildRequest(
				withOwnerURL(ownerUUID),
				withAuthHeader(
					priv,
					uuid.Must(uuid.FromString(otherUUID)),
					TokenAttachmentsWrite,
				),
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},

		{
			name: "should respond with '401 Hasta La Vista' on wrong scopes",
			middleware: auth.Verify(
				WithOwner(func(r *http.Request) uuid.UUID {
					return uuid.Must(uuid.FromString(ownerUUID))
				}),
				WithScopes(
					TokenAttachmentsWrite,
				),
			),
			request: buildRequest(
				withOwnerURL(ownerUUID),
				withAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					TokenAttachmentsRead,
				),
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},

		{
			name: "should fail to interpret broken bearer token",
			middleware: auth.Verify(
				WithOwner(func(r *http.Request) uuid.UUID {
					return uuid.Must(uuid.FromString(ownerUUID))
				}),
				WithScopes(
					TokenAttachmentsWrite,
				),
			),
			request: buildRequest(
				withOwnerURL(ownerUUID),
				func(r *http.Request) error {
					r.Header.Add("Authorization", "I haz master key!")
					return nil
				},
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},

		{
			name: "should fail to extract a token from malformed request",
			middleware: auth.Verify(
				WithOwner(func(r *http.Request) uuid.UUID {
					return uuid.Must(uuid.FromString(ownerUUID))
				}),
				WithScopes(
					TokenAttachmentsWrite,
				),
			),
			request: &http.Request{},
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
	} {
		tc := tc // Pin Variable

		t.Run(tc.name, func(t *testing.T) {
			handler := tc.middleware(http.HandlerFunc(okHandler))
			res := httptest.NewRecorder()

			handler.ServeHTTP(res, tc.request)

			for _, check := range tc.checks {
				if err := check(res); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
}
