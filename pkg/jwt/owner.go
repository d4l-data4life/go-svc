package jwt

import (
	"encoding/json"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

var (
	// ErrInvalidOwner is returned if the owner is invalid.
	ErrInvalidOwner = errors.New("invalid owner")

	// ErrSubjectNotOwner is returned if the subject is not an owner.
	ErrSubjectNotOwner = errors.New("subject is not an owner")
)

// Owner is the struct that references an owner user ID
type Owner struct {
	ID uuid.UUID
}

// MarshalJSON converts the owner to a JSON representation, prefixed with "owner:"
func (o Owner) MarshalJSON() ([]byte, error) {
	return []byte(`"owner:` + o.ID.String() + `"`), nil
}

// UnmarshalJSON parses the given src slice and extracts the owner from it.
func (o *Owner) UnmarshalJSON(src []byte) error {
	var s string
	if err := json.Unmarshal(src, &s); err != nil {
		return errors.Wrap(err, "error unmarshaling owner")
	}

	sub := strings.SplitN(s, `:`, 2)
	if len(sub) != 2 {
		return ErrInvalidOwner
	}

	if sub[0] != "owner" {
		return ErrSubjectNotOwner
	}

	id, err := uuid.FromString(sub[1])
	if err != nil {
		return errors.Wrap(err, "error parsing owner uuid")
	}

	o.ID = id

	return nil
}
