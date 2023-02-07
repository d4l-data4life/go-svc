package middlewares

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/tut"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func TestJwtAuthenticationMiddleware(t *testing.T) {
	privKey1, err := rsa.GenerateKey(rand.Reader, 512)

	if err != nil {
		fmt.Printf("failed to generate private key: %s\n", err)
		return
	}

	privKey2, err := rsa.GenerateKey(rand.Reader, 512)

	if err != nil {
		fmt.Printf("failed to generate private key: %s\n", err)
		return
	}
	// This is the key we will use to sign
	realKey, err := jwk.FromRaw(privKey1)
	if err != nil {
		fmt.Printf("failed to create JWK: %s\n", err)
		return
	}
	_ = realKey.Set(jwk.KeyIDKey, "mykey")
	_ = realKey.Set(jwk.AlgorithmKey, jwa.RS256)

	// For demonstration purposes, we also create a bogus key
	bogusKey, err := jwk.FromRaw(privKey2)
	if err != nil {
		fmt.Printf("failed to create bogus JWK: %s\n", err)
		return
	}
	_ = bogusKey.Set(jwk.KeyIDKey, "otherkey")
	_ = bogusKey.Set(jwk.AlgorithmKey, jwa.RS256)

	privset := jwk.NewSet()
	_ = privset.AddKey(realKey)
	publicset, err := jwk.PublicSetOf(privset)

	if err != nil {
		fmt.Printf("failed to create public JWKS: %s\n", err)
		return
	}
	mail := "schnöselberg@data4life.care"
	token, err1 := jwt.NewBuilder().Claim("email", mail).Build()

	audience := "phdp-userinfo"
	issuer := "azure"
	tokenAudIss, err2 := jwt.NewBuilder().Claim("email", mail).Audience([]string{audience}).Issuer(issuer).Build()

	if err1 != nil || err2 != nil {
		t.Fatal("Could not generate token")
	}

	goodSignedTokenMissingAudIss, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, realKey))

	if err != nil {
		t.Fatal("Could not sign token")
	}

	goodSignedToken, err := jwt.Sign(tokenAudIss, jwt.WithKey(jwa.RS256, realKey))
	if err != nil {
		t.Fatal("Could not sign token")
	}

	badSignedToken, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, bogusKey))
	if err != nil {
		t.Fatal("Could not sign token")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(publicset)
	}))

	defer srv.Close()

	for _, tc := range []struct {
		name     string
		issuer   string
		audience string
		request  *http.Request
		check    tut.ResponseCheckFunc
	}{
		{
			name:     "Middleware should not accept without valid issuer and audience",
			issuer:   "",
			audience: "",
			request: tut.Request(
				tut.ReqWithHeader("Authorization", fmt.Sprintf("Bearer %s", string(goodSignedTokenMissingAudIss))),
			),
			check: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusUnauthorized),
			),
		},
		{
			name:     "Middleware should only accept with valid issuer and audience",
			issuer:   issuer,
			audience: audience,
			request: tut.Request(
				tut.ReqWithHeader("Authorization", fmt.Sprintf("Bearer %s", string(goodSignedToken))),
			),
			check: tut.CheckResponse(
				tut.RespBodyHasKey("email", tut.ValueEquals(mail)),
				tut.RespHasStatusCode(http.StatusOK),
			),
		},
		{
			name:     "Middleware should not accept JWT token with wrong public key",
			issuer:   issuer,
			audience: audience,
			request: tut.Request(
				tut.ReqWithHeader("Authorization", fmt.Sprintf("Bearer %s", string(badSignedToken))),
			),
			check: tut.CheckResponse(tut.RespHasStatusCode(http.StatusUnauthorized)),
		}, {
			name:     "Middleware should not accept request with no token",
			issuer:   issuer,
			audience: audience,
			request:  tut.Request(),
			check:    tut.CheckResponse(tut.RespHasStatusCode(http.StatusUnauthorized)),
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			auth, err := NewJwtAuthenticator(
				context.Background(),
				tut.NewNopLogger().Logger,
				srv.URL,
				time.Hour,
			)

			if err != nil {
				t.Errorf("Could not start auth middleware: %v", err)
			}

			if keys := auth.keyStore.Len(); keys == 0 {
				t.Errorf("No keys parsed")
			}
			rec := httptest.NewRecorder()

			auth.JWTAuthorize(tc.audience, tc.issuer)(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					token, ok := r.Context().Value(JwtContext).(jwt.Token)
					if !ok {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					_ = json.NewEncoder(w).Encode(token)
				},
			)).ServeHTTP(rec, tc.request)

			if err := tc.check(rec.Result()); err != nil {
				t.Error(err)
			} else {
				t.Log(rec.Body)
			}
		})
	}
}
