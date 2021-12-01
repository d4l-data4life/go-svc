package jwt

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

func TestScope_Contains(t *testing.T) {
	for _, tc := range []struct {
		name             string
		s                Scope
		t                string
		containsExpected bool
	}{
		{
			name:             "zero value scope does contain a zero value token",
			containsExpected: true,
		},
		{
			name:             "zero value scope does not contain a token",
			t:                "foo",
			containsExpected: false,
		},
		{
			name: "zero value scope token slice does contain zero value token",
			s: Scope{
				Tokens: []string{},
			},
			containsExpected: true,
		},
		{
			name: "zero value scope token slice does not contain a token",
			s: Scope{
				Tokens: []string{},
			},
			t:                "foo",
			containsExpected: false,
		},
		{
			name: "token is contained",
			s: Scope{
				Tokens: []string{
					"foo",
				},
			},
			t:                "foo",
			containsExpected: true,
		},
		{
			name: "token is contained in multiple scope tokens",
			s: Scope{
				Tokens: []string{
					"foo",
					"bar",
				},
			},
			t:                "foo",
			containsExpected: true,
		},
		{
			name: "token is not contained in multiple scope tokens",
			s: Scope{
				Tokens: []string{
					"foo",
					"bar",
				},
			},
			t:                "baz",
			containsExpected: false,
		},
		{
			name: "zero value token is not contained in multiple scope tokens",
			s: Scope{
				Tokens: []string{
					"foo",
					"bar",
				},
			},
			containsExpected: false,
		},
	} {
		tc := tc // Pin Variable
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.s.Contains(tc.t); got != tc.containsExpected {
				t.Errorf("expected contained token %t, got %t", tc.containsExpected, got)
			}
		})
	}
}

func TestScope_ContainsCodeGrantScope(t *testing.T) {
	for _, tc := range []struct {
		name             string
		s                Scope
		r                Scope
		containsExpected bool
	}{
		{
			name:             "zero value scope does contain a zero value token",
			containsExpected: true,
		},
		{
			name: "zero value doesn't contain multiple tokens scope",
			r: Scope{
				Tokens: []string{
					"foo",
					"bar",
				},
			},
			containsExpected: false,
		},
		{
			name: "multiple tokens scope contain multiple tokens scope",
			s: Scope{
				Tokens: []string{
					"foo",
					"bar",
					"baz",
				},
			},
			r: Scope{
				Tokens: []string{
					"foo",
					"bar",
				},
			},
			containsExpected: true,
		},
		{
			name: "multiple tokens scope doesn't contain larger multiple tokens scope",
			s: Scope{
				Tokens: []string{
					"foo",
					"bar",
				},
			},
			r: Scope{
				Tokens: []string{
					"foo",
					"bar",
					"baz",
				},
			},
			containsExpected: false,
		},
		{
			name: "single token scope doesn't contain multiple tokens scope",
			s: Scope{
				Tokens: []string{
					"foo",
				},
			},
			r: Scope{
				Tokens: []string{
					"foo",
					"bar",
				},
			},
			containsExpected: false,
		},
		{
			name: "multiple tokens scope doesn't contain subset multiple tokens scope",
			s: Scope{
				Tokens: []string{
					"foo",
					"bar",
				},
			},
			r: Scope{
				Tokens: []string{
					"foo",
					"baz",
				},
			},
			containsExpected: false,
		},
		{
			name: "multiple tokens scope contains same set tokens scope",
			s: Scope{
				Tokens: []string{
					"foo",
					"bar",
				},
			},
			r: Scope{
				Tokens: []string{
					"foo",
					"bar",
				},
			},
			containsExpected: true,
		},
		{
			name: "custom tags are valid",
			s: Scope{
				Tokens: []string{
					"tag:*",
					"rec:r",
					"user:r",
				},
			},
			r: Scope{
				Tokens: []string{
					"tag:6b85d727a01eab1f27e763d757ac5d9465a17f5d5cb56f622d",
					"rec:r",
					"user:r",
				},
			},
			containsExpected: true,
		},
		{
			name: "extended tokens are valid",
			s: Scope{
				Tokens: []string{
					"ext:labOrder:r",
					"ext:labOrder:w",
					"usr:w",
					"perm:w",
				},
			},
			r: Scope{
				Tokens: []string{
					"ext:labOrder:r",
					"ext:labOrder:w",
					"usr:w",
					"perm:w",
				},
			},
			containsExpected: true,
		},
	} {
		tc := tc // Pin Variable
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.s.ContainsCodeGrantScope(tc.r); got != tc.containsExpected {
				t.Errorf("expected contained token %t, got %t", tc.containsExpected, got)
			}
		})
	}
}

