package middlewares

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/dynamic"
	"github.com/gesundheitscloud/go-svc/pkg/instrumented"

	"github.com/golang-jwt/jwt/v4"

	"github.com/stretchr/testify/assert"
)

func TestServiceSecret(t *testing.T) {
	validAuthHeader := "service-secret"

	handlerFactory := instrumented.NewHandlerFactory("d4l", instrumented.DefaultInstrumentInitOptions, instrumented.DefaultInstrumentOptions)
	auth := NewAuth(validAuthHeader, nil, handlerFactory)
	tests := []struct {
		name              string
		AuthHeaderContent string
		expectedStatus    int
	}{
		{"valid Service Auth", validAuthHeader, http.StatusOK},
		{"invalid Service Auth", "random", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/route", nil)
			req.Header.Add(AuthHeaderName, tt.AuthHeaderContent)
			res := httptest.NewRecorder()

			// target handler after auth check
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			authMiddleware := auth.ServiceSecret(handler)
			authMiddleware.ServeHTTP(res, req)
			assert.Equal(t, tt.expectedStatus, res.Code)
		})
	}
}

func TestJWTNewAuthentication(t *testing.T) {
	vc := getViperConfig("unit-test", t)
	assert.NoError(t, vc.Error)
	pk, err := vc.JWTPrivateKey()
	assert.NoError(t, err)
	privateKey := pk.Key

	claims := &jwt.RegisteredClaims{
		Issuer: "test",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	assert.NoError(t, err)
	_ = signedToken

	emptyVc := getEmptyViperConfig("unit-test", t)
	assert.NoError(t, vc.Error)

	tests := []struct {
		name              string
		authOption        AuthOptionJWTKeys
		AuthHeaderContent string
		expectedStatus    int
	}{
		{name: "JWT keys provider - happy path Service Auth",
			authOption:        AuthWithPublicKeyProvider(vc),
			AuthHeaderContent: signedToken,
			expectedStatus:    http.StatusOK,
		},
		{name: "JWT keys provider - invalid auth header",
			authOption:        AuthWithPublicKeyProvider(vc),
			AuthHeaderContent: "random",
			expectedStatus:    http.StatusUnauthorized,
		},
		{name: "JWT keys provider - public keys provider unavailable",
			authOption:        nil,
			AuthHeaderContent: signedToken,
			expectedStatus:    http.StatusInternalServerError,
		},
		{name: "JWT keys provider - empty viper config",
			authOption:        AuthWithPublicKeyProvider(emptyVc),
			AuthHeaderContent: signedToken,
			expectedStatus:    http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			handlerFactory := instrumented.NewHandlerFactory("d4l", instrumented.DefaultInstrumentInitOptions, instrumented.DefaultInstrumentOptions)
			options := []AuthOption{AuthWithLatencyBuckets([]float64{4, 8, 16}), AuthWithSizeBuckets([]float64{4, 8, 16})}
			auth := NewAuthentication("", tt.authOption, handlerFactory, options...)

			req, _ := http.NewRequest(http.MethodGet, "/route", nil)
			req.Header.Add(AuthHeaderName, fmt.Sprintf("Bearer %s", tt.AuthHeaderContent))
			res := httptest.NewRecorder()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			authMiddleware := auth.JWT(handler)
			authMiddleware.ServeHTTP(res, req)
			assert.Equal(t, tt.expectedStatus, res.Code)
		})
	}
}

// TestJWTNewAuth tests the deprectated constructor middlewares.NewAuth
func TestJWTNewAuth(t *testing.T) {
	vc := getViperConfig("unit-test", t)
	assert.NoError(t, vc.Error)
	pk, err := vc.JWTPrivateKey()
	assert.NoError(t, err)
	privateKey := pk.Key

	claims := &jwt.RegisteredClaims{
		Issuer: "test",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, _ := token.SignedString(privateKey)

	tests := []struct {
		name              string
		key               *rsa.PublicKey
		AuthHeaderContent string
		expectedStatus    int
	}{
		{name: "JWT key from file - happy path Service Auth",
			key:               &privateKey.PublicKey,
			AuthHeaderContent: signedToken,
			expectedStatus:    http.StatusOK,
		},
		{name: "JWT key from file - invalid auth header",
			key:               &privateKey.PublicKey,
			AuthHeaderContent: "random",
			expectedStatus:    http.StatusUnauthorized,
		},
		{name: "JWT key from file - key is nil",
			key:               nil,
			AuthHeaderContent: signedToken,
			expectedStatus:    http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			handlerFactory := instrumented.NewHandlerFactory("d4l", instrumented.DefaultInstrumentInitOptions, instrumented.DefaultInstrumentOptions)
			options := []AuthOption{AuthWithLatencyBuckets([]float64{4, 8, 16}), AuthWithSizeBuckets([]float64{4, 8, 16})}

			auth := NewAuth("", tt.key, handlerFactory, options...)
			assert.Equal(t, tt.key, auth.publicKey)

			req, _ := http.NewRequest(http.MethodGet, "/route", nil)
			req.Header.Add(AuthHeaderName, fmt.Sprintf("Bearer %s", tt.AuthHeaderContent))
			res := httptest.NewRecorder()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			authMiddleware := auth.JWT(handler)
			authMiddleware.ServeHTTP(res, req)
			assert.Equal(t, tt.expectedStatus, res.Code)
		})
	}
}

