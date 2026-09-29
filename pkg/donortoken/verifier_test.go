package donortoken_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/d4l-data4life/go-svc/pkg/donortoken"
)

func TestVerify(t *testing.T) {
	clock := newFakeClock()
	now := clock.Now()
	key := newSigningKey(t, testKid)
	impostor := newSigningKey(t, testKid) // same kid, key not published
	unpublished := newSigningKey(t, "unpublished")
	stub := newDispatcherStub(t, key)
	v := stub.verifier(t, clock)

	p384, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)
	p384Key, err := jwk.FromRaw(p384)
	require.NoError(t, err)
	require.NoError(t, p384Key.Set(jwk.KeyIDKey, testKid))

	x := key.public.(jwk.ECDSAPublicKey).X()
	valid := validClaims(now)
	validToken := valid.sign(t, jwa.ES256, key.private)
	parts := strings.Split(validToken, ".")
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	require.NoError(t, err)
	sig[10] ^= 1
	tampered := parts[0] + "." + parts[1] + "." + base64.RawURLEncoding.EncodeToString(sig)
	otherPayload := strings.Split(valid.with(jwt.SubjectKey, "someone-else").sign(t, jwa.ES256, key.private), ".")[1]

	tests := []struct {
		name    string
		token   string
		want    donortoken.Claims
		expired bool
	}{
		{name: "valid", token: validToken,
			want: donortoken.Claims{AlpID: testAlpID, StudyID: testStudyID, ExpiresAt: now.Add(4 * time.Minute)}},
		{name: "expired within skew", token: valid.with(jwt.ExpirationKey, now.Add(-10*time.Second)).sign(t, jwa.ES256, key.private),
			want: donortoken.Claims{AlpID: testAlpID, StudyID: testStudyID, ExpiresAt: now.Add(-10 * time.Second)}},
		{name: "expired", token: valid.with(jwt.ExpirationKey, now.Add(-time.Minute)).sign(t, jwa.ES256, key.private), expired: true},
		{name: "wrong issuer", token: valid.with(jwt.IssuerKey, "data-receiver").sign(t, jwa.ES256, key.private)},
		{name: "missing issuer", token: valid.with(jwt.IssuerKey, nil).sign(t, jwa.ES256, key.private)},
		{name: "missing sub", token: valid.with(jwt.SubjectKey, nil).sign(t, jwa.ES256, key.private)},
		{name: "empty sub", token: valid.with(jwt.SubjectKey, "").sign(t, jwa.ES256, key.private)},
		{name: "missing studyID", token: valid.with(donortoken.StudyIDClaim, nil).sign(t, jwa.ES256, key.private)},
		{name: "empty studyID", token: valid.with(donortoken.StudyIDClaim, "").sign(t, jwa.ES256, key.private)},
		{name: "non-string studyID", token: valid.with(donortoken.StudyIDClaim, 42).sign(t, jwa.ES256, key.private)},
		{name: "missing exp", token: valid.with(jwt.ExpirationKey, nil).sign(t, jwa.ES256, key.private)},
		{name: "issued in the future", token: valid.with(jwt.IssuedAtKey, now.Add(time.Minute)).sign(t, jwa.ES256, key.private)},
		{name: "not yet valid", token: valid.with(jwt.NotBeforeKey, now.Add(time.Minute)).sign(t, jwa.ES256, key.private)},
		{name: "HS256 with the public key as secret", token: valid.sign(t, jwa.HS256, x)},
		{name: "alg none", token: valid.unsigned(t)},
		{name: "ES384", token: valid.sign(t, jwa.ES384, p384Key)},
		{name: "missing kid", token: valid.sign(t, jwa.ES256, key.raw)},
		{name: "unknown kid", token: valid.sign(t, jwa.ES256, unpublished.private)},
		{name: "signed by another key with the same kid", token: valid.sign(t, jwa.ES256, impostor.private)},
		{name: "tampered signature", token: tampered},
		{name: "swapped payload", token: parts[0] + "." + otherPayload + "." + parts[2]},
		{name: "garbage", token: "not-a-jwt"},
		{name: "empty", token: ""},
		{name: "three garbage parts", token: "a.b.c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := v.Verify(context.Background(), tt.token)
			if tt.want.AlpID != "" {
				require.NoError(t, err)
				assert.Equal(t, tt.want.AlpID, got.AlpID)
				assert.Equal(t, tt.want.StudyID, got.StudyID)
				assert.True(t, tt.want.ExpiresAt.Equal(got.ExpiresAt), "exp %v != %v", tt.want.ExpiresAt, got.ExpiresAt)
				return
			}
			require.Error(t, err)
			require.ErrorIs(t, err, donortoken.ErrInvalidToken)
			assert.Equal(t, tt.expired, donortoken.IsExpired(err))
			assert.Equal(t, donortoken.Claims{}, got)
		})
	}
}