func TestScope_MarshalJSON(t *testing.T) {
	for _, tc := range []struct {
		name     string
		scope    Scope
		expected string
	}{
		{
			name:     "marshals zero value",
			scope:    Scope{},
			expected: `""`,
		},
		{
			name: "marshals single scope",
			scope: Scope{
				Tokens: []string{TokenUserRead},
			},
			expected: `"user:r"`,
		},
		{
			name: "marshals multi scope",
			scope: Scope{
				Tokens: []string{
					TokenUserRead,
					TokenPermissionsRead,
				},
			},
			expected: `"user:r perm:r"`,
		},
		{
			name: "marshals tags",
			scope: Scope{
				Tokens: []string{
					TokenUserRead,
					Tag("bar").String(),
				},
			},
			expected: `"user:r tag:bar"`,
		},
		{
			name: "marshals extended scopes",
			scope: Scope{
				Tokens: []string{
					TokenUserRead,
					ExtendedToken("labOrder:r").String(),
				},
			},
			expected: `"user:r ext:labOrder:r"`,
		},
	} {
		tc := tc // Pin Variable
		t.Run(tc.name, func(t *testing.T) {
			dst, err := json.Marshal(tc.scope)
			if err != nil {
				t.Errorf("expected no err, got %v", err)
			}

			if valid := json.Valid(dst); !valid {
				t.Errorf("expected valid json")
			}

			if string(dst) != tc.expected {
				t.Errorf("expected `%s`, got `%s`", tc.expected, dst)
			}
		})
	}
}

func TestScope_UnmarshalJSON(t *testing.T) {
	type checkFunc func(Scope, error) error
	checks := func(fns ...checkFunc) []checkFunc { return fns }

	hasSomeError := func(_ Scope, have error) error {
		if have == nil {
			return fmt.Errorf("expected error, found `%v`", have)
		}
		return nil
	}

	hasError := func(want error) checkFunc {
		return func(_ Scope, have error) error {
			if errors.Cause(have) != want {
				return fmt.Errorf("expected error `%v`, found `%v`", want, have)
			}
			return nil
		}
	}

	hasScope := func(expected Scope) checkFunc {
		return func(got Scope, _ error) error {
			if !reflect.DeepEqual(expected, got) {
				return fmt.Errorf("expected %q, got %q", expected, got)
			}
			return nil
		}
	}

	for _, tc := range [...]struct {
		name   string
		json   string
		checks []checkFunc
	}{
		{
			name: "invalid: empty string",
			json: ``,
			checks: checks(
				hasSomeError, // json lib does not return introspectable error
			),
		},
		{
			name: "invalid: empty JSON literal string",
			json: `""`,
			checks: checks(
				hasError(nil),
				hasScope(Scope{}),
			),
		},
		{
			name: "invalid tag delimiter are treated as unknown tag (ignored)",
			json: `"tag_bar"`,
			checks: checks(
				hasError(nil),
				hasScope(Scope{}),
			),
		},
		{
			name: "ignores invalid token: mix of valid and invalid tokens",
			json: `"foobar user:r"`,
			checks: checks(
				hasError(nil),
				hasScope(Scope{[]string{"user:r"}}),
			),
		},
		{
			name: "ignores invalid token: multiple invalid tokens",
			json: `"foobar barfoo"`,
			checks: checks(
				hasError(nil),
				hasScope(Scope{}),
			),
		},
		{
			name: "invalid: not a JSON string",
			json: `12`,
			checks: checks(
				hasSomeError,
			),
		},
		{
			name: "valid: JSON contains funky tag",
			json: `"tag:bar:baz"`,
			checks: checks(
				hasError(nil),
				hasScope(Scope{[]string{
					Tag("bar:baz").String(),
				}}),
			),
		},
		{
			name: "valid: JSON contains tags",
			json: `"tag:bar"`,
			checks: checks(
				hasError(nil),
				hasScope(Scope{[]string{
					Tag("bar").String(),
				}}),
			),
		},
		{
			name: "valid: JSON contains valid tokens and tags",
			json: `"user:r tag:bar"`,
			checks: checks(
				hasError(nil),
				hasScope(Scope{[]string{
					TokenUserRead,
					Tag("bar").String(),
				}}),
			),
		},
		{
			name: "valid: JSON contains one token",
			json: `"user:r"`,
			checks: checks(
				hasError(nil),
				hasScope(Scope{[]string{TokenUserRead}}),
			),
		},
		{
			name: "valid: JSON contains tokens",
			json: `"user:r perm:r"`,
			checks: checks(
				hasError(nil),
				hasScope(Scope{[]string{
					TokenUserRead,
					TokenPermissionsRead,
				}}),
			),
		},
		{
			name: "valid: JSON contains one tag",
			json: `"tag:bar"`,
			checks: checks(
				hasError(nil),
				hasScope(Scope{[]string{
					Tag("bar").String(),
				}}),
			),
		},
		{
			name: "valid: JSON contains one extended token",
			json: `"ext:labOrder:r"`,
			checks: checks(
				hasError(nil),
				hasScope(Scope{[]string{
					ExtendedToken("labOrder:r").String(),
				}}),
			),
		},
	} {
		tc := tc // Pin Variable
		t.Run(tc.name, func(t *testing.T) {
			var s Scope
			err := json.Unmarshal([]byte(tc.json), &s)

			for _, check := range tc.checks {
				if e := check(s, err); e != nil {
					t.Error(e)
				}
			}
		})
	}
}

