package jwt

import "testing"

func TestEncryptedTag(t *testing.T) {
	var s Scope
	s.Tokens = append(s.Tokens, TokenPermissionsRead, Tag("bar").String())

	if expected := TokenPermissionsRead + " tag:bar"; s.String() != expected {
		t.Errorf("expected scope %q, got %q", expected, s.String())
	}
}

func TestIsTag(t *testing.T) {
	for _, td := range [...]struct {
		name  string
		token string
		isTag bool
	}{
		{
			name:  "happy path",
			token: "tag:BgsIBQ0HAgcKAAEOCgsBDwIHDgcGAw0HBQcKDAUNCQQGBQoBBw8FDQUMCwUGDwYCAg0=",
			isTag: true,
		},
		{
			name:  "should return true on tag prefix, but invalid data",
			token: "tag:javascript:alert(\"boom, hacked!\")",
			isTag: true,
		},
		{
			name:  "should return false on valid token, but not a tag",
			token: TokenUserRead,
			isTag: false,
		},
		{
			name:  "should return false on empty string",
			token: "",
			isTag: false,
		},
		{
			name:  "should return false on random data",
			token: "javascript:alert(\"boom, hacked\")",
			isTag: false,
		},
	} {
		td := td
		t.Run(td.name, func(tc *testing.T) {
			if td.isTag != IsTag(td.token) {
				tc.Errorf("want: %t, have: %t", td.isTag, IsTag(td.token))
			}
		})
	}
}

func TestTagPrefixLen(t *testing.T) {
	if expectedLen := len(TagPrefix + ":"); expectedLen != tagPrefixLenWithColon {
		t.Errorf("want: %d, have: %d", expectedLen, tagPrefixLenWithColon)
	}
}