func TestVerify_UnknownKidRefreshesRotatedKey(t *testing.T) {
	clock := newFakeClock()
	oldKey, newKey := newSigningKey(t, "key-1"), newSigningKey(t, "key-2")
	stub := newDispatcherStub(t, oldKey)
	v := stub.verifier(t, clock)
	require.EqualValues(t, 1, stub.fetches.Load())

	stub.publish(newKey)
	clock.Advance(2 * time.Minute)

	claims, err := v.Verify(context.Background(), validClaims(clock.Now()).sign(t, jwa.ES256, newKey.private))
	require.NoError(t, err)
	assert.Equal(t, testAlpID, claims.AlpID)
	assert.EqualValues(t, 2, stub.fetches.Load())

	// The rotated-away key is gone from the refreshed set: fail closed.
	_, err = v.Verify(context.Background(), validClaims(clock.Now()).sign(t, jwa.ES256, oldKey.private))
	require.ErrorIs(t, err, donortoken.ErrInvalidToken)
}

func TestVerify_RefreshIsRateLimited(t *testing.T) {
	clock := newFakeClock()
	oldKey, newKey := newSigningKey(t, "key-1"), newSigningKey(t, "key-2")
	stub := newDispatcherStub(t, oldKey)
	v := stub.verifier(t, clock, donortoken.WithMinRefreshInterval(time.Minute))
	stub.publish(newKey)

	token := validClaims(clock.Now()).sign(t, jwa.ES256, newKey.private)
	// Within the interval of the initial fetch: no refetch, however many bogus kids arrive.
	for range 5 {
		_, err := v.Verify(context.Background(), token)
		require.ErrorIs(t, err, donortoken.ErrInvalidToken)
	}
	assert.EqualValues(t, 1, stub.fetches.Load())

	clock.Advance(time.Minute)
	_, err := v.Verify(context.Background(), token)
	require.NoError(t, err)
	assert.EqualValues(t, 2, stub.fetches.Load())

	// Once refreshed, a still unknown kid does not refetch again within the interval.
	_, err = v.Verify(context.Background(), validClaims(clock.Now()).sign(t, jwa.ES256, newSigningKey(t, "key-3").private))
	require.ErrorIs(t, err, donortoken.ErrInvalidToken)
	assert.EqualValues(t, 2, stub.fetches.Load())
}

func TestVerify_FailedRefreshKeepsCachedKeys(t *testing.T) {
	clock := newFakeClock()
	key := newSigningKey(t, testKid)
	stub := newDispatcherStub(t, key)
	v := stub.verifier(t, clock)
	stub.failWith(http.StatusInternalServerError)
	clock.Advance(2 * time.Minute)

	_, err := v.Verify(context.Background(), validClaims(clock.Now()).sign(t, jwa.ES256, newSigningKey(t, "key-2").private))
	require.ErrorIs(t, err, donortoken.ErrInvalidToken)
	assert.EqualValues(t, 2, stub.fetches.Load())

	_, err = v.Verify(context.Background(), validClaims(clock.Now()).sign(t, jwa.ES256, key.private))
	require.NoError(t, err)
}

