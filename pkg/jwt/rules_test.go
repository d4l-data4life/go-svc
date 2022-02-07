package jwt_test

import (
	"crypto/rsa"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/jwt"
	"github.com/gesundheitscloud/go-svc/pkg/jwt/testutils"

	"github.com/go-chi/chi"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
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
				testutils.WithMethod(http.MethodGet),
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
			name: "should succeed with right owner: JWT in cookie",
			middleware: authAcceptingCookie.Verify(
				jwt.WithChiOwner(ownerFlag),
			),
			request: testutils.BuildRequest(
				testutils.WithMethod(http.MethodGet),
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
			name: "should succeed with right owner: JWT in cookie",
			middleware: authAcceptingCookie.Verify(
				jwt.WithOwner(func(r *http.Request) uuid.UUID {
					return ownerUUID
				}),
			),
			request: testutils.BuildRequest(
				testutils.WithMethod(http.MethodGet),
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
				func(r *http.Request) {
					r.Header.Add("Authorization", "I haz master key!")
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
			name: "should work with the jwt cookie",
			middleware: authAcceptingCookie.Verify(
				jwt.WithAllScopes(
					jwt.TokenAttachmentsWrite,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithMethod(http.MethodGet),
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
			name: "should work with the jwt in cookie",
			middleware: authAcceptingCookie.Verify(
				jwt.WithAnyScope(
					jwt.TokenAttachmentsWrite,
				),
			),
			request: testutils.BuildRequest(
				testutils.WithMethod(http.MethodGet),
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
