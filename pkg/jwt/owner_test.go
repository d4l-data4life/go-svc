package jwt

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
)

func TestOwner_MarshalJSON(t *testing.T) {
	t.Run("marshals zero value", func(t *testing.T) {
		var o Owner
		_, err := json.Marshal(o)
		if err != nil {
			t.Errorf("expected no err, got %v", err)
		}
	})

	t.Run("marshals", func(t *testing.T) {
		id, err := uuid.FromString("11111111-1111-1111-1111-111111111111")
		if err != nil {
			t.Fatal("couldn't generate id", err)
		}

		o := Owner{
			ID: id,
		}

		dst, err := json.Marshal(o)
		if err != nil {
			t.Errorf("expected no err, got %v", err)
		}

		if valid := json.Valid(dst); !valid {
			t.Errorf("expected valid json")
		}

		if expected := `"owner:11111111-1111-1111-1111-111111111111"`; string(dst) != expected {
			t.Errorf("expected `%s`, got `%s`", expected, dst)
		}
	})

}

func TestOwner_UnmarshalJSON(t *testing.T) {
	type checkFunc func(Owner, error) error
	checks := func(fns ...checkFunc) []checkFunc { return fns }

	hasID := func(want string) checkFunc {
		return func(o Owner, _ error) error {
			if have := o.ID.String(); have != want {
				return fmt.Errorf("expected %s, found %s", want, have)
			}
			return nil
		}
	}

	hasError := func(want error) checkFunc {
		return func(_ Owner, have error) error {
			if have != want {
				return fmt.Errorf("expected error `%v`, found `%v`", want, have)
			}
			return nil
		}
	}

	hasSomeError := func(_ Owner, have error) error {
		if have == nil {
			return fmt.Errorf("expected error, found `%v`", have)
		}
		return nil
	}

	for _, tc := range [...]struct {
		name   string
		j      string
		checks []checkFunc
	}{
		{
			"unmarshals a namespaced uuid in canonical form",
			`"owner:11111111-1111-1111-1111-111111111111"`,
			checks(
				hasError(nil),
				hasID("11111111-1111-1111-1111-111111111111"),
			),
		},
		{
			"unmarshals a namespaced uuid",
			`"owner:11111121111111111111111111111111"`,
			checks(
				hasError(nil),
				hasID("11111121-1111-1111-1111-111111111111"),
			),
		},
		{
			"errors if namespace is missing",
			`"11111111-1111-1111-1111-111111111111"`,
			checks(
				hasSomeError,
			),
		},
		{
			"errors if invalid UUIDv4",
			`"11111111-1111-1111-1111-z11111111111"`,
			checks(
				hasSomeError,
			),
		},
		{
			"errors if ID is missing",
			`"owner:"`,
			checks(
				hasSomeError,
			),
		},
		{
			"errors if invalid JSON",
			`owner:`,
			checks(
				hasSomeError,
			),
		},
		{
			"errors if empty JSON string",
			`""`,
			checks(
				hasSomeError,
			),
		},
		{
			"errors if JSON number",
			`3`,
			checks(
				hasSomeError,
			),
		},
	} {
		tc := tc // Pin Variable
		t.Run(tc.name, func(t *testing.T) {
			var o Owner
			err := json.Unmarshal([]byte(tc.j), &o)

			for _, check := range tc.checks {
				if e := check(o, err); e != nil {
					t.Error(e)
				}
			}
		})
	}
}
