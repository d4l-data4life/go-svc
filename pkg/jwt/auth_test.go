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
	"github.com/justinas/nosurf"

	"github.com/gofrs/uuid"
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

	csrfCookie, csrfToken := testutils.CSRFValues()

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
				func(r *http.Request) {
					r.Header.Add("Authorization", "wrong")
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
				func(r *http.Request) {
					r.Header.Add("Authorization", "Bearer wrong")
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
				func(r *http.Request) {
					r.Form = url.Values{}
					r.Form.Add("access_token", "wrong")
					r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
			name:       "should work with a valid jwt in cookie and CSRF safe method",
			middleware: authAcceptingCookie.Verify(),
			request: testutils.BuildRequest(
				testutils.WithMethod(http.MethodGet),
				testutils.WithCookieAccessToken(
					priv,
				),
			),
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name:       "should work with a valid jwt in cookie with CSRF protection",
			middleware: authAcceptingCookie.Verify(),
			request: testutils.BuildRequest(
				testutils.WithMethod(http.MethodPost),
				testutils.WithCookieAccessToken(
					priv,
				),
				testutils.WithCookie(csrfCookie),
				testutils.WithHeader(map[string]string{nosurf.HeaderName: csrfToken}),
			),
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name:       "should fail with a valid jwt but option disabled - POST and CSRF protection",
			middleware: auth.Verify(),
			request: testutils.BuildRequest(
				testutils.WithMethod(http.MethodPost),
				testutils.WithCookieAccessToken(
					priv,
				),
				testutils.WithCookie(csrfCookie),
				testutils.WithHeader(map[string]string{nosurf.HeaderName: csrfToken}),
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name:       "should fail with a valid jwt but option disabled - GET",
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
		{
			name:       "should fail with a valid jwt in cookie but missing CSRF cookie",
			middleware: authAcceptingCookie.Verify(),
			request: testutils.BuildRequest(
				testutils.WithMethod(http.MethodPost),
				testutils.WithCookieAccessToken(
					priv,
				),
				testutils.WithHeader(map[string]string{nosurf.HeaderName: csrfToken}),
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name:       "should fail with a valid jwt in cookie but missing CSRF header",
			middleware: authAcceptingCookie.Verify(),
			request: testutils.BuildRequest(
				testutils.WithMethod(http.MethodPost),
				testutils.WithCookieAccessToken(
					priv,
				),
				testutils.WithCookie(csrfCookie),
			),
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name:       "should work if at least one valid token is found - valid token in header",
			middleware: authAcceptingCookie.Verify(),
			request: testutils.BuildRequest(
				testutils.WithMethod(http.MethodGet),
				testutils.WithAuthHeader(
					priv,
					jwt.WithUserID(ownerUUID),
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
				),
				testutils.WithForm(
					testutils.WithValue(jwt.AccessTokenArgumentName, "wrong"),
				),
				testutils.WithCookie(&http.Cookie{
					Name:  jwt.AccessCookieName,
					Value: "wrong",
				}),
			),
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name:       "should work if at least one valid token is found - valid token in form argument",
			middleware: authAcceptingCookie.Verify(),
			request: testutils.BuildRequest(
				testutils.WithMethod(http.MethodGet),
				// no invalid token in header can be included - that would shadow the token in the form
				testutils.WithFormAccessToken(
					priv,
				),
				testutils.WithCookie(&http.Cookie{
					Name:  jwt.AccessCookieName,
					Value: "wrong",
				}),
			),
			checks: checks(
				hasStatusCode(http.StatusOK),
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
    ` + string(testutils.GeneratePEMPublicKey(t, &priv1.PublicKey, 4)) + `
- name: "key2"
  comment: "valid test key2"
  not_before: 1410-01-01
  not_after: 2099-01-01
  key: |
    ` + string(testutils.GeneratePEMPublicKey(t, &priv2.PublicKey, 4)) + `
- name: "expiredKey3"
  comment: "valid key but metadata says it should not be used"
  not_before: 1999-01-01
  not_after: 1999-01-02
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

	for _, tc := range [...]testData{
		{
			name:       "should succeed with key1",
			middleware: auth.Verify(),
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
			name:       "should succeed with key2",
			middleware: auth.Verify(),
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
			name:       "should ignore metadata and work with key3", // TODO-PR: Change this case when handling of metadata will be implemented
			middleware: auth.Verify(),
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
			name:       "should fail with all 3 keys not matching the private key",
			middleware: auth.Verify(),
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
	pkp := testutils.DummyKeyProvider{Key: &priv.PublicKey}
	l := testutils.Logger{}
	auth := jwt.NewAuthenticator(&pkp, &l)
	authAcceptingCookie := jwt.NewAuthenticatorWithOptions(&pkp, &l, jwt.AcceptAccessCookie)

	userID := uuid.Must(uuid.NewV4())
	clientID := uuid.Must(uuid.NewV4())
	tenantID := "some-tenant"

	csrfCookie, csrfToken := testutils.CSRFValues()

	for _, tc := range [...]struct {
		name           string
		middleware     func(http.Handler) http.Handler
		request        *http.Request
		reqChecks      checkReqFunc
		responseChecks checkFunc
	}{
		{
			name:       "should succeed with request with valid JWT in header",
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
			responseChecks: hasStatusCode(http.StatusOK),
		},
		{
			name:       "should succeed with request with valid JWT in form parameter",
			middleware: auth.Extract,
			request: testutils.BuildRequest(
				testutils.WithFormAccessToken(
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
			responseChecks: hasStatusCode(http.StatusOK),
		},
		{
			name:       "should succeed with request with valid JWT in access cookie and valid CSRF",
			middleware: authAcceptingCookie.Extract,
			request: testutils.BuildRequest(
				testutils.WithMethod(http.MethodPost),
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithUserID(userID),
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
					jwt.WithClientID(clientID.String()),
					jwt.WithTenantID(tenantID),
				),
				testutils.WithCookie(csrfCookie),
				testutils.WithHeader(map[string]string{nosurf.HeaderName: csrfToken}),
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
			responseChecks: hasStatusCode(http.StatusOK),
		},
		{
			name:       "should not extract a valid access cookie if cookie option is not enabled",
			middleware: auth.Extract,
			request: testutils.BuildRequest(
				testutils.WithMethod(http.MethodPost),
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithUserID(userID),
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
					jwt.WithClientID(clientID.String()),
					jwt.WithTenantID(tenantID),
				),
				testutils.WithCookie(csrfCookie),
				testutils.WithHeader(map[string]string{nosurf.HeaderName: csrfToken}),
			),
			reqChecks: checkReqAll(
				hasInContext(d4lcontext.ClientIDContextKey, nil),
				hasInContext(d4lcontext.UserIDContextKey, nil),
				hasInContext(d4lcontext.TenantIDContextKey, nil),
			),
			responseChecks: hasStatusCode(http.StatusOK),
		},
		{
			name:       "should succeed with request with valid JWT in access cookie and CSRF-safe method",
			middleware: authAcceptingCookie.Extract,
			request: testutils.BuildRequest(
				testutils.WithMethod(http.MethodGet),
				testutils.WithCookieAccessToken(
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
			responseChecks: hasStatusCode(http.StatusOK),
		},
		{
			name:       "should not extract a valid access cookie with missing CSRF cookie",
			middleware: authAcceptingCookie.Extract,
			request: testutils.BuildRequest(
				testutils.WithMethod(http.MethodPost),
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithUserID(userID),
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
					jwt.WithClientID(clientID.String()),
					jwt.WithTenantID(tenantID),
				),
				testutils.WithHeader(map[string]string{nosurf.HeaderName: csrfToken}),
			),
			reqChecks: checkReqAll(
				hasInContext(d4lcontext.ClientIDContextKey, nil),
				hasInContext(d4lcontext.UserIDContextKey, nil),
				hasInContext(d4lcontext.TenantIDContextKey, nil),
			),
			responseChecks: hasStatusCode(http.StatusOK),
		},
		{
			name:       "should not extract a valid access cookie with missing CSRF header",
			middleware: authAcceptingCookie.Extract,
			request: testutils.BuildRequest(
				testutils.WithMethod(http.MethodPost),
				testutils.WithCookieAccessToken(
					priv,
					jwt.WithUserID(userID),
					jwt.WithScopeStrings(jwt.TokenAttachmentsWrite),
					jwt.WithClientID(clientID.String()),
					jwt.WithTenantID(tenantID),
				),
				testutils.WithCookie(csrfCookie),
			),
			reqChecks: checkReqAll(
				hasInContext(d4lcontext.ClientIDContextKey, nil),
				hasInContext(d4lcontext.UserIDContextKey, nil),
				hasInContext(d4lcontext.TenantIDContextKey, nil),
			),
			responseChecks: hasStatusCode(http.StatusOK),
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
