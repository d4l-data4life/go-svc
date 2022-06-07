package tut

import (
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/gesundheitscloud/go-svc/pkg/jwt"

	"github.com/gofrs/uuid"
	jwtgo "github.com/golang-jwt/jwt/v4"
)

type JWTCheckFunc func(jwt *jwtgo.Token) error

func CreateJWTChecker(pubKeys ...*rsa.PublicKey) func(jwtChecks ...JWTCheckFunc) func(token interface{}) error {
	return func(jwtChecks ...JWTCheckFunc) func(token interface{}) error {
		errJWTPublicKeyMatch := errors.New("the public key doesn't match")

		testFunc := func(pubKey *rsa.PublicKey, t string) error {
			token, err := jwtgo.ParseWithClaims(
				t,
				&jwt.Claims{},
				func(*jwtgo.Token) (interface{}, error) {
					return pubKey, nil
				},
			)
			if err != nil {
				return fmt.Errorf("%w: jwt.parseWithClaims failed: %v", errJWTPublicKeyMatch, err)
			}

			for _, check := range jwtChecks {
				if err := check(token); err != nil {
					return fmt.Errorf("access token check failed: %w", err)
				}
			}
			return nil
		}

		return func(token interface{}) error {
			var err error
			t, ok := token.(string)
			if !ok {
				return fmt.Errorf("token is not of the expected type; want string, got: %v", token)
			}
			for _, key := range pubKeys {
				err = testFunc(key, t)
				if err == nil {
					return nil
				}
				if !errors.Is(err, errJWTPublicKeyMatch) {
					// an actual JWT check failed, stop immediately
					return err
				}
			}
			// if all public keys produce public keys match errors, then return any error as result
			return err
		}
	}
}

func HasTenantID(tenantID string) JWTCheckFunc {
	return func(t *jwtgo.Token) error {
		claims := t.Claims.(*jwt.Claims)
		if claims.TenantID != tenantID {
			return fmt.Errorf("unexpected tenant ID: want %s, have %s", tenantID, claims.TenantID)
		}

		return nil
	}
}

func HasScopeStrings(scopes ...string) JWTCheckFunc {
	return func(t *jwtgo.Token) error {
		claims := t.Claims.(*jwt.Claims)
		if len(claims.Scope.Tokens) != len(scopes) {
			return fmt.Errorf("unexpected scopes. want: %s, have: %s", scopes, claims.Scope.String())
		}
		for _, s := range scopes {
			if !claims.Scope.Contains(s) {
				return fmt.Errorf("scope %s is not contained in JWT", s)
			}
		}
		return nil
	}
}

func HasSubject(userID uuid.UUID) JWTCheckFunc {
	return func(t *jwtgo.Token) error {
		claims := t.Claims.(*jwt.Claims)
		if claims.Subject.ID != userID {
			return fmt.Errorf("unexpected subject: want %s, have %s", userID, claims.Subject)
		}
		return nil
	}
}