func TestRequestedScope(t *testing.T) {
	for i, tc := range []struct {
		scope, expectedScope string
		errExpected          bool
		expectedTokenLen     int
	}{
		{
			scope:       "foo",
			errExpected: true,
		},
		{
			scope:            "attachment:w",
			expectedScope:    "attachment:w",
			expectedTokenLen: 1,
		},
		{
			scope:            "user:r tag:foo",
			expectedScope:    "user:r tag:foo",
			expectedTokenLen: 2,
		},
		{
			scope:            "perm:r perm:w",
			expectedScope:    "perm:r perm:w",
			expectedTokenLen: 2,
		},
	} {
		actualScope, err := NewScope(tc.scope)
		if got := err != nil; got != tc.errExpected {
			t.Errorf("%d expected err %v, got %v, err %v", i, tc.errExpected, got, err)
			continue
		}

		if actualScope.String() != tc.expectedScope {
			t.Errorf("%d expected scope %q, got %q", i, tc.expectedScope, actualScope)
		}

		if l := len(actualScope.Tokens); l != tc.expectedTokenLen {
			t.Errorf("%d expected %d tokens, got %d, tokens %v", i, tc.expectedTokenLen, l, actualScope)
		}
	}
}

func TestScan(t *testing.T) {
	type checkFunc func(Scope, error) error
	checks := func(fns ...checkFunc) []checkFunc { return fns }

	hasToken := func(want string) checkFunc {
		return func(s Scope, _ error) error {
			for _, have := range s.Tokens {
				if have == want {
					return nil
				}
			}
			return fmt.Errorf("expected token %q, not found in scope", want)
		}
	}
	hasError := func(want error) checkFunc {
		return func(_ Scope, have error) error {
			if want != have {
				return fmt.Errorf("expected error `%v`, found `%v`", want, have)
			}
			return nil
		}
	}

	for _, tc := range [...]struct {
		name   string
		v      interface{}
		checks []checkFunc
	}{
		{
			"one valid scope token",
			"user:q",
			checks(
				hasToken("user:q"),
				hasError(nil),
			),
		},
		{
			"two valid scope tokens",
			"user:q rec:r",
			checks(
				hasToken("user:q"),
				hasToken("rec:r"),
				hasError(nil),
			),
		},
		{
			"wrong type: int",
			12,
			checks(
				hasError(ErrScopeIsNotAString),
			),
		},
		{
			"wrong type: nil",
			nil,
			checks(
				hasError(ErrScopeIsNotAString),
			),
		},
		{
			"wrong type: byte slice",
			[]byte("user:q rec:r"),
			checks(
				hasError(ErrScopeIsNotAString),
			),
		},
	} {
		tc := tc // Pin Variable
		t.Run(tc.name, func(t *testing.T) {
			var s Scope
			err := s.Scan(tc.v)
			for _, check := range tc.checks {
				if e := check(s, err); e != nil {
					t.Error(e)
				}
			}
		})
	}
}

func TestValue(t *testing.T) {
	s, err := NewScope("rec:r attachment:w")
	if err != nil {
		t.Fatal(err)
	}

	v, err := s.Value()
	if err != nil {
		t.Errorf("no error expected, found %v", err)
	}

	str, ok := v.(string)

	if !ok {
		t.Fatalf("expected to marshal to string, found %q", reflect.TypeOf(v))
	}

	slice := strings.Split(str, " ")

	assertHasToken := func(want string) {
		for _, have := range slice {
			if have == want {
				return
			}
		}
		t.Errorf("expected token %q, not found in string", want)
	}

	assertHasToken("rec:r")
	assertHasToken("attachment:w")

	if have := len(slice); have != 2 {
		t.Errorf("expected the string to contain 2 tokens, found %d", have)
	}
}

