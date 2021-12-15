package jwt_test

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/d4lcontext"
	"github.com/gesundheitscloud/go-svc/pkg/dynamic"
	"github.com/gesundheitscloud/go-svc/pkg/jwt"
	"github.com/gesundheitscloud/go-svc/pkg/jwt/testutils"

	"github.com/go-chi/chi"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestWithGorillaOwner(t *testing.T) {
	read := rand.New(rand.NewSource(time.Now().Unix()))
	priv, err := rsa.GenerateKey(read, 1024)
	if err != nil {
		t.Fatal(err)
	}
	pkp := testutils.DummyKeyProvider{Key: &priv.PublicKey}
	l := testutils.Logger{}
	auth := jwt.NewAuthenticator(&pkp, &l)
	authAcceptingCookie := jwt.NewAuthenticatorWithOptions(&pkp, &l, jwt.AcceptAccessCookie)

	ownerFlag := "owner"
	ownerUUID := uuid.Must(uuid.NewV4())
	otherUUID := uuid.Must(uuid.NewV4())

	for _, tc := range [...]testData{
		{
			name: "should succeed with right owner: JWT in authorization header",
			middleware: auth.Verify(
				jwt.WithGorillaOwner(ownerFlag),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", ownerUUID)),
				testutils.WithAuthHeader(
					priv,
					jwt.WithUserID(ownerUUID),
				),
			),
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should succeed with right owner: JWT in form body",
			middleware: auth.Verify(
				jwt.WithGorillaOwner(ownerFlag),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", ownerUUID)),
				testutils.WithFormAccessToken(
					priv,
					jwt.WithUserID(ownerUUID),
				),
			),
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should succeed with right owner: JWT in cookie",
			middleware: authAcceptingCookie.Verify(
				jwt.WithGorillaOwner(ownerFlag),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", ownerUUID)),
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithUserID(ownerUUID),
				),
			),
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should succeed with valid cookie if option is disabled",
			middleware: auth.Verify(
				jwt.WithGorillaOwner(ownerFlag),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", ownerUUID)),
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithUserID(ownerUUID),
				),
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should fail if non sense is given as vars key",
			middleware: auth.Verify(
				jwt.WithGorillaOwner("GG"),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", ownerUUID)),
				testutils.WithAuthHeader(
					priv,
					jwt.WithUserID(ownerUUID),
				),
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should fail on uuid.Nil in request path",
			middleware: auth.Verify(
				jwt.WithGorillaOwner(ownerFlag),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", uuid.Nil.String())),
				testutils.WithAuthHeader(
					priv,
					jwt.WithUserID(ownerUUID),
				),
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should fail on wrong user ID in request path",
			middleware: auth.Verify(
				jwt.WithGorillaOwner(ownerFlag),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", otherUUID)),
				testutils.WithAuthHeader(
					priv,
					jwt.WithUserID(ownerUUID),
				),
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			handler := tc.middleware(http.HandlerFunc(testutils.OkHandler))
			res := httptest.NewRecorder()

			router := mux.NewRouter()
			router.Handle(
				"/users/{"+ownerFlag+"}/records",
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

func TestWithChiOwner(t *testing.T) {
	read := rand.New(rand.NewSource(time.Now().Unix()))
	priv, err := rsa.GenerateKey(read, 1024)
	if err != nil {
		t.Fatal(err)
	}
	pkp := testutils.DummyKeyProvider{Key: &priv.PublicKey}
	l := testutils.Logger{}
	auth := jwt.NewAuthenticator(&pkp, &l)
	authAcceptingCookie := jwt.NewAuthenticatorWithOptions(&pkp, &l, jwt.AcceptAccessCookie)
	ownerFlag := "owner"
	ownerUUID := uuid.Must(uuid.NewV4())
	otherUUID := uuid.Must(uuid.NewV4())

	for _, tc := range [...]testData{
		{
			name: "should succeed with right owner: JWT in authorization header",
			middleware: auth.Verify(
				jwt.WithChiOwner(ownerFlag),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", ownerUUID)),
				testutils.WithAuthHeader(
					priv,
					jwt.WithUserID(ownerUUID),
				),
			),
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should succeed with right owner: JWT in form body",
			middleware: auth.Verify(
				jwt.WithChiOwner(ownerFlag),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", ownerUUID)),
				testutils.WithFormAccessToken(
					priv,
					jwt.WithUserID(ownerUUID),
				),
			),
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should succeed with right owner: JWT in cookie and cookie option enabled",
			middleware: authAcceptingCookie.Verify(
				jwt.WithChiOwner(ownerFlag),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", ownerUUID)),
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithUserID(ownerUUID),
				),
			),
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should fail with right owner if cookie option disabled",
			middleware: auth.Verify(
				jwt.WithChiOwner(ownerFlag),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", ownerUUID)),
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithUserID(ownerUUID),
				),
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should fail if non sense is given as vars key",
			middleware: auth.Verify(
				jwt.WithChiOwner("GG"),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", ownerUUID)),
				testutils.WithAuthHeader(
					priv,
					jwt.WithUserID(ownerUUID),
				),
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should fail on uuid.Nil in request path",
			middleware: auth.Verify(
				jwt.WithChiOwner(ownerFlag),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", uuid.Nil.String())),
				testutils.WithAuthHeader(
					priv,
					jwt.WithUserID(ownerUUID),
				),
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should fail on wrong user ID in request path",
			middleware: auth.Verify(
				jwt.WithChiOwner(ownerFlag),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", otherUUID)),
				testutils.WithAuthHeader(
					priv,
					jwt.WithUserID(ownerUUID),
				),
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			handler := tc.middleware(http.HandlerFunc(testutils.OkHandler))
			res := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Handle(
				"/users/{"+ownerFlag+"}/records",
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

func TestWithOwner(t *testing.T) {
	read := rand.New(rand.NewSource(time.Now().Unix()))
	priv, err := rsa.GenerateKey(read, 1024)
	if err != nil {
		t.Fatal(err)
	}
	pkp := testutils.DummyKeyProvider{Key: &priv.PublicKey}
	l := testutils.Logger{}
	auth := jwt.NewAuthenticator(&pkp, &l)
	authAcceptingCookie := jwt.NewAuthenticatorWithOptions(&pkp, &l, jwt.AcceptAccessCookie)
	ownerUUID := uuid.Must(uuid.NewV4())
	otherUUID := uuid.Must(uuid.NewV4())

	for _, tc := range [...]testData{
		{
			name: "should succeed with right owner: JWT in authorization header",
			middleware: auth.Verify(
				jwt.WithOwner(func(r *http.Request) uuid.UUID {
					return ownerUUID
				}),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", ownerUUID)),
				testutils.WithAuthHeader(
					priv,
					jwt.WithUserID(ownerUUID),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should succeed with right owner: JWT in form body",
			middleware: auth.Verify(
				jwt.WithOwner(func(r *http.Request) uuid.UUID {
					return ownerUUID
				}),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", ownerUUID)),
				testutils.WithFormAccessToken(
					priv,
					jwt.WithUserID(ownerUUID),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should succeed with right owner: JWT in cookie and option enabled",
			middleware: authAcceptingCookie.Verify(
				jwt.WithOwner(func(r *http.Request) uuid.UUID {
					return ownerUUID
				}),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", ownerUUID)),
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithUserID(ownerUUID),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should fail with right owner if JWT in cookie and option disabled",
			middleware: auth.Verify(
				jwt.WithOwner(func(r *http.Request) uuid.UUID {
					return ownerUUID
				}),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", ownerUUID)),
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithUserID(ownerUUID),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should respond with 401 on wrong owner",
			middleware: auth.Verify(
				jwt.WithOwner(func(r *http.Request) uuid.UUID {
					return uuid.Must(uuid.NewV4())
				}),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", ownerUUID)),
				testutils.WithAuthHeader(
					priv,
					jwt.WithUserID(otherUUID),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should fail to interpret broken bearer token",
			middleware: auth.Verify(
				jwt.WithOwner(func(r *http.Request) uuid.UUID {
					return ownerUUID
				}),
			),
			request: testutils.BuildRequest(
				testutils.WithTargetURL(fmt.Sprintf("/users/%s/records", ownerUUID)),
				func(r *http.Request) error {
					r.Header.Add("Authorization", "I haz master key!")
					return nil
				},
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should fail to extract a token from malformed request",
			middleware: auth.Verify(
				jwt.WithOwner(func(r *http.Request) uuid.UUID {
					return ownerUUID
				}),
			),
			request:    &http.Request{},
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			handler := tc.middleware(http.HandlerFunc(tc.endHandler))
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

func TestWithAllScopes(t *testing.T) {
	read := rand.New(rand.NewSource(time.Now().Unix()))
	priv, err := rsa.GenerateKey(read, 1024)
	if err != nil {
		t.Fatal(err)
	}
	pkp := testutils.DummyKeyProvider{Key: &priv.PublicKey}
	l := testutils.Logger{}
	auth := jwt.NewAuthenticator(&pkp, &l)
	authAcceptingCookie := jwt.NewAuthenticatorWithOptions(&pkp, &l, jwt.AcceptAccessCookie)

	for _, tc := range [...]testData{
		{
			name: "should succeed with right scopes: one scope",
			middleware: auth.Verify(
				jwt.WithAllScopes(
					jwt.TokenAttachmentsWrite,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should succeed with right scopes: multiple scopes",
			middleware: auth.Verify(
				jwt.WithAllScopes(
					jwt.TokenAttachmentsWrite, jwt.TokenAppKeysRead,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite, jwt.TokenAppKeysRead),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should succeed with right scopes but different order",
			middleware: auth.Verify(
				jwt.WithAllScopes(
					jwt.TokenAttachmentsWrite, jwt.TokenAppKeysRead,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
					jwt.WithScopeStrings(
						jwt.TokenAppKeysRead,
						jwt.TokenAttachmentsWrite,
					),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should succeed if more scopes than required are included",
			middleware: auth.Verify(
				jwt.WithAllScopes(
					jwt.TokenAttachmentsWrite, jwt.TokenAppKeysRead,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
					jwt.WithScopeStrings(
						jwt.TokenAppKeysRead,
						jwt.TokenAttachmentsWrite,
						jwt.TokenAppKeysAppend,
					),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should ignore unknown scopes",
			middleware: auth.Verify(
				jwt.WithAllScopes(
					jwt.TokenAttachmentsWrite, jwt.TokenAppKeysRead,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
					jwt.WithScopeStrings(
						jwt.TokenAttachmentsWrite,
						jwt.TokenAppKeysRead,
						"unknown",
					),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should respond with 401 on missing scope",
			middleware: auth.Verify(
				jwt.WithAllScopes(
					jwt.TokenAttachmentsWrite,
					jwt.TokenAttachmentsRead,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
					jwt.WithScopeStrings(jwt.TokenAttachmentsRead),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should work with the jwt in form body",
			middleware: auth.Verify(
				jwt.WithAllScopes(
					jwt.TokenAttachmentsWrite,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithFormAccessToken(
					priv,
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should work with the jwt cookie if option is enables",
			middleware: authAcceptingCookie.Verify(
				jwt.WithAllScopes(
					jwt.TokenAttachmentsWrite,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should fail with the jwt cookie if option is disabled",
			middleware: auth.Verify(
				jwt.WithAllScopes(
					jwt.TokenAttachmentsWrite,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			handler := tc.middleware(http.HandlerFunc(tc.endHandler))
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

func TestWithAnyScopes(t *testing.T) {
	read := rand.New(rand.NewSource(time.Now().Unix()))
	priv, err := rsa.GenerateKey(read, 1024)
	if err != nil {
		t.Fatal(err)
	}
	pkp := testutils.DummyKeyProvider{Key: &priv.PublicKey}
	l := testutils.Logger{}
	auth := jwt.NewAuthenticator(&pkp, &l)
	authAcceptingCookie := jwt.NewAuthenticatorWithOptions(&pkp, &l, jwt.AcceptAccessCookie)

	for _, tc := range [...]testData{
		{
			name: "should succeed with right scopes: one scope",
			middleware: auth.Verify(
				jwt.WithAnyScope(
					jwt.TokenAttachmentsWrite,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should succeed with right scopes: multiple scopes",
			middleware: auth.Verify(
				jwt.WithAnyScope(
					jwt.TokenAttachmentsWrite, jwt.TokenAppKeysRead,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite, jwt.TokenAppKeysRead),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should succeed with right scopes but different order",
			middleware: auth.Verify(
				jwt.WithAnyScope(
					jwt.TokenAttachmentsWrite, jwt.TokenAppKeysRead,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
					jwt.WithScopeStrings(
						jwt.TokenAppKeysRead,
						jwt.TokenAttachmentsWrite,
					),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should succeed if more scopes than required are included",
			middleware: auth.Verify(
				jwt.WithAnyScope(
					jwt.TokenAttachmentsWrite, jwt.TokenAppKeysRead,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
					jwt.WithScopeStrings(
						jwt.TokenAppKeysRead,
						jwt.TokenAttachmentsWrite,
						jwt.TokenAppKeysAppend,
					),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should succeed if a subset of scopes are included",
			middleware: auth.Verify(
				jwt.WithAnyScope(
					jwt.TokenAttachmentsWrite,
					jwt.TokenAttachmentsRead,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
					jwt.WithScopeStrings(jwt.TokenAttachmentsRead),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should ignore unknown scopes",
			middleware: auth.Verify(
				jwt.WithAnyScope(
					jwt.TokenAttachmentsWrite, jwt.TokenAppKeysRead,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
					jwt.WithScopeStrings(
						"unknown",
						jwt.TokenAttachmentsWrite,
					),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should respond with 401 if none of the scopes are included - other scope",
			middleware: auth.Verify(
				jwt.WithAnyScope(
					jwt.TokenAttachmentsWrite,
					jwt.TokenAttachmentsRead,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
					jwt.WithScopeStrings(jwt.TokenRecordsWrite),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should respond with 401 if none of the scopes are included - no scope at all",
			middleware: auth.Verify(
				jwt.WithAnyScope(
					jwt.TokenAttachmentsWrite,
					jwt.TokenAttachmentsRead,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should work with the jwt in form body",
			middleware: auth.Verify(
				jwt.WithAnyScope(
					jwt.TokenAttachmentsWrite,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithFormAccessToken(
					priv,
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should work with the jwt in cookie if option is enabled",
			middleware: authAcceptingCookie.Verify(
				jwt.WithAnyScope(
					jwt.TokenAttachmentsWrite,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should fail with the jwt in cookie if option is disabled",
			middleware: auth.Verify(
				jwt.WithAnyScope(
					jwt.TokenAttachmentsWrite,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			handler := tc.middleware(http.HandlerFunc(tc.endHandler))
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

func TestVerify(t *testing.T) {
	read := rand.New(rand.NewSource(time.Now().Unix()))
	priv, err := rsa.GenerateKey(read, 1024)
	if err != nil {
		t.Fatal(err)
	}
	pkp := testutils.DummyKeyProvider{Key: &priv.PublicKey}
	l := testutils.Logger{}
	auth := jwt.NewAuthenticator(&pkp, &l)
	authAcceptingCookie := jwt.NewAuthenticatorWithOptions(&pkp, &l, jwt.AcceptAccessCookie)
	ownerUUID := uuid.Must(uuid.NewV4())

	for _, tc := range [...]struct {
		name       string
		middleware func(http.Handler) http.Handler
		request    *http.Request
		checks     []checkFunc
		reqChecks  checkReqFunc
	}{
		{
			name:       "should fail on broken Authorization header",
			middleware: auth.Verify(),
			request: testutils.BuildRequest(
				func(r *http.Request) error {
					r.Header.Add("Authorization", "wrong")
					return nil
				},
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name:       "should fail to interpret a broken bearer token",
			middleware: auth.Verify(),
			request: testutils.BuildRequest(
				func(r *http.Request) error {
					r.Header.Add("Authorization", "Bearer wrong")
					return nil
				},
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name:       "should fail missing Authorization header",
			middleware: auth.Verify(),
			request:    testutils.BuildRequest(),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},

		{
			name:       "end handler should receive a request with JWT claims in the context",
			middleware: auth.Verify(),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
					jwt.WithUserID(ownerUUID),
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
			),
			reqChecks: checkReqAll(
				hasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetSubjectID(ctx)
				}, ownerUUID),
				hasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetTagsInScope(ctx)
				}, []jwt.Tag{jwt.TokenAttachmentsWrite}),
			),
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name:       "should fail on broken form access token",
			middleware: auth.Verify(),
			request: testutils.BuildRequest(
				func(r *http.Request) error {
					r.Form = url.Values{}
					r.Form.Add("access_token", "wrong")
					r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					return nil
				},
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name:       "should work with a valid jwt in authorization header",
			middleware: auth.Verify(),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
				),
			),
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name:       "should work with a valid jwt in form body",
			middleware: auth.Verify(),
			request: testutils.BuildRequest(
				testutils.WithFormAccessToken(
					priv,
				),
			),
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name:       "should work with a valid jwt in cookie if option enabled",
			middleware: authAcceptingCookie.Verify(),
			request: testutils.BuildRequest(
				testutils.WithCookieAccessToken(
					priv,
				),
			),
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name:       "should fail with a valid jwt in cookie if option disabled",
			middleware: auth.Verify(),
			request: testutils.BuildRequest(
				testutils.WithCookieAccessToken(
					priv,
				),
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			handler := tc.middleware(http.HandlerFunc(testutils.OkHandler))
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

func generatePEMPublicKey(t *testing.T, pk *rsa.PublicKey, indent int) []byte {
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

func TestNewAuthenticator_MultiplePubKeys(t *testing.T) {
	read := rand.New(rand.NewSource(time.Now().Unix()))
	priv1, err := rsa.GenerateKey(read, 1024)
	assert.NoError(t, err)
	priv2, err := rsa.GenerateKey(read, 1024)
	assert.NoError(t, err)
	priv3, err := rsa.GenerateKey(read, 1024)
	assert.NoError(t, err)
	priv4, err := rsa.GenerateKey(read, 1024)
	assert.NoError(t, err)
	ownerUUID := uuid.Must(uuid.NewV4())

	// use 2-space indent inside this string
	configYaml := []byte(`
JWTPublicKey:
- name: "key1"
  comment: "valid test key1"
  not_before: 1410-01-01
  not_after: 2099-01-01
  key: |
    ` + string(generatePEMPublicKey(t, &priv1.PublicKey, 4)) + `
- name: "key2"
  comment: "valid test key2"
  not_before: 1410-01-01
  not_after: 2099-01-01
  key: |
    ` + string(generatePEMPublicKey(t, &priv2.PublicKey, 4)) + `
- name: "expiredKey3"
  comment: "valid key but metadata says it should not be used"
  not_before: 1999-01-01
  not_after: 1999-01-02
  key: |
    ` + string(generatePEMPublicKey(t, &priv3.PublicKey, 4)) + `
`)
	t.Logf("using config:\n%s", configYaml)

	kp := dynamic.NewViperConfig("auth_test",
		dynamic.WithConfigSource(bytes.NewBuffer(configYaml)),
		dynamic.WithConfigFormat("yaml"),
	)
	err = kp.Bootstrap()
	assert.NoError(t, err)

	keys, err := kp.JWTPublicKeys()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(keys))

	auth := jwt.NewAuthenticator(kp, &testutils.Logger{})

	for _, tc := range [...]testData{
		{
			name: "should succeed with key1",
			middleware: auth.Verify(
				jwt.WithOwner(func(r *http.Request) uuid.UUID {
					return ownerUUID
				}),
				jwt.WithAllScopes(
					jwt.TokenAttachmentsWrite,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv1,
					jwt.WithUserID(ownerUUID),
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should succeed with key2",
			middleware: auth.Verify(
				jwt.WithOwner(func(r *http.Request) uuid.UUID {
					return ownerUUID
				}),
				jwt.WithAllScopes(
					jwt.TokenAttachmentsWrite,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv2,
					jwt.WithUserID(ownerUUID),
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should ignore metadata and work with key3", // TODO-PR: Change this case when handling of metadata will be implemented
			middleware: auth.Verify(
				jwt.WithOwner(func(r *http.Request) uuid.UUID {
					return ownerUUID
				}),
				jwt.WithAllScopes(
					jwt.TokenAttachmentsWrite,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv3,
					jwt.WithUserID(ownerUUID),
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should fail with all 3 keys not matching the private key",
			middleware: auth.Verify(
				jwt.WithOwner(func(r *http.Request) uuid.UUID {
					return ownerUUID
				}),
				jwt.WithAllScopes(
					jwt.TokenAttachmentsWrite,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv4,
					jwt.WithUserID(ownerUUID),
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
			),
			endHandler: testutils.OkHandler,
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			handler := tc.middleware(http.HandlerFunc(tc.endHandler))
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

func TestExtract(t *testing.T) {
	read := rand.New(rand.NewSource(time.Now().Unix()))
	priv, err := rsa.GenerateKey(read, 1024)
	if err != nil {
		t.Fatal(err)
	}
	auth := jwt.NewAuthenticator(&testutils.DummyKeyProvider{Key: &priv.PublicKey}, &testutils.Logger{})

	userID := uuid.Must(uuid.NewV4())
	clientID := uuid.Must(uuid.NewV4())
	tenantID := "some-tenant"

	for _, tc := range [...]struct {
		name           string
		middleware     func(http.Handler) http.Handler
		request        *http.Request
		reqChecks      checkReqFunc
		responseChecks checkFunc
	}{
		{
			name:       "should succeed with request with valid JWT",
			middleware: auth.Extract,
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
					jwt.WithUserID(userID),
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
					jwt.WithClientID(clientID.String()),
					jwt.WithTenantID(tenantID),
				),
			),
			reqChecks: checkReqAll(
				hasInContext(d4lcontext.ClientIDContextKey, clientID.String()),
				hasInContext(d4lcontext.UserIDContextKey, userID.String()),
				hasInContext(d4lcontext.TenantIDContextKey, tenantID),
				hasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetSubjectID(ctx)
				}, userID),
				hasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetTenantID(ctx)
				}, tenantID),
				hasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetClientID(ctx)
				}, clientID.String()),
			),
		},
		{
			name:           "should not break the middleware chain with a request without a JWT",
			middleware:     auth.Extract,
			request:        testutils.BuildRequest(),
			responseChecks: hasStatusCode(http.StatusOK),
		},
		{
			name:       "should not break the middleware chain with a request with an unknown scope",
			middleware: auth.Extract,
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
					jwt.WithUserID(userID),
					jwt.WithScopeStrings("unknown"),
					jwt.WithClientID(clientID.String()),
					jwt.WithTenantID(tenantID),
				),
			),
			responseChecks: hasStatusCode(http.StatusOK),
		},
		{
			name:       "should not break the middleware chain with a JWT with missing data",
			middleware: auth.Extract,
			request: testutils.BuildRequest(
				testutils.WithAuthHeader(
					priv,
					jwt.WithUserID(uuid.Nil),
					jwt.WithScopeStrings(""),
					jwt.WithClientID(""),
					jwt.WithTenantID(""),
				),
			),
			responseChecks: hasStatusCode(http.StatusOK),
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var haveReq *http.Request
			hasBeenCalled := false

			handler := tc.middleware(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
				haveReq = req
				hasBeenCalled = true
			}))

			respRecorder := httptest.NewRecorder()
			handler.ServeHTTP(respRecorder, tc.request)

			if !hasBeenCalled {
				t.Fatal(errors.New("handler should have been called, was not."))
			}
			if tc.reqChecks != nil {
				if err := tc.reqChecks(haveReq); err != nil {
					t.Error(err)
				}
			}
			if tc.responseChecks != nil {
				if err := tc.responseChecks(respRecorder); err != nil {
					t.Error(err)
				}
			}
		})
	}
}
