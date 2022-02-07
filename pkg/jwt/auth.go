package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"

	"github.com/gesundheitscloud/go-svc/pkg/d4lcontext"
	"github.com/gesundheitscloud/go-svc/pkg/dynamic"

	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrNoClaims      = errors.New("missing claims")
	ErrMissingScope  = errors.New("necessary scope not in jwt")
	ErrInvalidToken  = errors.New("token is invalid")
	ErrTokenNotFound = errors.New("access token not found in the request")
)

// tokenExtractor is an interface for extracting a token from an HTTP request.
type tokenExtractor interface {
	ExtractToken(*http.Request) (string, error)
}

type JWTPublicKeysProvider interface {
	JWTPublicKeys() ([]dynamic.JWTPublicKey, error)
}

// Authenticator contains the public key necessary to verify the signature.
type Authenticator struct {
	keyProvider     JWTPublicKeysProvider
	logger          logger
	tokenExtractors []tokenExtractor
}

// NewAuthenticator creates an Authenticator that can be used for auth Middleware for
// JWT verification against multiple publick keys provided by a KeyProvider
// It checks for access tokens in the Authorization header and in the "access_token" parameter
// in case of a form request body.
func NewAuthenticator(pkp JWTPublicKeysProvider, l logger) *Authenticator {
	return &Authenticator{
		keyProvider: pkp,
		logger:      l,
		tokenExtractors: []tokenExtractor{
			newHeaderExtractor(),
			newArgumentExtractor(),
		},
	}
}

type authenticatorOption func(*Authenticator)

// NewAuthenticatorWithOptions creates an Authenticator that creates an auth Middleware for
// JWT verification against multiple public keys provided by a KeyProvider.
// It accepts extra options that can be used to customize the behavior of the authenticator.
func NewAuthenticatorWithOptions(pkp JWTPublicKeysProvider, l logger, options ...authenticatorOption) *Authenticator {
	a := NewAuthenticator(pkp, l)

	for _, option := range options {
		option(a)
	}

	return a
}

// AcceptAccessCookie enables the authenticator to check for access
// tokens in cookies sent with the request using the jwt.AccessCookieName
// cookie name.
// Only use this if you have a proper CSRF protection in place.
func AcceptAccessCookie(a *Authenticator) {
	a.tokenExtractors = append(
		a.tokenExtractors,

		// the cookie extractor checks if the token is included in the access cookie.
		// The access token from the cookie is extracted only if the CSRF protection
		// is also valid on the request.
		newCookieExtractor(),
	)
}

// Extract extracts the claims from the JWT and puts it into the context.
// It checks if any of the many JWT keys work for verifying the claims.
// It never fails, so it is not intended to be used for access control, just for
// making the information in the JWT available to other middlewares.
// It sets in the context of the request:
// 1. the d4lcontext keys (currently client ID, user ID, tenant ID)
// 2. its own internal context keys such that a downstream middleware has access to any of these.
func (auth *Authenticator) Extract(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// extract the raw token
		rawToken, err := auth.extractToken(r)
		if err != nil {
			_ = auth.logger.InfoGeneric(r.Context(), fmt.Errorf("jwt.Extract: %w", err).Error())
			next.ServeHTTP(w, r)
			return
		}

		candidateKeys, err := auth.keyProvider.JWTPublicKeys()
		if err != nil {
			_ = auth.logger.InfoGeneric(r.Context(), fmt.Errorf("jwt.Extract: keyProvider.PublicKeys() failed: %w", err).Error())
			next.ServeHTTP(w, r)
			return
		}

		for _, key := range candidateKeys {
			claims, err := verifyPubKey(key.Key, rawToken)
			if err == nil {
				r = addClaimsToContext(r, claims)
				break
			}
		}

		next.ServeHTTP(w, r)
	})
}

