package jwt

import (
	"crypto/rsa"
	"fmt"
	"net/http"

	"github.com/gesundheitscloud/go-svc/pkg/d4lcontext"
	"github.com/gesundheitscloud/go-svc/pkg/dynamic"
	"github.com/go-chi/chi"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
	jwtReq "github.com/golang-jwt/jwt/v4/request"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

const ErrMsgVerifier = "verification failed"

var (
	ErrNoClaims           = errors.New("missing claims")
	ErrMissingScope       = errors.New("necessary scope not in jwt")
	ErrInvalidToken       = errors.New("token is invalid")
	ErrPubKeyVerification = errors.New("JWT public key verification failed")
	ErrRulesVerification  = errors.New("JWT rules verification failed")
)

// DummyKeyProvider will be used when the deprecated constructor `New` is used
// We want to prevent setting Authenticator.keyProvider to nil, so instead this struct will be used
type DummyKeyProvider struct {
	Key *rsa.PublicKey
}

func (dkp *DummyKeyProvider) JWTPublicKeys() ([]dynamic.JWTPublicKey, error) {
	jwtpk := dynamic.JWTPublicKey{Key: dkp.Key, Name: "arbitrary", Comment: "generated in code jwt.New()"}
	return []dynamic.JWTPublicKey{jwtpk}, nil
}

type JWTPublicKeysProvider interface {
	JWTPublicKeys() ([]dynamic.JWTPublicKey, error)
}

// Authenticator contains the public key necessary to verify the signature.
type Authenticator struct {
	keyProvider JWTPublicKeysProvider
	logger      logger
}

// New (DEPRECATED in favour of NewAuthenticator) creates an Authenticator that creates an auth Middleware.
// It supports single public key in rsa form
func New(pk *rsa.PublicKey, l logger) *Authenticator {
	dkp := &DummyKeyProvider{Key: pk}
	return &Authenticator{keyProvider: dkp, logger: l}
}

// NewAuthenticator creates an Authenticator that creates an auth Middleware for
// JWT verification against multiple publick keys provided by a KeyProvider
func NewAuthenticator(pkp JWTPublicKeysProvider, l logger) *Authenticator {
	return &Authenticator{keyProvider: pkp, logger: l}
}

type rule func(*http.Request) error

// Verify checks if JWT satisfies the given rules for at least one of public keys provided by the KeyProvider
func (auth *Authenticator) Verify(rules ...rule) func(handler http.Handler) http.Handler {
	return auth.VerifyAny(rules...)
}

// Extract extracts the claims from the JWT and puts it into the context.
// It checks if any of the many JWT keys work for verifying the claims.
// It never fails, so it is not intended to be used for access control, just for
// making the information in the JWT available to other middlewares.
// It sets 1. the d4lcontext keys (currently client ID, user ID, tenant ID) and 2. its own
// internal context keys such that a downstream middleware has access to any of these.
func (auth *Authenticator) Extract(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		candidateKeys, err := auth.keyProvider.JWTPublicKeys()
		if err != nil {
			_ = auth.logger.ErrGeneric(r.Context(), fmt.Errorf("jwt.Extract: keyProvider.PublicKeys() failed: %w", err))
			next.ServeHTTP(w, r)
			return
		}

		for _, key := range candidateKeys {
			_, claims, err := auth.verifyPubKey(r, key.Key, key.Name)
			if err == nil {
				// add values to d4lcontext
				r = d4lcontext.WithClientID(r, claims.ClientID)
				r = d4lcontext.WithUserID(r, claims.Subject.ID.String())
				r = d4lcontext.WithTenantID(r, claims.TenantID)

				// also write the claims into the context for services using this package's context keys
				r = r.WithContext(NewContext(r.Context(), claims))

				break
			}
		}

		next.ServeHTTP(w, r)
	})
}

