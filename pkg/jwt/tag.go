package jwt

import (
	"regexp"

	"github.com/pkg/errors"
)

const (
	TagPrefix             = "tag"
	tagPrefixLenWithColon = 4
	tagRegex              = `^tag:.*$`
)

var (
	// ErrInvalidTag is returned when the parsed tag is not in the whitelist.
	ErrInvalidTag = errors.New("invalid tag")
	validTagToken = regexp.MustCompile(tagRegex)
)

// Tag is a scope for an encrypted tag
type Tag string

func (g Tag) String() string {
	return TagPrefix + ":" + string(g)
}

// NewTag creates a tag and validates it
func NewTag(src string) (Tag, error) {
	if !IsTag(src) {
		return "", ErrInvalidTag
	}

	return Tag(src[tagPrefixLenWithColon:]), nil
}

// IsTag returns true if the token is a Tag. It doesn't
// tell if the Tag is valid or not.
func IsTag(token string) bool {
	return validTagToken.MatchString(token)
}