// Verify returns a middleware that allows a call only if it contains a valid token
// that passes all the configured rules.
// For simplicity, only the first token found is checked currently. Therefore, calls should only
// include valid tokens, as invalid ones might shadow valid ones (e.g. an invalid token in header
// would result in a failure, even if the cookie contains a valid one.
// The order in which the tokens are checked is: header, form body, cookie (if enabled).
// For validating the token signature all configured keys are tried.
func (auth *Authenticator) Verify(rules ...rule) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// extract the raw token
			rawToken, err := auth.extractToken(r)
			if err != nil {
				_ = auth.logger.ErrUserAuth(r.Context(), fmt.Errorf("cannot extract token from request: %w", err))
				http.Error(w, "", http.StatusUnauthorized)
				return
			}

			// get all the current keys
			candidateKeys, err := auth.keyProvider.JWTPublicKeys()
			if err != nil {
				_ = auth.logger.ErrUserAuth(r.Context(), fmt.Errorf("getting the public keys failed: %w", err))
				http.Error(w, "", http.StatusInternalServerError)
				return
			}

			for i, key := range candidateKeys {
				pkName := fmt.Sprintf("public key '%s' (%d of %d) ", key.Name, i+1, len(candidateKeys))

				// check if the current key can be used to validate the token signature
				claims, err := verifyPubKey(key.Key, rawToken)
				if err != nil {
					// the public key is not a match: continue to the next key
					_ = auth.logger.InfoGeneric(r.Context(), fmt.Errorf("%s does not match: %w", pkName, err).Error())
					continue
				}

				// we've found a valid token. Check if the claims pass the rules
				err = verifyAllRules(r, claims, rules...)
				if err != nil {
					_ = auth.logger.ErrUserAuth(r.Context(), fmt.Errorf("rules verification failed: %w", err))
					// we stop trying when a matching pubkey is found but rules verification failed
					// it is impossible that any other pubkey will match and pass the rules validation
					http.Error(w, "", http.StatusUnauthorized)
					return
				}

				// found valid pub-key that also passed all the rules checks
				// must write claims into the context - other middleware and handlers depend on the claims being in the context
				r = addClaimsToContext(r, claims)
				next.ServeHTTP(w, r)
				return
			}

			// haven't found any valid key
			_ = auth.logger.ErrUserAuth(r.Context(),
				fmt.Errorf("verification failed for all %d public keys", len(candidateKeys)))
			http.Error(w, "", http.StatusUnauthorized)
		})
	}
}

// extractTokens extracts one candidate token from a request.
// As multiple ways to include a token are possible, a request may contain multiple tokens.
// This method iterates over all the extractors configured for the Authenticator
// and returns the first token found. For simplicity, only the first found token is considered.
// Returns ErrTokenNotFound
func (auth *Authenticator) extractToken(r *http.Request) (string, error) {
	for _, e := range auth.tokenExtractors {
		rawToken, err := e.ExtractToken(r)
		if err != nil {
			// log the error as info level. It might help for debug but at this stage we can't
			// assume that this is an error.
			_ = auth.logger.InfoGeneric(r.Context(), fmt.Errorf("extract: %w", err).Error())
		} else {
			return rawToken, nil
		}
	}

	return "", ErrTokenNotFound
}

// verifyPubKey verifies a raw token against a single JWT public key
// If the token is valid, it returns the JWT-claims object
func verifyPubKey(pubKey *rsa.PublicKey, rawToken string) (*Claims, error) {
	if pubKey == nil {
		return nil, fmt.Errorf("public key missing")
	}

	parsedToken, err := jwt.ParseWithClaims(rawToken, &Claims{},
		func(_ *jwt.Token) (interface{}, error) {
			return pubKey, nil
		})
	if err != nil {
		return nil, fmt.Errorf("cannot parse token: %w", err)
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("token invalid")
	}

	claims, ok := parsedToken.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("%w: cannot understand claims", ErrNoClaims)
	}
	return claims, nil
}

// verifyAllRules verifies that the claims and the request pass all the given rules
func verifyAllRules(r *http.Request, claims *Claims, rules ...rule) error {
	for _, rule := range rules {
		if err := rule(r, claims); err != nil {
			return fmt.Errorf("rule verification failed: %w", err)
		}
	}
	return nil
}

func addClaimsToContext(r *http.Request, claims *Claims) *http.Request {
	newR := d4lcontext.WithClientID(r, claims.ClientID)
	newR = d4lcontext.WithUserID(newR, claims.Subject.ID.String())
	newR = d4lcontext.WithTenantID(newR, claims.TenantID)

	// also write the claims into the context for services using this package's context keys
	newR = newR.WithContext(NewContext(newR.Context(), claims))
	return newR
}
