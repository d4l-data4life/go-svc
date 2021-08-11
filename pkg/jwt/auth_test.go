package jwt

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/d4lcontext"
	"github.com/gesundheitscloud/go-svc/pkg/dynamic"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
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
					WithUserID(uuid.Must(uuid.FromString(ownerUUID))),
					WithScopeStrings(TokenAttachmentsWrite),
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
					WithUserID(uuid.Must(uuid.FromString(ownerUUID))),
					WithScopeStrings(TokenAttachmentsWrite),
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
					WithUserID(uuid.Must(uuid.FromString(ownerUUID))),
					WithScopeStrings(TokenAttachmentsWrite),
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
					WithUserID(uuid.Must(uuid.FromString(ownerUUID))),
					WithScopeStrings(TokenAttachmentsWrite),
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
					WithUserID(uuid.Must(uuid.FromString(ownerUUID))),
					WithScopeStrings(TokenAttachmentsWrite),
				),
			),
			endHandler: okHandler,
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
					WithUserID(uuid.Must(uuid.FromString(ownerUUID))),
					WithScopeStrings(TokenAttachmentsWrite),
				),
			),
			endHandler: okHandler,
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
					WithUserID(uuid.Must(uuid.FromString(otherUUID))),
					WithScopeStrings(TokenAttachmentsWrite),
				),
			),
			endHandler: okHandler,
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
					WithUserID(uuid.Must(uuid.FromString(ownerUUID))),
					WithScopeStrings(TokenAttachmentsRead),
				),
			),
			endHandler: okHandler,
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
			endHandler: okHandler,
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
			request:    &http.Request{},
			endHandler: okHandler,
			checks: checks(
				hasStatusCode(http.StatusUnauthorized),
			),
		},

		{
			name: "end handler should receive a request with JWT claims in the context",
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
					WithUserID(uuid.Must(uuid.FromString(ownerUUID))),
					WithScopeStrings(TokenAttachmentsWrite),
				),
			),
			endHandler: func(w http.ResponseWriter, r *http.Request) {
				claims, ok := r.Context().Value(jwtClaimsContextKey).(*Claims)
				if !ok {
					httpClientError(w, 591) // error - custom codes just for the test to find it easily
					return
				}
				if claims == nil {
					httpClientError(w, 592) // error - custom codes just for the test to find it easily
					return
				}
				httpClientError(w, 299) // success - using 299 to be sure that some other handler won't interfere with standard codes
			},
			checks: checks(
				hasStatusCode(299),
			),
		},
	} {
		tc := tc // Pin Variable

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

	auth := NewAuthenticator(kp, &testLogger{})

	for _, tc := range [...]testData{
		{
			name: "should succeed with key1",
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
					priv1,
					WithUserID(uuid.Must(uuid.FromString(ownerUUID))),
					WithScopeStrings(TokenAttachmentsWrite),
				),
			),
			endHandler: okHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should succeed with key2",
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
					priv2,
					WithUserID(uuid.Must(uuid.FromString(ownerUUID))),
					WithScopeStrings(TokenAttachmentsWrite),
				),
			),
			endHandler: okHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should ignore metadata and work with key3", // TODO-PR: Change this case when handling of metadata will be implemented
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
					priv3,
					WithUserID(uuid.Must(uuid.FromString(ownerUUID))),
					WithScopeStrings(TokenAttachmentsWrite),
				),
			),
			endHandler: okHandler,
			checks: checks(
				hasStatusCode(http.StatusOK),
			),
		},
		{
			name: "should fail with all 3 keys not matching the private key",
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
					priv4,
					WithUserID(uuid.Must(uuid.FromString(ownerUUID))),
					WithScopeStrings(TokenAttachmentsWrite),
				),
			),
			endHandler: okHandler,
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
	auth := New(&priv.PublicKey, &testLogger{})

	userID := uuid.Must(uuid.NewV4())
	clientID := uuid.Must(uuid.NewV4())
	tenantID := "some-tenant"

	for _, tc := range [...]struct {
		name       string
		middleware func(http.Handler) http.Handler
		request    *http.Request
		reqChecks  checkReqFunc
	}{
		{
			name:       "should succeed with request with valid JWT",
			middleware: auth.Extract,
			request: buildRequest(
				withOwnerURL(ownerUUID),
				withAuthHeader(
					priv,
					WithUserID(userID),
					WithScopeStrings(TokenAttachmentsWrite),
					WithClientID(clientID.String()),
					WithTenantID(tenantID),
				),
			),
			reqChecks: checkReqAll(
				hasInContext(d4lcontext.ClientIDContextKey, clientID.String()),
				hasInContext(d4lcontext.UserIDContextKey, userID.String()),
				hasInContext(d4lcontext.TenantIDContextKey, tenantID),
				hasKeyInContext(jwtClaimsContextKey),
			),
		},
		{
			name:       "should not break the middleware chain with a request without a JWT",
			middleware: auth.Extract,
			request: buildRequest(
				withOwnerURL(ownerUUID),
			),
			reqChecks: checkReqAll(),
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

			handler.ServeHTTP(httptest.NewRecorder(), tc.request)

			if !hasBeenCalled {
				t.Fatal(errors.New("handler should have been called, was not."))
			}
			if err := tc.reqChecks(haveReq); err != nil {
				t.Error(err)
			}
		})
	}
}
