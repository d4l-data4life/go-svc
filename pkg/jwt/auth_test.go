package jwt_test

import (
	"bytes"
	"context"
	"crypto/rsa"
	"errors"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/d4lcontext"
	"github.com/gesundheitscloud/go-svc/pkg/dynamic"
	"github.com/gesundheitscloud/go-svc/pkg/jwt"
	"github.com/gesundheitscloud/go-svc/pkg/jwt/testutils"
	"github.com/gesundheitscloud/go-svc/pkg/tut"

	"github.com/gofrs/uuid"
	"github.com/justinas/nosurf"
	"github.com/stretchr/testify/assert"
)

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

	csrfCookie, csrfToken := tut.CSRFValues()

	for _, tc := range [...]struct {
		name       string
		middleware func(http.Handler) http.Handler
		request    *http.Request
		checks     tut.ResponseCheckFunc
		reqChecks  tut.RequestCheckFunc
	}{
		{
			name:       "should fail on broken Authorization header",
			middleware: auth.Verify(),
			request: tut.Request(
				func(r *http.Request) {
					r.Header.Add("Authorization", "wrong")
				},
			),
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name:       "should fail to interpret a broken bearer token",
			middleware: auth.Verify(),
			request: tut.Request(
				func(r *http.Request) {
					r.Header.Add("Authorization", "Bearer wrong")
				},
			),
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name:       "should fail missing Authorization header",
			middleware: auth.Verify(),
			request:    tut.Request(),
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusUnauthorized),
			),
		},

		{
			name:       "end handler should receive a request with JWT claims in the context",
			middleware: auth.Verify(),
			request: tut.Request(
				tut.ReqWithAuthHeader(
					ownerUUID,
					priv,
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
			),
			reqChecks: tut.CheckRequest(
				tut.ReqHasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetSubjectID(ctx)
				}, ownerUUID),
				tut.ReqHasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetTagsInScope(ctx)
				}, []jwt.Tag{jwt.TokenAttachmentsWrite}),
			),
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusOK),
			),
		},
		{
			name:       "should fail on broken form access token",
			middleware: auth.Verify(),
			request: tut.Request(
				func(r *http.Request) {
					r.Form = url.Values{}
					r.Form.Add("access_token", "wrong")
					r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				},
			),
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name:       "should work with a valid jwt in authorization header",
			middleware: auth.Verify(),
			request: tut.Request(
				tut.ReqWithAuthHeader(
					ownerUUID,
					priv,
				),
			),
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusOK),
			),
		},
		{
			name:       "should work with a valid jwt in form body",
			middleware: auth.Verify(),
			request: tut.Request(
				testutils.WithFormAccessToken(
					priv,
				),
			),
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusOK),
			),
		},
		{
			name:       "should work with a valid jwt in cookie and CSRF safe method",
			middleware: authAcceptingCookie.Verify(),
			request: tut.Request(
				tut.ReqWithMethod(http.MethodGet),
				testutils.WithCookieAccessToken(
					priv,
				),
			),
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusOK),
			),
		},
		{
			name:       "should work with a valid jwt in cookie with CSRF protection",
			middleware: authAcceptingCookie.Verify(),
			request: tut.Request(
				tut.ReqWithMethod(http.MethodPost),
				testutils.WithCookieAccessToken(
					priv,
				),
				tut.ReqWithCookies(csrfCookie),
				tut.ReqWithHeader(nosurf.HeaderName, csrfToken),
			),
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusOK),
			),
		},
		{
			name:       "should fail with a valid jwt but option disabled - POST and CSRF protection",
			middleware: auth.Verify(),
			request: tut.Request(
				tut.ReqWithMethod(http.MethodPost),
				testutils.WithCookieAccessToken(
					priv,
				),
				tut.ReqWithCookies(csrfCookie),
				tut.ReqWithHeader(nosurf.HeaderName, csrfToken),
			),
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name:       "should fail with a valid jwt but option disabled - GET",
			middleware: auth.Verify(),
			request: tut.Request(
				testutils.WithCookieAccessToken(
					priv,
				),
			),
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name:       "should fail with a valid jwt in cookie but missing CSRF cookie",
			middleware: authAcceptingCookie.Verify(),
			request: tut.Request(
				tut.ReqWithMethod(http.MethodPost),
				testutils.WithCookieAccessToken(
					priv,
				),
				tut.ReqWithHeader(nosurf.HeaderName, csrfToken),
			),
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name:       "should fail with a valid jwt in cookie but missing CSRF header",
			middleware: authAcceptingCookie.Verify(),
			request: tut.Request(
				tut.ReqWithMethod(http.MethodPost),
				testutils.WithCookieAccessToken(
					priv,
				),
				tut.ReqWithCookies(csrfCookie),
			),
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name:       "should work if at least one valid token is found - valid token in header",
			middleware: authAcceptingCookie.Verify(),
			request: tut.Request(
				tut.ReqWithMethod(http.MethodGet),
				tut.ReqWithAuthHeader(
					ownerUUID,
					priv,
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
				tut.ReqWithFormValue(jwt.AccessTokenArgumentName, "wrong"),
				tut.ReqWithCookies(&http.Cookie{
					Name:  jwt.AccessCookieName,
					Value: "wrong",
				}),
			),
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusOK),
			),
		},
		{
			name:       "should work if at least one valid token is found - valid token in form argument",
			middleware: authAcceptingCookie.Verify(),
			request: tut.Request(
				tut.ReqWithMethod(http.MethodGet),
				// no invalid token in header can be included - that would shadow the token in the form
				testutils.WithFormAccessToken(
					priv,
				),
				tut.ReqWithCookies(&http.Cookie{
					Name:  jwt.AccessCookieName,
					Value: "wrong",
				}),
			),
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusOK),
			),
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			handler := tc.middleware(http.HandlerFunc(testutils.OkHandler))
			res := httptest.NewRecorder()

			handler.ServeHTTP(res, tc.request)

			if err := tc.checks(res.Result()); err != nil {
				t.Error(err)
				return
			}
		})
	}
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
  not_before: "1410-01-01"
  not_after: "2099-01-01"
  key: |
    ` + string(testutils.GeneratePEMPublicKey(t, &priv1.PublicKey, 4)) + `
