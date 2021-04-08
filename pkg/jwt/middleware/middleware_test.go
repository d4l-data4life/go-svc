package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/jwt"
	"github.com/gesundheitscloud/go-svc/pkg/jwt/test"
	"github.com/gofrs/uuid"
)

const (
	ownerUUID = "926c97ac-fabf-4e3d-b9ca-0930b3bb7c3c"
	otherUUID = "91424b78-d372-452e-9ff3-81aba040735e"
)

type testData struct {
	name    string
	request *http.Request
	handler http.Handler
	checks  []test.CheckFunc
}

func handlerWithAuthOpts(priv *rsa.PrivateKey, opt func(*verifier)) http.Handler {
	mw := Auth(
		&priv.PublicKey,
		&test.Logger{},
		WithGorillaMux("owner"),
		opt,
	)

	return test.BuildGorillaHandler(
		test.WithGorillaHandler(
			"/users/{owner:[A-Fa-f0-9-]+}/records/{record:[A-Fa-f0-9-]+}",
			http.HandlerFunc(test.OkHandler),
		),
		test.WithGorillaMiddleware(mw),
	)
}

func handlerWithAuthGorilla(priv *rsa.PrivateKey, scope ...string) http.Handler {
	auth := NewGorillaAuth(&priv.PublicKey, &test.Logger{}, "owner")

	return test.BuildGorillaHandler(
		test.WithGorillaHandler(
			"/users/{owner:[A-Fa-f0-9-]+}/records/{record:[A-Fa-f0-9-]+}",
			http.HandlerFunc(test.OkHandler),
		),
		test.WithGorillaMiddleware(auth.WithScopes(scope...)),
	)
}

func handlerWithAuthGorillaNoOwner(priv *rsa.PrivateKey, scope ...string) http.Handler {
	auth := NewGorillaAuth(&priv.PublicKey, &test.Logger{}, "owner")

	return test.BuildGorillaHandler(
		test.WithGorillaHandler(
			"/records",
			http.HandlerFunc(test.OkHandler),
		),
		test.WithGorillaMiddleware(auth.WithScopesNoOwner(scope...)),
	)
}

