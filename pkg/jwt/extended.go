package jwt

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

const (
	extPrefix             = "ext"
	extPrefixLenWithColon = 4
	// starts with ext: and afterwards it has 1 to 32 alphabetical chars or colon.
	extRegex = `^ext:[a-zA-Z:]{1,32}$`
)

var (
	// ErrInvalidExtendedToken is returned when the parsed token is not an ExtendedToken.
	ErrInvalidExtendedToken = errors.New("token is not an Extended Token")
	validExtendedToken      = regexp.MustCompile(extRegex)
)

// ExtendedToken is a Token that is used for Tokens that should not be checked by KnownTokens.
type ExtendedToken string

// String returns the string representation of an ExtendedToken.
func (e ExtendedToken) String() string {
	var builder strings.Builder

	builder.WriteString(extPrefix)
	builder.WriteString(":")
	builder.WriteString(string(e))

	return builder.String()
}

// NewExtendedToken transforms a string that satisfies the specification of an ExtendedToken into an ExtendedToken.
func NewExtendedToken(token string) (ExtendedToken, error) {
	if !IsExtendedToken(token) {
		return "", ErrInvalidExtendedToken
	}

	return ExtendedToken(token[extPrefixLenWithColon:]), nil
}

// IsExtendedToken checks if the given string satisfies the specification of an ExtendedToken.
func IsExtendedToken(src string) bool {
	return validExtendedToken.MatchString(src)
}
