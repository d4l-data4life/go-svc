// Package donortoken verifies the short-lived donor access tokens issued by data-dispatcher.
//
// A donor's app obtains an access token from data-dispatcher (which knows the donor's
// registration) and presents it to other services as `Authorization: Bearer <jwt>`.
// Those services never ask the dispatcher about a donor: they verify the token against
// the dispatcher's public signing keys (JWKS) and trust its claims.
//
// Token contract: ES256 JWT with header `kid`; claims `iss` = Issuer, `sub` = the donor's
// AlpID, `studyID` (StudyIDClaim) = the program name the donor registered for (it may end
// in `~preview`), `iat`, `exp` and `jti`. The keys are published as a JWKS at
// `{dispatcherBaseURL}` + JWKSPath.
//
// Use NewVerifier to create a Verifier and either call Verify directly or wrap handlers with
// HTTPMiddleware / UnaryServerInterceptor, which put the verified Claims into the context
// (read them with FromContext).
package donortoken

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// Contract of the data-dispatcher donor access tokens.
const (
	// Issuer is the `iss` claim every dispatcher access token carries.
	Issuer = "data-dispatcher"
	// JWKSPath is the dispatcher route serving the token signing keys, relative to its base URL.
	JWKSPath = "/api/v2/jwks"
	// StudyIDClaim is the private claim holding the program name the token was issued for.
	StudyIDClaim = "studyID"
)

const (
	// DefaultMinRefreshInterval bounds how often the JWKS is fetched: both the background
	// refresh (which otherwise follows the endpoint's Cache-Control max-age) and the forced
	// refresh on an unknown `kid`, so a flood of bogus tokens cannot turn into a flood of
	// requests to the dispatcher.
	DefaultMinRefreshInterval = time.Minute
	// DefaultHTTPTimeout is the timeout of the default JWKS HTTP client.
	DefaultHTTPTimeout = 10 * time.Second

	// clockSkew tolerates small clock differences between dispatcher and verifier.
	clockSkew = 30 * time.Second
	// maxJWKSBody caps the JWKS response size.
	maxJWKSBody = 1 << 20
)

var (
	// ErrInvalidToken is matched (errors.Is) by every error Verify returns.
	ErrInvalidToken = errors.New("invalid donor access token")
	// ErrExpired is additionally matched by errors for expired tokens, the one failure
	// expected in normal operation (see IsExpired).
	ErrExpired = errors.New("donor access token expired")
)

// Claims are the verified claims of a donor access token.
type Claims struct {
	// AlpID is the donor's pseudonymous per-study subject id (`sub`).
	AlpID string
	// StudyID is the program the token was issued for (`studyID`), as registered: it may
	// carry the "~preview" suffix.
	StudyID string
	// ExpiresAt is the token's expiry (`exp`).
	ExpiresAt time.Time
}

// verifyError describes why a token was rejected. Its text is meant for server-side logs
// only; callers must answer with a generic message.
type verifyError struct {
	reason  string
	cause   error
	expired bool
}

func (e *verifyError) Error() string {
	if e.cause == nil {
		return ErrInvalidToken.Error() + ": " + e.reason
	}
	return ErrInvalidToken.Error() + ": " + e.reason + ": " + e.cause.Error()
}

func (e *verifyError) Unwrap() []error {
	errs := []error{ErrInvalidToken}
	if e.expired {
		errs = append(errs, ErrExpired)
	}
	if e.cause != nil {
		errs = append(errs, e.cause)
	}
	return errs
}

func reject(reason string, cause error) error {
	return &verifyError{reason: reason, cause: cause}
}

// IsExpired reports whether err is the rejection of an otherwise valid but expired token.
func IsExpired(err error) bool {
	return errors.Is(err, ErrExpired)
}

// Option configures a Verifier.
type Option func(*config)

type config struct {
	httpClient         *http.Client
	now                func() time.Time
	minRefreshInterval time.Duration
}