func TestGetAuthSecret(t *testing.T) {
	tests := []struct {
		name              string
		authHeaderContent string
		expectedSecret    string
	}{
		{"Service secret without prefix", "secret", "secret"},
		{"Service secret with prefix", "Bearer anothersecret", "anothersecret"},
	}

	auth := &Auth{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/route", nil)
			req.Header.Add(AuthHeaderName, tt.authHeaderContent)

			authToken, _ := auth.getAuthSecret(req)
			assert.Equal(t, authToken, tt.expectedSecret)
		})
	}
}

// getEmptyViperConfig provides a valid (but empty) ViperConfig for test purposes
func getEmptyViperConfig(name string, t *testing.T) *dynamic.ViperConfig {
	// important - watch indentation here! this must produce valid yaml
	var yamlExample = []byte(`
JWTPublicKey: []
JWTPrivateKey: []
`)
	vc := dynamic.NewViperConfig(name,
		dynamic.WithConfigFormat("yaml"),
		dynamic.WithConfigSource(bytes.NewBuffer(yamlExample)),
		dynamic.WithAutoBootstrap(false),
		dynamic.WithWatchChanges(false),
	)
	err := vc.Bootstrap()
	assert.NoError(t, err, "viperConfig bootstrap error")
	return vc
}

// getViperConfig provides a valid ViperConfig for test purposes
func getViperConfig(name string, t *testing.T) *dynamic.ViperConfig {
	// important - watch indentation here! this must produce valid yaml
	var yamlExample = []byte(`
JWTPublicKey:
` + fixturePubKeyEntry("public") + `
JWTPrivateKey:
` + fixturePrivKeyEntry("private", true) + `
`)

	vc := dynamic.NewViperConfig(name,
		dynamic.WithConfigFormat("yaml"),
		dynamic.WithConfigSource(bytes.NewBuffer(yamlExample)),
		dynamic.WithAutoBootstrap(false),
		dynamic.WithWatchChanges(false),
	)
	err := vc.Bootstrap()
	assert.NoError(t, err, "viperConfig bootstrap error")

	pubK, err := vc.JWTPublicKeys()
	assert.NoError(t, err)
	// we expect 2 pub keys, not 3, because one (name prefixed with "broken-") cannot be parsed
	// see internal/viperJWT/testing.go
	assert.Len(t, pubK, 2, "not all public keys were read")

	privK, err := vc.JWTPrivateKey()
	assert.NoError(t, err)
	assert.NotEqual(t, rsa.PrivateKey{}, privK.Key)
	assert.NotNil(t, privK.Key.PublicKey, "vega requires the JWT pubKey to be included in the JWT privKey")
	return vc
}