// VerifyAny checks if any of the many JWT keys satisfies given rules
func (auth *Authenticator) VerifyAny(rules ...rule) func(handler http.Handler) http.Handler {
	candidateKeys, err := auth.keyProvider.JWTPublicKeys()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err != nil {
				err := fmt.Errorf("jwt.VerifyAny: keyProvider.PublicKeys() failed: %w", err)
				_ = auth.logger.ErrUserAuth(r.Context(), errors.Wrap(err, ErrMsgVerifier))
				httpClientError(w, http.StatusInternalServerError)
				return
			}

			lastStatus := 0
			for i, key := range candidateKeys {
				msg := fmt.Sprintf("public key '%s' (%d of %d) ", key.Name, i+1, len(candidateKeys))
				req, status, err := auth.verify(r, key, rules...) // this should write claims to r's context and return it as req
				lastStatus = status
				switch {
				case errors.Is(err, ErrPubKeyVerification):
					_ = auth.logger.ErrUserAuth(r.Context(),
						fmt.Errorf("%s does not match: %w", msg, err))
				case errors.Is(err, ErrRulesVerification):
					_ = auth.logger.ErrUserAuth(r.Context(),
						fmt.Errorf("%s failed rules verification: %w", msg, err))
					// we stop trying when a matching pubkey is found but rules verification failed
					// it is impossible that any other pubkey will match and pass the rules validation
					httpClientError(w, http.StatusUnauthorized)
					return
				case err == nil: // found valid pub-key
					_ = auth.logger.InfoGeneric(r.Context(), fmt.Sprintf("%s matches", msg))
					next.ServeHTTP(w, req) // important to use req, so that the r's context includes JWT claims
					return
				}
			}
			// haven't found any valid key
			err := fmt.Errorf("jwt.VerifyAny: verification failed for all %d public keys", len(candidateKeys))
			_ = auth.logger.ErrUserAuth(r.Context(), errors.Wrap(err, ErrMsgVerifier))
			httpClientError(w, lastStatus)
		})
	}
}

// verify combines verifyPubKey and verifyRules and writes JWT claims into request's context
func (auth *Authenticator) verify(r *http.Request, pubKey dynamic.JWTPublicKey, rules ...rule) (*http.Request, int, error) {
	status, claims, err := auth.verifyPubKey(r, pubKey.Key, pubKey.Name)
	if err != nil {
		return r, status, fmt.Errorf("error verifying public key: %w: %v", ErrPubKeyVerification, err)
	}
	status2, err := auth.verifyRules(r, claims, rules...)
	if err != nil {
		return r, status2, fmt.Errorf("error verifying rules: %w: %v", ErrRulesVerification, err)
	}
	// must write claims into the context - other services depend on this value
	return r.WithContext(NewContext(r.Context(), claims)), http.StatusOK, nil
}

// verifyPubKey verifies the request against a single JWT public key
// It returns recommended status code, JWT-claims object, and error
// It does not write claims into context - this is a read-only function - ensure to write claims into context when status code is 200
func (auth *Authenticator) verifyPubKey(r *http.Request, pubKey *rsa.PublicKey, keyName string) (int, *Claims, error) {
	if pubKey == nil {
		err := fmt.Errorf("public key missing")
		_ = auth.logger.ErrUserAuth(r.Context(), errors.Wrap(err, ErrMsgVerifier))
		return http.StatusInternalServerError, nil, err
	}
	rawToken, err := jwtReq.OAuth2Extractor.ExtractToken(r)
	if err != nil {
		_ = auth.logger.ErrUserAuth(r.Context(), errors.Wrap(err, ErrMsgVerifier))
		return http.StatusUnauthorized, nil, fmt.Errorf("cannot extract token from request")
	}

	parsedToken, err := jwt.ParseWithClaims(rawToken, &Claims{},
		func(_ *jwt.Token) (interface{}, error) {
			return pubKey, nil
		})
	if err != nil {
		_ = auth.logger.ErrUserAuth(r.Context(), errors.Wrap(err, ErrMsgVerifier))
		return http.StatusUnauthorized, nil, fmt.Errorf("cannot parse token")
	}

	if !parsedToken.Valid {
		_ = auth.logger.ErrUserAuth(r.Context(), errors.Wrap(ErrInvalidToken, ErrMsgVerifier))
		return http.StatusBadRequest, nil, fmt.Errorf("token invalid")
	}

	claims, ok := parsedToken.Claims.(*Claims)
	if !ok {
		_ = auth.logger.ErrUserAuth(r.Context(), errors.Wrap(ErrNoClaims, ErrMsgVerifier))
		return http.StatusBadRequest, nil, fmt.Errorf("cannot understand claims")
	}
	return http.StatusOK, claims, nil
}

