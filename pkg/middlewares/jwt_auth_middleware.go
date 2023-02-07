package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/log"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type ContextKey string

const (
	JwtContext ContextKey = "jwtContext"
)

type JwtAuthenticator struct {
	keyStore jwk.Set
	uri      string
	l        *log.Logger
}

// Creates a new authentication middleware and fetches public key from a discovery url (presumably provided by ms azure).
// Afterwards, fetching happens automatically when the middleware is used based on an interval. All key's algorithm's are automatically set to RS256.
func NewJwtAuthenticator(ctx context.Context, l *log.Logger, azureADPublicKeyDiscoveryUrl string, refreshInterval time.Duration) (*JwtAuthenticator, error) {
	if _, err := url.Parse(azureADPublicKeyDiscoveryUrl); err != nil {
		errormsg := fmt.Errorf("could not parse url")
		_ = l.ErrGeneric(ctx, errormsg)
		return nil, errormsg
	}

	c := jwk.NewCache(ctx)

	// set the algorithm to RS256 because azure doesn't due this to help jwk to reference the key
	err := c.Register(azureADPublicKeyDiscoveryUrl,
		jwk.WithMinRefreshInterval(refreshInterval),
		jwk.WithPostFetcher(jwk.PostFetchFunc(func(uri string, keyset jwk.Set) (jwk.Set, error) {
			for it := keyset.Keys(ctx); it.Next(ctx); {
				key, ok := it.Pair().Value.(jwk.Key)
				if ok && key.Algorithm() != jwa.RS256 {
					_ = key.Set("alg", "RS256")
				}
			}
			return keyset, nil
		})),
	)

	if err != nil {
		_ = l.ErrGeneric(ctx, fmt.Errorf("could not register key store"))
		return nil, err
	}

	_, err = c.Refresh(ctx, azureADPublicKeyDiscoveryUrl)

	if err != nil {
		_ = l.ErrGeneric(ctx, fmt.Errorf("could not refresh key store"))
		return nil, err
	}

	cached := jwk.NewCachedSet(c, azureADPublicKeyDiscoveryUrl)

	_ = l.InfoGeneric(ctx, fmt.Sprintf("loaded keystore with %d keys", cached.Len()))

	ja := JwtAuthenticator{
		uri:      azureADPublicKeyDiscoveryUrl,
		l:        l,
		keyStore: cached,
	}

	return &ja, nil
}

// Middleware that checks for signed jwt tokens in the `Authorization` header. Token field `issuer` and `audience`
// must match with the supplied parameters. Issuer is the clientID and audience is the user group which can be taken
// from the azure configuration.
func (ja *JwtAuthenticator) JWTAuthorize(audience, issuer string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			token, err := jwt.ParseRequest(
				r,
				jwt.WithContext(r.Context()),
				jwt.WithKeySet(ja.keyStore),
				jwt.WithValidate(true),
				jwt.WithAudience(audience),
				jwt.WithIssuer(issuer),
			)

			if err != nil {
				_ = ja.l.ErrGeneric(r.Context(), err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), JwtContext, token))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
