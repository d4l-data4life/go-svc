package dynamic_test

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/dynamic"
	"github.com/gesundheitscloud/go-svc/pkg/instrumented"
	"github.com/gesundheitscloud/go-svc/pkg/jwt"
	"github.com/gesundheitscloud/go-svc/pkg/middlewares"
	"github.com/gesundheitscloud/go-svc/pkg/prom"
	"github.com/go-chi/chi"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func serverFun(w http.ResponseWriter, req *http.Request) {
	_, _ = w.Write([]byte("hello"))
}

func mimicHandlerFun(auth *jwt.Authenticator, mws ...func(http.Handler) http.Handler) func(r chi.Router) {
	return func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Use(mws...)
			r.Use(auth.Extract)
			r.Use(auth.Verify(jwt.WithAllScopes(jwt.TokenPermissionsWrite)))
			r.Post("/", serverFun)
		})
	}
}

func getRouterVega(t *testing.T, vc *dynamic.ViperConfig) *chi.Mux {
	logger := newTestLogger(t, t.Name())
	// create authentication middleware and a basic handler - inspired by Vega
	auth := jwt.NewAuthenticator(vc, logger)
	// configure routing as it is done in Vega
	r := chi.NewRouter()
	r.Route("/", mimicHandlerFun(auth))
	return r
}

func getRouterCdsUserdata(t *testing.T, vc *dynamic.ViperConfig) *chi.Mux {
	authMiddleware := middlewares.NewAuthentication(
		"svc-secret",
		middlewares.AuthWithPublicKeyProvider(vc),
		instrumented.NewHandlerFactory("unittest", []prom.InitOption{}, []prom.Option{}),
	)
	mux := chi.NewRouter()
	jwtRoutes := func() *chi.Mux {
		router := chi.NewRouter()
		router.Post("/", serverFun)
		return router
	}
	mux.Route("/", func(r chi.Router) {
		r.With(authMiddleware.JWT).Group(func(r chi.Router) {
			r.Mount("/auth", jwtRoutes())
		})
	})
	return mux
}

func TestViperConfigHandlers(t *testing.T) {
	tests := []struct {
		name           string
		handlerBuilder func(*testing.T, *dynamic.ViperConfig) *chi.Mux
	}{
		{
			name:           "vega-routing",
			handlerBuilder: getRouterVega,
		},
		{
			name:           "cds-userdata-routing",
			handlerBuilder: getRouterCdsUserdata,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// create config.yaml with 4 keys
			tmpDir := generateAndWriteConfig(t, 4, t.Name()+"/pre", "")
			defer func() {
				os.RemoveAll(tmpDir) // clean up the temp dir
			}()
			// build VC object
			logger := newTestLogger(t, t.Name())
			vc := dynamic.NewViperConfig("test"+t.Name(),
				dynamic.WithConfigFormat("yaml"),
				dynamic.WithConfigFileName("config"),
				dynamic.WithConfigFilePaths(tmpDir),
				dynamic.WithAutoBootstrap(true),
				dynamic.WithWatchChanges(true),
				dynamic.WithLogger(logger),
			)
			if vc.Error != nil {
				t.Fatal(vc.Error, "unable to bootstrap VC")
			}
			// sanity check - phase 1 - pre rotation
			t.Logf("======== SANITY CHECK PHASE 1")
			arr, err := vc.JWTPublicKeys()
			assert.NoError(t, err)
			assert.Equal(t, 4, len(arr), "pre-rotation number of keys does not match")

			pkey, err := vc.JWTPrivateKey()
			assert.NoError(t, err)
			assert.Equal(t, t.Name()+"/pre0", pkey.Name, "pre-rotation name of the active key does not match")
			t.Logf("======== SANITY CHECK PHASE 1 DONE")
			// end sanity check

			mux := tt.handlerBuilder(t, vc)
			ts := httptest.NewServer(mux)
			defer ts.Close()
			// generate JWT from the active private key (pre-phase)
			token := generateAccessToken(uuid.Must(uuid.NewV4()), newViperJWTPrivKeyProvider(vc), jwt.WithScopeStrings(allTokens()...))

			t.Logf("======== REQUEST PHASE 1")
			// fire a request to ensure that all 4 pub keys are tested (need to look into logs to see this)
			if resp, body := testRequest(t, ts, "POST", "/auth", token, nil); body != "hello" || resp.StatusCode != 200 {
				t.Fatalf(body, "pre-rotation call should succeed")
			}

			// KEY ROTATION ON DISK IS HAPPENING HERE
			// replace 4 key entries named `pre` with only one key named `post`
			tmpDir = generateAndWriteConfig(t, 1, t.Name()+"/post", tmpDir)
			// wait a bit for the filesystem to catch the changes in the config files and notify Viper
			<-time.After(500 * time.Millisecond)

			// sanity check - phase 2 - post rotation - make sure that viper read everything correctly
			t.Logf("======== SANITY CHECK PHASE 2")
			arr, err = vc.JWTPublicKeys()
			assert.NoError(t, err)
			assert.Equal(t, 1, len(arr), "post-rotation number of keys does not match")
			pkey, err = vc.JWTPrivateKey()
			assert.NoError(t, err)
			assert.Equal(t, t.Name()+"/post0", pkey.Name, "post-rotation name of the active key does not match")
			t.Logf("======== SANITY CHECK PHASE 2 DONE")
			// end sanity check

			t.Logf("======== REQUEST PHASE 2")
			// use the same token for request (from the pre-phase) - it should fail now with 401 as the public key from post-phase does not match the private key from pre-phase
			// pay close attention to logs - there may be many reasons for a 401!
			if resp, body := testRequest(t, ts, "POST", "/auth", token, nil); resp.StatusCode != 401 {
				t.Fatalf(body)
			}
		})
	}
}

func allTokens() []string {
	var tokens []string
	for t := range jwt.KnownTokens {
		tokens = append(tokens, t)
	}
	return tokens
}

type viperJWTPrivKeyProvider struct {
	vc *dynamic.ViperConfig
}

func newViperJWTPrivKeyProvider(vc *dynamic.ViperConfig) *viperJWTPrivKeyProvider {
	return &viperJWTPrivKeyProvider{vc: vc}
}

func (kp *viperJWTPrivKeyProvider) JWTPrivateKey() (*rsa.PrivateKey, error) {
	if kp.vc == nil {
		return nil, fmt.Errorf("viper config nil")
	}
	privKey, err := kp.vc.JWTPrivateKey()
	return privKey.Key, err
}

type JWTPrivateKeyProvider interface {
	JWTPrivateKey() (*rsa.PrivateKey, error)
}

func generateAccessToken(userID uuid.UUID, jwtPrivateKeyProvider JWTPrivateKeyProvider, claimsOptions ...jwt.TokenOption) string {
	options := []jwt.TokenOption{
		jwt.WithUserID(userID),
		jwt.WithAppID(uuid.Must(uuid.NewV4())),
		jwt.WithClientID(uuid.Must(uuid.NewV4()).String()),
		jwt.WithExpirationDuration(time.Minute),
	}
	options = append(options, claimsOptions...)

	privK, err := jwtPrivateKeyProvider.JWTPrivateKey()
	if err != nil {
		panic(fmt.Errorf("could not find jwt private key: %w", err))
	}
	t, err := jwt.CreateAccessToken(privK, options...)
	if err != nil {
		panic(fmt.Errorf("could not sign the token: %w", err))
	}

	return t.AccessToken
}