func TestVerify_IgnoresUnusableKeys(t *testing.T) {
	clock := newFakeClock()
	stub := newDispatcherStub(t)

	rsaRaw, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	rsaKey, err := jwk.FromRaw(rsaRaw.Public())
	require.NoError(t, err)
	require.NoError(t, rsaKey.Set(jwk.KeyIDKey, "rsa"))

	p384Raw, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)
	p384Key, err := jwk.FromRaw(p384Raw.Public())
	require.NoError(t, err)
	require.NoError(t, p384Key.Set(jwk.KeyIDKey, "p384"))

	wrongAlg := newSigningKey(t, "wrong-alg")
	require.NoError(t, wrongAlg.public.Set(jwk.AlgorithmKey, jwa.ES384))
	encryption := newSigningKey(t, "enc")
	require.NoError(t, encryption.public.Set(jwk.KeyUsageKey, jwk.ForEncryption))
	noMetadata := newSigningKey(t, "bare")
	require.NoError(t, noMetadata.public.Remove(jwk.AlgorithmKey))
	require.NoError(t, noMetadata.public.Remove(jwk.KeyUsageKey))

	stub.publishRaw(rsaKey, p384Key, wrongAlg.public, encryption.public, noMetadata.public)
	v := stub.verifier(t, clock)

	for _, k := range []signingKey{wrongAlg, encryption} {
		_, err = v.Verify(context.Background(), validClaims(clock.Now()).sign(t, jwa.ES256, k.private))
		require.ErrorIs(t, err, donortoken.ErrInvalidToken, k.private.KeyID())
	}
	for _, kid := range []string{"rsa", "p384"} {
		forged := newSigningKey(t, kid)
		_, err = v.Verify(context.Background(), validClaims(clock.Now()).sign(t, jwa.ES256, forged.private))
		require.ErrorIs(t, err, donortoken.ErrInvalidToken, kid)
	}
	// A P-256 key without alg/use is fine.
	_, err = v.Verify(context.Background(), validClaims(clock.Now()).sign(t, jwa.ES256, noMetadata.private))
	require.NoError(t, err)
}

func TestNewVerifier(t *testing.T) {
	t.Run("JWKS URL derived from the base URL", func(t *testing.T) {
		stub := newDispatcherStub(t, newSigningKey(t, testKid))
		for _, base := range []string{stub.server.URL, stub.server.URL + "/", stub.server.URL + "//"} {
			v, err := donortoken.NewVerifier(context.Background(), base, donortoken.WithHTTPClient(stub.server.Client()))
			require.NoError(t, err, base)
			v.Close()
		}
	})
	t.Run("empty URL", func(t *testing.T) {
		_, err := donortoken.NewVerifier(context.Background(), " ")
		require.Error(t, err)
	})
	t.Run("unreachable", func(t *testing.T) {
		stub := newDispatcherStub(t, newSigningKey(t, testKid))
		stub.server.Close()
		_, err := donortoken.NewVerifier(context.Background(), stub.server.URL)
		require.Error(t, err)
	})
	t.Run("error status", func(t *testing.T) {
		stub := newDispatcherStub(t, newSigningKey(t, testKid))
		stub.failWith(http.StatusServiceUnavailable)
		_, err := donortoken.NewVerifier(context.Background(), stub.server.URL)
		require.Error(t, err)
	})
	t.Run("not a JWKS", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("<html>"))
		}))
		defer server.Close()
		_, err := donortoken.NewVerifier(context.Background(), server.URL)
		require.Error(t, err)
	})
	t.Run("oversized JWKS", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"keys":[{"kty":"EC","x":"` + strings.Repeat("A", 2<<20) + `"}]}`))
		}))
		defer server.Close()
		_, err := donortoken.NewVerifier(context.Background(), server.URL)
		require.Error(t, err)
	})
}

func TestContext(t *testing.T) {
	_, ok := donortoken.FromContext(context.Background())
	assert.False(t, ok)

	want := donortoken.Claims{AlpID: testAlpID, StudyID: testStudyID}
	got, ok := donortoken.FromContext(donortoken.WithClaims(context.Background(), want))
	assert.True(t, ok)
	assert.Equal(t, want, got)
}