// THESE KEYS WILL BE USED IN TESTS - they shall be treated as test fixtures
// fixturePrivKeyEntry is test helper to generate a valid priv Key entry
func fixturePrivKeyEntry(name string, enabled bool) string {
	return `
- name: "` + name + `"
  comment: "generated with: openssl genrsa -out private.pem 2048"
  enabled: ` + fmt.Sprintf("%t", enabled) + `
  created_at: "unknown"
  author: "Jenkins"
  key: |
    -----BEGIN RSA PRIVATE KEY-----
    MIIEpAIBAAKCAQEAterVQa0ygpUkQXdvKtXClkocx2SK2fDzzENsYmq7fbLS4RQy
    L8YFygz8CADvZjzcv4QokU7tAQLu948Cwd7lXLmPLIqRBgMSb+BkAOPurvP079Wz
    7x6oolP/7B5bo4Q75C5gTxmvuDcDKjjGKna+FbUFqk+DrY/lZWPsxjj+NbcCXBvt
    t15+uUVDV2agCkKL6GGc7pbq468oTwKriiv9EFBuKRn4ocbOfraZnJSiLA7+89vk
    0H8aCzyfF1jV8+wzoKl9MHJyr5M+moHYNpBZzxPCAHinn5UHf6zlPAwPPt4roZ55
    XNWVGuWfbWTsmfYJYma1RC38emHJ8Ihs3gHwnwIDAQABAoIBAQCK++0ODlrmtTdL
    5QnDuii+VcUC+Wez9ojs6B4oWs7/y92dJKbrJOlLYvwyyTQd8iXdFAVCbwBXo3wb
    GuHKaJbnbsVaDEucQkCVxOPiYkH63Fun2Kdt6wh/bJm8Nb1hgieXv27JQCCmJzF9
    0n5j9vBm+TRo1/MMaUGjYuKE1wow0mVvHj83KExGPT8AgcOix4Je7cdZvIcsQazg
    +Ow2Cc4qDLzEv5pwBNlr9KJJGGQz21mv1tsuzwWOqpnQSPktykULDXtkivXigNWF
    NqCERvDmjZsv5xXzjY0yleRSF85gpM6NjqfdUt6lyE2lp3WLSQSSZd6d8GKthNZN
    VB2/ja6RAoGBANpnUlLYZ3162hG8msqrLhN1kF2Od9kbmlbgcVB62X07vMy3dWcj
    sjiKkYIIjnEXwZP8Rz18dEJkSotQw8PCfWL0KrxD20W3nWJsWYvjKW57QvL1W9BI
    5lgivuF0m+DG8/OejsZOcYmb1SvXuNHAJ1fPTffPhCUfhHSskKZIWcnLAoGBANU7
    n1nJmxwK53hrNlKx0gIFylVk02esXYmP/x9To9VZODqy6VcYdhHvnVhMRv7W2Syy
    ZBOPA0ijJqz7jB25SK3mVDPHaHul0vMQrQ9A3ZpNOkxZ+MQE3oRVJxRlV7LmIaqj
    2ANo5QLp5bimLl0eojPSdoO+MPM1zyalUHKFXCn9AoGAYo1/G30lbfzyzFAkNVH7
    T7KcO2tfb2vCQHO1DlDxNU6wilw6sRjtghAdSuUbibLjmiib6QXw3EivTqBaRkrM
    E8wEQMIJ/zK05UXpPnN6La2Xb1UCYkGTF7BOHCRndo2wZX1mBdw95Y+ZKNiGQLgJ
    yNj14N4WTj4johaAi1hYk/MCgYEAnzMum/i7h8pUW0Ggg0kkBEKSeAMZG1RDWcta
    rObjcQx1wM2HDXHD5UxC64O3ldiOuKJPuZKS3w6Ad7IvQJnvO3a18xq0VWzO/I68
    xqClUujJ1+tsod0IzUBONxoaygSrqh09z/3mzbAXxS69euS+MXa26VF8dnj8Olw+
    neIXl3kCgYBLOBghFBRFdQQ9tR/PJ4gT7PguWaMg2pY8lKGX3sgY2RWVdspEZK6y
    1O7YfUFHkQR7FADcKgKq6Udi1x2R9qFX/r6cX++dB3SxOK5RB7nP+VQNycRrcB7n
    WUWcKJhUbocP74i01EfIh7c4sxqZGR0t/jlGIvJAf9vL07cR1nW1Fg==
    -----END RSA PRIVATE KEY-----`
}

// fixturePubKeyEntry is test helper to generate a valid config containing pub Key entries
// one of them matches the private key from fixturePrivKeyEntry
func fixturePubKeyEntry(name string) string {
	return `
- name: "` + name + `"
  comment: "generated with: openssl rsa -in private.pem -pubout -outform PEM -out public.pem"
  not_before: "2020-01-01"
  not_after: "2022-01-01"
  key: |
    -----BEGIN PUBLIC KEY-----
    MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAterVQa0ygpUkQXdvKtXC
    lkocx2SK2fDzzENsYmq7fbLS4RQyL8YFygz8CADvZjzcv4QokU7tAQLu948Cwd7l
    XLmPLIqRBgMSb+BkAOPurvP079Wz7x6oolP/7B5bo4Q75C5gTxmvuDcDKjjGKna+
    FbUFqk+DrY/lZWPsxjj+NbcCXBvtt15+uUVDV2agCkKL6GGc7pbq468oTwKriiv9
    EFBuKRn4ocbOfraZnJSiLA7+89vk0H8aCzyfF1jV8+wzoKl9MHJyr5M+moHYNpBZ
    zxPCAHinn5UHf6zlPAwPPt4roZ55XNWVGuWfbWTsmfYJYma1RC38emHJ8Ihs3gHw
    nwIDAQAB
    -----END PUBLIC KEY-----
- name: "broken-` + name + `"
  comment: "broken, not parsable key"
  not_before: "2020-01-01"
  not_after: "2022-01-01"
  key: |
    -----BEGIN PUBLIC KEY-----
    gibberish
    -----END PUBLIC KEY-----
- name: "other-` + name + `"
  comment: "not matching any private key, but fully valid public key"
  not_before: "2020-01-01"
  not_after: "2022-01-01"
  key: |
    -----BEGIN PUBLIC KEY-----
    MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAw7NCc5g910XRqfgIfue0
    bQJ/EV62UTZSkimn+jkpH/BKZmIZCc+k9Fh4n90rByBJ5GHOs2KRfGtrpofiymbb
    jH8dS3UK2iYK43B4hEFV3I/E8BYXEfVdfpRiblWHzQ9ZRlwzAnEAFHpxxenzpLwc
    5og5+f55l2JzL49hyE5tyP2SENVMFyS/XuVdlpLadnAVbwJHVf3T3TBxzBfiXOQt
    QvsbuP9uQGafh2Q4nv933/kVGyqSvpnrBKpY5ux/YBfeR70QyauR+4rqE0R6wId4
    BM5gcL6xDNIjmLPjV7DiCA32ySX0ruJZY7jRv37evqpap9qnh9RGlZDK1ruMCDua
    fQIDAQAB
    -----END PUBLIC KEY-----`
}