func TestAuthGorillaMiddleware(t *testing.T) {
	// Prepare data
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range [...]testData{
		{
			name: "should fail without authorization header",
			request: test.BuildRequest(
				test.WithOwnerURL(ownerUUID),
			),
			handler: handlerWithAuthGorilla(
				priv,
				jwt.TokenAttachmentsRead,
				jwt.TokenAttachmentsWrite,
			),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},

		{
			name: "should fail with wrong owner ID in authorization header",
			request: test.BuildRequest(
				test.WithOwnerURL(ownerUUID),
				test.WithAuthHeader(
					priv,
					uuid.Must(uuid.FromString(otherUUID)),
					jwt.TokenAttachmentsRead,
					jwt.TokenAttachmentsWrite,
				),
			),
			handler: handlerWithAuthGorilla(
				priv,
				jwt.TokenAttachmentsRead,
				jwt.TokenAttachmentsWrite,
			),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should fail with jwt uuid.Nil in authorization header for nil user",
			request: test.BuildRequest(
				test.WithOwnerURL(uuid.Nil.String()),
				test.WithAuthHeader(
					priv,
					uuid.Nil,
					jwt.TokenAttachmentsRead,
					jwt.TokenAttachmentsWrite,
				),
			),
			handler: handlerWithAuthGorilla(
				priv,
				jwt.TokenAttachmentsRead,
				jwt.TokenAttachmentsWrite,
			),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should fail with jwt uuid.Nil in authorization header for non nil user",
			request: test.BuildRequest(
				test.WithOwnerURL(ownerUUID),
				test.WithAuthHeader(
					priv,
					uuid.Nil,
					jwt.TokenAttachmentsRead,
					jwt.TokenAttachmentsWrite,
				),
			),
			handler: handlerWithAuthGorilla(
				priv,
				jwt.TokenAttachmentsRead,
				jwt.TokenAttachmentsWrite,
			),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should fail with jwt non nil user in authorization header for nil user",
			request: test.BuildRequest(
				test.WithOwnerURL(uuid.Nil.String()),
				test.WithAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					jwt.TokenAttachmentsRead,
					jwt.TokenAttachmentsWrite,
				),
			),
			handler: handlerWithAuthGorilla(
				priv,
				jwt.TokenAttachmentsRead,
				jwt.TokenAttachmentsWrite,
			),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should fail for right owner on required attachment rw scopes, but wrong scopes",

			request: test.BuildRequest(
				test.WithOwnerURL(ownerUUID),
				test.WithAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					jwt.TokenTags,
				),
			),
			handler: handlerWithAuthGorilla(
				priv,
				jwt.TokenAttachmentsRead,
				jwt.TokenAttachmentsWrite,
			),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should fail for right owner on required records rw scopes, but wrong scopes",
			request: test.BuildRequest(
				test.WithOwnerURL(ownerUUID),
				test.WithAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					jwt.TokenTags,
				),
			),
			handler: handlerWithAuthGorilla(
				priv,
				jwt.TokenRecordsRead,
				jwt.TokenRecordsWrite,
			),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should fail for right owner on required records rw scopes, but wrong scopes",
			request: test.BuildRequest(
				test.WithOwnerURL(ownerUUID),
				test.WithAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					jwt.TokenTags,
				),
			),
			handler: handlerWithAuthGorilla(
				priv,
				jwt.TokenAttachmentsWrite,
				jwt.TokenPermissionsWrite,
				jwt.TokenRecordsWrite,
				jwt.TokenUserWrite,
			),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should fail with non-uuid owner",
			request: test.BuildRequest(
				test.WithOwnerURL("deadbeef"),
				test.WithAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					jwt.TokenTags,
				),
			),
			handler: handlerWithAuthGorilla(
				priv,
				jwt.TokenAttachmentsRead,
				jwt.TokenAttachmentsWrite,
			),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "should succeed with right owner and right scopes",
			request: test.BuildRequest(
				test.WithOwnerURL(ownerUUID),
				test.WithAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					jwt.TokenAttachmentsRead,
					jwt.TokenAttachmentsWrite,
				),
			),
			handler: handlerWithAuthGorilla(
				priv,
				jwt.TokenAttachmentsRead,
				jwt.TokenAttachmentsWrite,
			),
			checks: test.Checks(
				test.HasStatusCode(http.StatusOK),
			),
		},
		{
			name: "no owner - should fail with missing scope",
			request: test.BuildRequest(
				test.WithURLNoOwner(),
				test.WithAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					jwt.TokenAttachmentsRead,
				),
			),
			handler: handlerWithAuthGorillaNoOwner(
				priv,
				jwt.TokenAttachmentsRead,
				jwt.TokenAttachmentsWrite,
			),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name: "no owner - should succeed with right scopes",
			request: test.BuildRequest(
				test.WithURLNoOwner(),
				test.WithAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					jwt.TokenAttachmentsRead,
					jwt.TokenAttachmentsWrite,
				),
			),
			handler: handlerWithAuthGorillaNoOwner(
				priv,
				jwt.TokenAttachmentsRead,
				jwt.TokenAttachmentsWrite,
			),
			checks: test.Checks(
				test.HasStatusCode(http.StatusOK),
			),
		},
		{ // This test was added for 100% test coverage for the package
			name: "should fail due to misconfiguration",
			request: test.BuildRequest(
				test.WithOwnerURL(ownerUUID),
				test.WithAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					jwt.TokenAttachmentsRead, jwt.TokenAttachmentsWrite,
				),
			),
			handler: func() http.Handler {
				mw := Auth(
					&priv.PublicKey,
					&test.Logger{},
					WithGorillaMux("owner"),
					WithScopes(
						jwt.TokenAttachmentsRead,
						jwt.TokenAttachmentsWrite,
					),
				)

				return test.BuildGorillaHandler(
					test.WithGorillaHandler(
						"/users/{haxxor:[A-Fa-f0-9-]+}/records/{record:[A-Fa-f0-9-]+}",
						http.HandlerFunc(test.OkHandler),
					),
					test.WithGorillaMiddleware(mw),
				)
			}(),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := httptest.NewRecorder()

			tc.handler.ServeHTTP(res, tc.request)

			for _, check := range tc.checks {
				if err := check(res); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
}