// WithHTTPClient sets the HTTP client used to fetch the JWKS (default: a client with
// DefaultHTTPTimeout).
func WithHTTPClient(c *http.Client) Option {
	return func(cfg *config) {
		cfg.httpClient = c
	}
}

// WithClock sets the clock used to validate token lifetimes and to rate-limit refreshes.
func WithClock(now func() time.Time) Option {
	return func(cfg *config) {
		cfg.now = now
	}
}

// WithMinRefreshInterval overrides DefaultMinRefreshInterval.
func WithMinRefreshInterval(d time.Duration) Option {
	return func(cfg *config) {
		cfg.minRefreshInterval = d
	}
}

// Verifier validates donor access tokens against the dispatcher's JWKS. It caches the key
// set (refreshed in the background following the endpoint's cache headers), re-fetches it
// (rate-limited) when a token names an unknown `kid`, and fails closed when no cached key
// matches. It is safe for concurrent use.
type Verifier struct {
	jwksURL            string
	cache              *jwk.Cache
	cancel             context.CancelFunc
	now                func() time.Time
	minRefreshInterval time.Duration

	refreshMu     sync.Mutex
	lastRefreshAt time.Time
}

// NewVerifier creates a Verifier for the tokens of the data-dispatcher at dispatcherURL
// (its base URL; the JWKS is fetched from dispatcherURL + JWKSPath). The key set is fetched
// once synchronously, so an unreachable dispatcher is reported here. Background refreshes
// run until ctx is done or Close is called.
func NewVerifier(ctx context.Context, dispatcherURL string, opts ...Option) (*Verifier, error) {
	cfg := config{
		httpClient:         &http.Client{Timeout: DefaultHTTPTimeout},
		now:                time.Now,
		minRefreshInterval: DefaultMinRefreshInterval,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	if strings.TrimSpace(dispatcherURL) == "" {
		return nil, errors.New("donortoken: dispatcher URL is empty")
	}

	cacheCtx, cancel := context.WithCancel(ctx)
	v := &Verifier{
		jwksURL:            strings.TrimRight(dispatcherURL, "/") + JWKSPath,
		cache:              jwk.NewCache(cacheCtx),
		cancel:             cancel,
		now:                cfg.now,
		minRefreshInterval: cfg.minRefreshInterval,
	}
	err := v.cache.Register(v.jwksURL,
		jwk.WithHTTPClient(limitedClient{cfg.httpClient}),
		jwk.WithMinRefreshInterval(cfg.minRefreshInterval),
	)
	if err == nil {
		_, err = v.cache.Refresh(ctx, v.jwksURL)
	}
	if err != nil {
		cancel()
		return nil, fmt.Errorf("donortoken: fetching the dispatcher JWKS from %s: %w", v.jwksURL, err)
	}
	v.lastRefreshAt = v.now()
	return v, nil
}

// Close stops the background refresh of the key set.
func (v *Verifier) Close() {
	v.cancel()
}

// Verify checks the token's ES256 signature against the dispatcher's keys and validates its
// claims (issuer, lifetime, subject and study). Every failure matches ErrInvalidToken, expiry
// additionally ErrExpired. The error text is for server-side logs; never return it to callers.
func (v *Verifier) Verify(ctx context.Context, raw string) (Claims, error) {
	msg, err := jws.Parse([]byte(raw))
	if err != nil {
		return Claims{}, reject("malformed token", nil)
	}
	sigs := msg.Signatures()
	if len(sigs) != 1 {
		return Claims{}, reject("unexpected number of signatures", nil)
	}
	header := sigs[0].ProtectedHeaders()
	// ES256 only: no HMAC (the public key would double as the secret) and no "none".
	if header.Algorithm() != jwa.ES256 {
		return Claims{}, reject(fmt.Sprintf("unexpected signing algorithm %q", header.Algorithm()), nil)
	}
	kid := header.KeyID()
	if kid == "" {
		return Claims{}, reject("missing kid", nil)
	}
	key, err := v.lookupKey(ctx, kid)
	if err != nil {
		return Claims{}, err
	}

	token, err := jwt.Parse([]byte(raw),
		jwt.WithKey(jwa.ES256, key),
		jwt.WithValidate(true),
		jwt.WithClock(jwt.ClockFunc(v.now)),
		jwt.WithAcceptableSkew(clockSkew),
		jwt.WithIssuer(Issuer),
		jwt.WithRequiredClaim(jwt.ExpirationKey),
		jwt.WithRequiredClaim(jwt.SubjectKey),
		jwt.WithRequiredClaim(StudyIDClaim),
	)
	if err != nil {
		return Claims{}, classify(err)
	}
	studyID, _ := token.PrivateClaims()[StudyIDClaim].(string)
	claims := Claims{AlpID: token.Subject(), StudyID: studyID, ExpiresAt: token.Expiration()}
	if claims.AlpID == "" || claims.StudyID == "" {
		return Claims{}, reject("empty sub or studyID", nil)
	}
	return claims, nil
}

// classify maps jwx parse/validation errors to rejections.
func classify(err error) error {
	switch {
	case errors.Is(err, jwt.ErrTokenExpired()):
		return &verifyError{reason: "token expired", expired: true}
	case errors.Is(err, jwt.ErrInvalidIssuer()):
		return reject("unexpected issuer", nil)
	case jwt.IsValidationError(err):
		return reject("claim validation failed", err)
	default:
		return reject("signature verification failed", nil)
	}
}

// lookupKey resolves kid to a usable ES256 key of the cached set, refreshing the set once
// (rate-limited) when the id is unknown.
func (v *Verifier) lookupKey(ctx context.Context, kid string) (jwk.Key, error) {
	set, err := v.cache.Get(ctx, v.jwksURL)
	if err != nil {
		return nil, reject("no key set available", err)
	}
	if key, ok := usableKey(set, kid); ok {
		return key, nil
	}
	set, err = v.refresh(ctx)
	if err != nil {
		return nil, reject("unknown kid, refreshing the key set failed", err)
	}
	if key, ok := usableKey(set, kid); ok {
		return key, nil
	}
	return nil, reject(fmt.Sprintf("unknown kid %q", kid), nil)
}

// refresh re-fetches the key set unless a fetch was done within the minimum refresh interval,
// in which case the cached set is returned. Concurrent callers wait for the in-flight fetch.
func (v *Verifier) refresh(ctx context.Context) (jwk.Set, error) {
	v.refreshMu.Lock()
	defer v.refreshMu.Unlock()
	now := v.now()
	if now.Sub(v.lastRefreshAt) < v.minRefreshInterval {
		return v.cache.Get(ctx, v.jwksURL)
	}
	v.lastRefreshAt = now
	return v.cache.Refresh(ctx, v.jwksURL)
}

// usableKey returns the key with the given id if it is a P-256 EC key usable for ES256
// signatures.
func usableKey(set jwk.Set, kid string) (jwk.Key, bool) {
	key, ok := set.LookupKeyID(kid)
	if !ok || key.KeyType() != jwa.EC {
		return nil, false
	}
	ecKey, ok := key.(jwk.ECDSAPublicKey)
	if !ok || ecKey.Crv() != jwa.P256 {
		return nil, false
	}
	if alg := key.Algorithm(); alg != nil && alg.String() != "" && alg.String() != jwa.ES256.String() {
		return nil, false
	}
	if use := key.KeyUsage(); use != "" && use != string(jwk.ForSignature) {
		return nil, false
	}
	return key, true
}

// limitedClient caps the size of JWKS responses.
type limitedClient struct {
	client *http.Client
}

func (c limitedClient) Get(url string) (*http.Response, error) {
	// The jwk cache fetches without a context; the client timeout bounds the request.
	res, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	res.Body = struct {
		io.Reader
		io.Closer
	}{io.LimitReader(res.Body, maxJWKSBody), res.Body}
	return res, nil
}
