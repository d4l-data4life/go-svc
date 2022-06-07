package tut

import (
	"github.com/gesundheitscloud/go-svc/pkg/jwt"
)

// AllScopesExcept returns all the known jwt scopes excluding
// the one passed as an argument
func AllScopesExcept(excludedScopes ...string) []string {
	var tokens []string
	for t := range jwt.KnownTokens {
		toExclude := false
		for _, e := range excludedScopes {
			if t == e {
				toExclude = true
				break
			}
		}
		if !toExclude {
			tokens = append(tokens, t)
		}
	}

	return tokens
}