func TestNew(t *testing.T) {
	t.Run("creates an empty scope", func(t *testing.T) {
		s, err := NewScope("")
		if err != nil {
			t.Errorf("no error expected, found %v", err)
		}

		if want, have := 0, len(s.Tokens); have != want {
			t.Errorf("expected %d tokens, found %d", want, have)
		}
	})
}

func TestScope_FromCodeGrantTokens(t *testing.T) {
	for i, tc := range []struct {
		scope            []string
		expectedScope    string
		errExpected      bool
		expectedTokenLen int
	}{
		{
			scope:       []string{},
			errExpected: false,
		},
		{
			scope:       []string{"foo"},
			errExpected: true,
		},
		{
			scope:            []string{"rec:r", "user:r", "user:q"},
			expectedScope:    "rec:r user:r user:q",
			errExpected:      false,
			expectedTokenLen: 3,
		},
		{
			scope:            []string{"rec:r"},
			expectedScope:    "rec:r",
			expectedTokenLen: 1,
		},
		{
			scope:            []string{"user:r", "user:w"},
			expectedScope:    "user:r user:w",
			expectedTokenLen: 2,
		},
		{
			scope:            []string{"perm:r", "perm:w"},
			expectedScope:    "perm:r perm:w",
			expectedTokenLen: 2,
		},
		{
			scope:            []string{"user:r", "ext:labOrder:r"},
			expectedScope:    "user:r ext:labOrder:r",
			expectedTokenLen: 2,
		},
	} {
		actualScope, err := FromCodeGrantTokens(tc.scope)
		if got := err != nil; got != tc.errExpected {
			t.Errorf("%d expected err %v, got %v, err %v", i, tc.errExpected, got, err)
			continue
		}

		if actualScope.String() != tc.expectedScope {
			t.Errorf("%d expected scope %q, got %q", i, tc.expectedScope, actualScope)
		}

		if l := len(actualScope.Tokens); l != tc.expectedTokenLen {
			t.Errorf("%d expected %d tokens, got %d, tokens %v", i, tc.expectedTokenLen, l, actualScope)
		}
	}
}

func TestScope_Tags(t *testing.T) {
	for i, tc := range []struct {
		name         string
		scope        string
		expectedTags []Tag
		errExpected  bool
	}{
		{
			name:         "no TokenTag in Scope returns nil slice",
			scope:        "rec:r user:r user:q",
			expectedTags: nil,
			errExpected:  false,
		},
		{
			name:         "1 TokenTag in Scope returns slice of length 1",
			scope:        "user:r user:w tag:YeH/t6DDJl9m/PUMcZiIMkL7ykGIQxa5LQ==",
			expectedTags: []Tag{"YeH/t6DDJl9m/PUMcZiIMkL7ykGIQxa5LQ=="},
			errExpected:  false,
		},
		{
			name:         "2 TokenTag in Scope returns slice of length 2",
			scope:        "tag:YeH/t6DDJl9m/PUMcZiIMkL7ykGIQxa5LQ== tag:gdjaG33gDMoIycXWeno+b70NHqX/3reCPQ==",
			expectedTags: []Tag{"YeH/t6DDJl9m/PUMcZiIMkL7ykGIQxa5LQ==", "gdjaG33gDMoIycXWeno+b70NHqX/3reCPQ=="},
			errExpected:  false,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			actualScope, err := NewScope(tc.scope)
			if err != nil {
				t.Fatal(err)
			}

			actualTags, err := actualScope.Tags()
			if got := err != nil; got != tc.errExpected {
				t.Errorf("%d expected err %v, got %v, err %v", i, tc.errExpected, got, err)
				return
			}

			if !reflect.DeepEqual(actualTags, tc.expectedTags) {
				t.Errorf("%d expected tags %q, got %q", i, tc.expectedTags, actualTags)
			}
		})
	}
}

func TestScope_ExtendedToken(t *testing.T) {
	s, err := NewScope("user:r perm:r ext:labOrder:r ext:CovHub:magic")
	if err != nil {
		t.Fatal(err)
	}

	exts, _ := s.ExtendedToken()
	exps := []ExtendedToken{"labOrder:r", "CovHub:magic"}
	if !reflect.DeepEqual(exts, exps) {
		t.Errorf("expected: %s, have: %s", exps, exts)
	}
}