// verifyRules verifies against the set of rules
func (auth *Authenticator) verifyRules(r *http.Request, claims *Claims, rules ...rule) (int, error) {
	if claims != nil {
		r = r.WithContext(NewContext(r.Context(), claims))
	}
	for _, rule := range rules {
		if err := rule(r); err != nil {
			_ = auth.logger.ErrUserAuth(r.Context(), errors.Wrap(err, ErrMsgVerifier))
			return http.StatusUnauthorized, fmt.Errorf("rule verification failed: %w", err)
		}
	}
	return http.StatusOK, nil
}

// WithOwner verifies that the given function returns the UUID of the JWT's subject ID.
func WithOwner(owner func(r *http.Request) uuid.UUID) rule {
	return func(r *http.Request) error {
		claims, ok := fromContext(r.Context())
		if !ok {
			return ErrNoClaims
		}

		haveID := owner(r)
		if haveID != claims.Subject.ID || haveID == uuid.Nil {
			return ErrSubjectNotOwner
		}

		return nil
	}
}

// WithGorillaOwner provides a Gorilla/Mux specific solution for parsing the owner from the path.
// This logic started to be replicated all around the services and was the initial reason for
// adding the middleware package.
func WithGorillaOwner(ownerKey string) rule {
	return func(r *http.Request) error {
		claims, ok := fromContext(r.Context())
		if !ok {
			return ErrNoClaims
		}

		vars := mux.Vars(r)
		value, ok := vars[ownerKey]
		if !ok {
			return ErrSubjectNotOwner
		}

		haveID, err := uuid.FromString(value)
		if err != nil || haveID != claims.Subject.ID || haveID == uuid.Nil {
			return ErrSubjectNotOwner
		}

		return nil
	}
}

// WithChiOwner provides a Chi/Mux specific solution for parsing the owner from the path.
// It returns a function that verifies that the owner from the path matches the subject ID from the JWT.
func WithChiOwner(ownerKey string) rule {
	return func(r *http.Request) error {
		claims, ok := fromContext(r.Context())
		if !ok {
			return ErrNoClaims
		}

		value := chi.URLParam(r, ownerKey)
		if value == "" {
			return ErrSubjectNotOwner
		}

		haveID, err := uuid.FromString(value)
		if err != nil || haveID != claims.Subject.ID || haveID == uuid.Nil {
			return ErrSubjectNotOwner
		}

		return nil
	}
}

// WithAnyScope verifies that at lest one of the given scopes is in the JWT.
func WithAnyScope(scopes ...string) rule {
	return func(r *http.Request) error {
		claims, ok := fromContext(r.Context())
		if !ok {
			return ErrNoClaims
		}

		for _, scope := range scopes {
			if claims.Scope.Contains(scope) {
				return nil
			}
		}

		return ErrMissingScope
	}
}

// WithAllScopes verifies that all the given scopes are in the JWT.
func WithAllScopes(scopes ...string) rule {
	return func(r *http.Request) error {
		claims, ok := fromContext(r.Context())
		if !ok {
			return ErrNoClaims
		}

		for _, scope := range scopes {
			if !claims.Scope.Contains(scope) {
				return ErrMissingScope
			}
		}

		return nil
	}
}

// WithScopes verifies that the given scopes are in the JWT.
// DEPRECATED!
func WithScopes(scopes ...string) rule {
	return WithAllScopes(scopes...)
}