func TestGorillaMiddleware(t *testing.T) {
	// Prepare data
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range [...]testData{
		{
			name: "should fail without authorization header",
			request: test.BuildRequest(
				test.WithOwnerURL(ownerUUID),
			),
			handler: handlerWithAuthOpts(priv, WithScopes(
				jwt.TokenAttachmentsRead,
				jwt.TokenAttachmentsWrite,
			)),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},

		{
			name: "should fail with wrong owner ID in authorization header",
			request: test.BuildRequest(
				test.WithOwnerURL(ownerUUID),
				test.WithAuthHeader(
					priv,
					uuid.Must(uuid.FromString(otherUUID)),
					jwt.TokenAttachmentsRead,
					jwt.TokenAttachmentsWrite,
				),
			),
			handler: handlerWithAuthOpts(priv, WithScopes(
				jwt.TokenAttachmentsRead,
				jwt.TokenAttachmentsWrite,
			)),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},

		{
			name: "should fail with jwt uuid.Nil in authorization header for nil user",
			request: test.BuildRequest(
				test.WithOwnerURL(uuid.Nil.String()),
				test.WithAuthHeader(
					priv,
					uuid.Nil,
					jwt.TokenAttachmentsRead,
					jwt.TokenAttachmentsWrite,
				),
			),
			handler: handlerWithAuthOpts(priv, WithScopes(
				jwt.TokenAttachmentsRead,
				jwt.TokenAttachmentsWrite,
			)),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},

		{
			name: "should fail with jwt uuid.Nil in authorization header for non nil user",
			request: test.BuildRequest(
				test.WithOwnerURL(ownerUUID),
				test.WithAuthHeader(
					priv,
					uuid.Nil,
					jwt.TokenAttachmentsRead,
					jwt.TokenAttachmentsWrite,
				),
			),
			handler: handlerWithAuthOpts(priv, WithScopes(
				jwt.TokenAttachmentsRead,
				jwt.TokenAttachmentsWrite,
			)),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},

		{
			name: "should fail with jwt non nil user in authorization header for nil user",
			request: test.BuildRequest(
				test.WithOwnerURL(uuid.Nil.String()),
				test.WithAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					jwt.TokenAttachmentsRead,
					jwt.TokenAttachmentsWrite,
				),
			),
			handler: handlerWithAuthOpts(priv, WithScopes(
				jwt.TokenAttachmentsRead,
				jwt.TokenAttachmentsWrite,
			)),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},

		{
			name: "should fail for right owner on required attachment rw scopes, but wrong scopes",
			request: test.BuildRequest(
				test.WithOwnerURL(ownerUUID),
				test.WithAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					jwt.TokenTags,
				),
			),
			handler: handlerWithAuthOpts(priv, WithScopes(
				jwt.TokenAttachmentsRead,
				jwt.TokenAttachmentsWrite,
			)),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},

		{
			name: "should fail for right owner on required records rw scopes, but wrong scopes",
			request: test.BuildRequest(
				test.WithOwnerURL(ownerUUID),
				test.WithAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					jwt.TokenTags,
				),
			),
			handler: handlerWithAuthOpts(priv, WithScopes(
				jwt.TokenRecordsRead,
				jwt.TokenRecordsWrite,
			)),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},

		{
			name: "should fail for right owner on required records rw scopes, but wrong scopes",
			request: test.BuildRequest(
				test.WithOwnerURL(ownerUUID),
				test.WithAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					jwt.TokenTags,
				),
			),
			handler: handlerWithAuthOpts(priv, WithScopes(
				jwt.TokenAttachmentsWrite,
				jwt.TokenPermissionsWrite,
				jwt.TokenRecordsWrite,
				jwt.TokenUserWrite,
			)),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},

		{
			name: "should fail with non-uuid owner",
			request: test.BuildRequest(
				test.WithOwnerURL("deadbeef"),
				test.WithAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					jwt.TokenTags,
				),
			),
			handler: handlerWithAuthOpts(priv, WithScopes(
				jwt.TokenAttachmentsRead,
				jwt.TokenAttachmentsWrite,
			)),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},

		{
			name: "should succeed with right owner and right scopes",
			request: test.BuildRequest(
				test.WithOwnerURL(ownerUUID),
				test.WithAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					jwt.TokenAttachmentsRead,
					jwt.TokenAttachmentsWrite,
				),
			),
			handler: handlerWithAuthOpts(priv, WithScopes(
				jwt.TokenAttachmentsRead,
				jwt.TokenAttachmentsWrite,
			)),
			checks: test.Checks(
				test.HasStatusCode(http.StatusOK),
			),
		},

		{ // This test was added for 100% test coverage for the package
			name: "should fail due to misconfiguration",
			request: test.BuildRequest(
				test.WithOwnerURL(ownerUUID),
				test.WithAuthHeader(
					priv,
					uuid.Must(uuid.FromString(ownerUUID)),
					jwt.TokenAttachmentsRead, jwt.TokenAttachmentsWrite,
				),
			),
			handler: func() http.Handler {
				auth := NewGorillaAuth(
					&priv.PublicKey,
					&test.Logger{},
					"owner",
				)

				return test.BuildGorillaHandler(
					test.WithGorillaHandler(
						"/users/{haxxor:[A-Fa-f0-9-]+}/records/{record:[A-Fa-f0-9-]+}",
						http.HandlerFunc(test.OkHandler),
					),
					test.WithGorillaMiddleware(
						auth.WithScopes(
							jwt.TokenAttachmentsRead,
							jwt.TokenAttachmentsWrite,
						),
					),
				)
			}(),
			checks: test.Checks(
				test.HasStatusCode(http.StatusUnauthorized),
			),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := httptest.NewRecorder()

			tc.handler.ServeHTTP(res, tc.request)

			for _, check := range tc.checks {
				if err := check(res); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
}
