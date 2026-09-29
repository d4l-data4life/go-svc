package donortoken_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/d4l-data4life/go-svc/pkg/donortoken"
)

const (
	testKid     = "dispatcher-key-1"
	testAlpID   = "7b6c5d5e-2c1f-4f7e-9d1a-0d8e2b7f4a11"
	testStudyID = "ecov~preview"
)

// fakeClock is a manually advanced clock.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Now().Truncate(time.Second)}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// signingKey is a dispatcher ES256 key pair.
type signingKey struct {
	raw     *ecdsa.PrivateKey
	private jwk.Key
	public  jwk.Key
}

func newSigningKey(t *testing.T, kid string) signingKey {
	t.Helper()
	raw, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	private, err := jwk.FromRaw(raw)
	require.NoError(t, err)
	require.NoError(t, private.Set(jwk.KeyIDKey, kid))
	require.NoError(t, private.Set(jwk.AlgorithmKey, jwa.ES256))
	require.NoError(t, private.Set(jwk.KeyUsageKey, jwk.ForSignature))
	public, err := private.PublicKey()
	require.NoError(t, err)
	return signingKey{raw: raw, private: private, public: public}
}

// dispatcherStub is a data-dispatcher publishing its current signing key(s) as JWKS.
type dispatcherStub struct {
	t       *testing.T
	server  *httptest.Server
	mu      sync.Mutex
	keys    []jwk.Key
	status  int
	fetches atomic.Int32
}

func newDispatcherStub(t *testing.T, keys ...signingKey) *dispatcherStub {
	t.Helper()
	s := &dispatcherStub{t: t, status: http.StatusOK}
	s.publish(keys...)
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, donortoken.JWKSPath, r.URL.Path)
		s.fetches.Add(1)
		s.mu.Lock()
		status := s.status
		set := jwk.NewSet()
		for _, k := range s.keys {
			assert.NoError(t, set.AddKey(k))
		}
		s.mu.Unlock()
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "max-age=300")
		assert.NoError(t, json.NewEncoder(w).Encode(set))
	}))
	t.Cleanup(s.server.Close)
	return s
}

// publish replaces the published key set with the public halves of keys.
func (s *dispatcherStub) publish(keys ...signingKey) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys = s.keys[:0]
	for _, k := range keys {
		s.keys = append(s.keys, k.public)
	}
}

// publishRaw publishes arbitrary JWKs (e.g. keys the verifier must ignore).
func (s *dispatcherStub) publishRaw(keys ...jwk.Key) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys = keys
}

func (s *dispatcherStub) failWith(status int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = status
}

func (s *dispatcherStub) verifier(t *testing.T, clock *fakeClock, opts ...donortoken.Option) *donortoken.Verifier {
	t.Helper()
	opts = append([]donortoken.Option{donortoken.WithClock(clock.Now)}, opts...)
	v, err := donortoken.NewVerifier(context.Background(), s.server.URL+"/", opts...)
	require.NoError(t, err)
	t.Cleanup(v.Close)
	return v
}

// claims are the token claims; a nil value omits the claim.
type claims map[string]any

func validClaims(now time.Time) claims {
	return claims{
		jwt.IssuerKey:           donortoken.Issuer,
		jwt.SubjectKey:          testAlpID,
		donortoken.StudyIDClaim: testStudyID,
		jwt.IssuedAtKey:         now.Add(-time.Minute),
		jwt.ExpirationKey:       now.Add(4 * time.Minute),
		jwt.JwtIDKey:            "jti-1",
	}
}

func (c claims) with(key string, value any) claims {
	out := claims{}
	for k, v := range c {
		out[k] = v
	}
	if value == nil {
		delete(out, key)
	} else {
		out[key] = value
	}
	return out
}

func (c claims) token(t *testing.T) jwt.Token {
	t.Helper()
	tok := jwt.New()
	for k, v := range c {
		require.NoError(t, tok.Set(k, v))
	}
	return tok
}

// sign signs the claims with key (a jwk.Key carrying the kid, or a raw key without kid).
func (c claims) sign(t *testing.T, alg jwa.SignatureAlgorithm, key any) string {
	t.Helper()
	signed, err := jwt.Sign(c.token(t), jwt.WithKey(alg, key))
	require.NoError(t, err)
	return string(signed)
}

// unsigned builds an `alg: none` token.
func (c claims) unsigned(t *testing.T) string {
	t.Helper()
	payload, err := json.Marshal(c.token(t))
	require.NoError(t, err)
	header := `{"alg":"none","kid":"` + testKid + `","typ":"JWT"}`
	return base64.RawURLEncoding.EncodeToString([]byte(header)) + "." + base64.RawURLEncoding.EncodeToString(payload) + "."
}