- name: "key2"
  comment: "valid test key2"
  not_before: "1410-01-01"
  not_after: "2099-01-01"
  key: |
    ` + string(testutils.GeneratePEMPublicKey(t, &priv2.PublicKey, 4)) + `
- name: "expiredKey3"
  comment: "valid key but metadata says it should not be used"
  not_before: "1999-01-01"
  not_after: "1999-01-02"
  key: |
    ` + string(testutils.GeneratePEMPublicKey(t, &priv3.PublicKey, 4)) + `
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

	for _, tc := range []struct {
		name       string
		request    *http.Request
		middleware func(http.Handler) http.Handler
		checks     tut.ResponseCheckFunc
		endHandler func(http.ResponseWriter, *http.Request)
	}{
		{
			name:       "should succeed with key1",
			middleware: auth.Verify(),
			request: tut.Request(
				tut.ReqWithAuthHeader(
					ownerUUID,
					priv1,
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
			),
			endHandler: testutils.OkHandler,
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusOK),
			),
		},
		{
			name:       "should succeed with key2",
			middleware: auth.Verify(),
			request: tut.Request(
				tut.ReqWithAuthHeader(
					ownerUUID,
					priv2,
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
			),
			endHandler: testutils.OkHandler,
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusOK),
			),
		},
		{
			name:       "should ignore metadata and work with key3", // TODO-PR: Change this case when handling of metadata will be implemented
			middleware: auth.Verify(),
			request: tut.Request(
				tut.ReqWithAuthHeader(
					ownerUUID,
					priv3,
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
			),
			endHandler: testutils.OkHandler,
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusOK),
			),
		},
		{
			name:       "should fail with all 3 keys not matching the private key",
			middleware: auth.Verify(),
			request: tut.Request(
				tut.ReqWithAuthHeader(
					ownerUUID,
					priv4,
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
			),
			endHandler: testutils.OkHandler,
			checks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusUnauthorized),
			),
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			handler := tc.middleware(http.HandlerFunc(tc.endHandler))
			res := httptest.NewRecorder()

			handler.ServeHTTP(res, tc.request)

			if err := tc.checks(res.Result()); err != nil {
				t.Error(err)
				return
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
	pkp := testutils.DummyKeyProvider{Key: &priv.PublicKey}
	l := testutils.Logger{}
	auth := jwt.NewAuthenticator(&pkp, &l)
	authAcceptingCookie := jwt.NewAuthenticatorWithOptions(&pkp, &l, jwt.AcceptAccessCookie)

	userID := uuid.Must(uuid.NewV4())
	clientID := uuid.Must(uuid.NewV4())
	tenantID := "some-tenant"

	csrfCookie, csrfToken := tut.CSRFValues()

	for _, tc := range [...]struct {
		name           string
		middleware     func(http.Handler) http.Handler
		request        *http.Request
		reqChecks      tut.RequestCheckFunc
		responseChecks tut.ResponseCheckFunc
	}{
		{
			name:       "should succeed with request with valid JWT in header",
			middleware: auth.Extract,
			request: tut.Request(
				tut.ReqWithAuthHeader(
					userID,
					priv,
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
					jwt.WithClientID(clientID.String()),
					jwt.WithTenantID(tenantID),
				),
			),
			reqChecks: tut.CheckRequest(
				tut.ReqHasInContext(d4lcontext.ClientIDContextKey, clientID.String()),
				tut.ReqHasInContext(d4lcontext.UserIDContextKey, userID),
				tut.ReqHasInContext(d4lcontext.TenantIDContextKey, tenantID),
				tut.ReqHasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetSubjectID(ctx)
				}, userID),
				tut.ReqHasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetTenantID(ctx)
				}, tenantID),
				tut.ReqHasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetClientID(ctx)
				}, clientID.String()),
			),
			responseChecks: tut.RespHasStatusCode(http.StatusOK),
		},
		{
			name:       "should succeed with request with valid JWT in form parameter",
			middleware: auth.Extract,
			request: tut.Request(
				testutils.WithFormAccessToken(
					priv,
					jwt.WithUserID(userID),
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
					jwt.WithClientID(clientID.String()),
					jwt.WithTenantID(tenantID),
				),
			),
			reqChecks: tut.CheckRequest(
				tut.ReqHasInContext(d4lcontext.ClientIDContextKey, clientID.String()),
				tut.ReqHasInContext(d4lcontext.UserIDContextKey, userID),
				tut.ReqHasInContext(d4lcontext.TenantIDContextKey, tenantID),
				tut.ReqHasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetSubjectID(ctx)
				}, userID),
				tut.ReqHasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetTenantID(ctx)
				}, tenantID),
				tut.ReqHasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetClientID(ctx)
				}, clientID.String()),
			),
			responseChecks: tut.RespHasStatusCode(http.StatusOK),
		},
		{
			name:       "should succeed with request with valid JWT in access cookie and valid CSRF",
			middleware: authAcceptingCookie.Extract,
			request: tut.Request(
				tut.ReqWithMethod(http.MethodPost),
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithUserID(userID),
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
					jwt.WithClientID(clientID.String()),
					jwt.WithTenantID(tenantID),
				),
				tut.ReqWithCookies(csrfCookie),
				tut.ReqWithHeader(nosurf.HeaderName, csrfToken),
			),
			reqChecks: tut.CheckRequest(
				tut.ReqHasInContext(d4lcontext.ClientIDContextKey, clientID.String()),
				tut.ReqHasInContext(d4lcontext.UserIDContextKey, userID),
				tut.ReqHasInContext(d4lcontext.TenantIDContextKey, tenantID),
				tut.ReqHasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetSubjectID(ctx)
				}, userID),
				tut.ReqHasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetTenantID(ctx)
				}, tenantID),
				tut.ReqHasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetClientID(ctx)
				}, clientID.String()),
			),
			responseChecks: tut.RespHasStatusCode(http.StatusOK),
		},
		{
			name:       "should not extract a valid access cookie if cookie option is not enabled",
			middleware: auth.Extract,
			request: tut.Request(
				tut.ReqWithMethod(http.MethodPost),
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithUserID(userID),
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
					jwt.WithClientID(clientID.String()),
					jwt.WithTenantID(tenantID),
				),
				tut.ReqWithCookies(csrfCookie),
				tut.ReqWithHeader(nosurf.HeaderName, csrfToken),
			),
			reqChecks: tut.CheckRequest(
				tut.ReqHasInContext(d4lcontext.ClientIDContextKey, nil),
				tut.ReqHasInContext(d4lcontext.UserIDContextKey, nil),
				tut.ReqHasInContext(d4lcontext.TenantIDContextKey, nil),
			),
			responseChecks: tut.RespHasStatusCode(http.StatusOK),
		},
		{
			name:       "should succeed with request with valid JWT in access cookie and CSRF-safe method",
			middleware: authAcceptingCookie.Extract,
			request: tut.Request(
				tut.ReqWithMethod(http.MethodGet),
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithUserID(userID),
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
					jwt.WithClientID(clientID.String()),
					jwt.WithTenantID(tenantID),
				),
			),
			reqChecks: tut.CheckRequest(
				tut.ReqHasInContext(d4lcontext.ClientIDContextKey, clientID.String()),
				tut.ReqHasInContext(d4lcontext.UserIDContextKey, userID),
				tut.ReqHasInContext(d4lcontext.TenantIDContextKey, tenantID),
				tut.ReqHasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetSubjectID(ctx)
				}, userID),
				tut.ReqHasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetTenantID(ctx)
				}, tenantID),
				tut.ReqHasInContextExtract(func(ctx context.Context) (interface{}, error) {
					return jwt.GetClientID(ctx)
				}, clientID.String()),
			),
			responseChecks: tut.RespHasStatusCode(http.StatusOK),
		},
		{
			name:       "should not extract a valid access cookie with missing CSRF cookie",
			middleware: authAcceptingCookie.Extract,
			request: tut.Request(
				tut.ReqWithMethod(http.MethodPost),
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithUserID(userID),
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
					jwt.WithClientID(clientID.String()),
					jwt.WithTenantID(tenantID),
				),
				tut.ReqWithHeader(nosurf.HeaderName, csrfToken),
			),
			reqChecks: tut.CheckRequest(
				tut.ReqHasInContext(d4lcontext.ClientIDContextKey, nil),
				tut.ReqHasInContext(d4lcontext.UserIDContextKey, nil),
				tut.ReqHasInContext(d4lcontext.TenantIDContextKey, nil),
			),
			responseChecks: tut.RespHasStatusCode(http.StatusOK),
		},
		{
			name:       "should not extract a valid access cookie with missing CSRF header",
			middleware: authAcceptingCookie.Extract,
			request: tut.Request(
				tut.ReqWithMethod(http.MethodPost),
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithUserID(userID),
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
					jwt.WithClientID(clientID.String()),
					jwt.WithTenantID(tenantID),
				),
				tut.ReqWithCookies(csrfCookie),
			),
			reqChecks: tut.CheckRequest(
				tut.ReqHasInContext(d4lcontext.ClientIDContextKey, nil),
				tut.ReqHasInContext(d4lcontext.UserIDContextKey, nil),
				tut.ReqHasInContext(d4lcontext.TenantIDContextKey, nil),
			),
			responseChecks: tut.RespHasStatusCode(http.StatusOK),
		},
		{
			name:           "should not break the middleware chain with a request without a JWT",
			middleware:     auth.Extract,
			request:        tut.Request(),
			responseChecks: tut.RespHasStatusCode(http.StatusOK),
		},
		{
			name:       "should not break the middleware chain with a request with an unknown scope",
			middleware: auth.Extract,
			request: tut.Request(
				tut.ReqWithAuthHeader(
					userID,
					priv,
					jwt.WithScopeStrings("unknown"),
					jwt.WithClientID(clientID.String()),
					jwt.WithTenantID(tenantID),
				),
			),
			responseChecks: tut.RespHasStatusCode(http.StatusOK),
		},
		{
			name:       "should not break the middleware chain with a JWT with missing data",
			middleware: auth.Extract,
			request: tut.Request(
				tut.ReqWithAuthHeader(
					uuid.Nil,
					priv,
					jwt.WithScopeStrings(""),
					jwt.WithClientID(""),
					jwt.WithTenantID(""),
				),
			),
			responseChecks: tut.RespHasStatusCode(http.StatusOK),
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
				if err := tc.responseChecks(respRecorder.Result()); err != nil {
					t.Error(err)
				}
			}
		})
	}
}
