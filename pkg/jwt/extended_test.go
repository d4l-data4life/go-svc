package jwt

import "testing"

func TestNewExtendedToken(t *testing.T) {
	for _, td := range [...]struct {
		name        string
		token       string
		expectedErr error
	}{
		{
			name:        "happy path",
			token:       "ext:labOrder:r",
			expectedErr: nil,
		},
		{
			name:        "fails, its a tag!",
			token:       "tag:base64MAGIC==",
			expectedErr: ErrInvalidExtendedToken,
		},
		{
			name:        "fails, its a token!",
			token:       "user:r",
			expectedErr: ErrInvalidExtendedToken,
		},
		{
			name:        "fails, it's something!",
			token:       "all the people unite!",
			expectedErr: ErrInvalidExtendedToken,
		},
		{
			name:        "fails, it's empty",
			token:       "",
			expectedErr: ErrInvalidExtendedToken,
		},
	} {
		td := td
		t.Run(td.name, func(t *testing.T) {
			ext, err := NewExtendedToken(td.token)
			if err != td.expectedErr {
				t.Fatal(err)
			}
			if err != nil {
				return
			}

			if td.token != ext.String() {
				t.Errorf("want: %s, have: %s", td.token, ext.String())
			}
		})
	}
}

func TestIsExtendedToken(t *testing.T) {
	for _, td := range [...]struct {
		name     string
		src      string
		expected bool
	}{
		{
			name:     "happy path",
			src:      "ext:labOrder:r",
			expected: true,
		},
		{
			name:     "not ext:",
			src:      "labOrder:r",
			expected: false,
		},
		{
			name:     "just ext:",
			src:      "ext:",
			expected: false,
		},
		{
			name:     "too long",
			src:      "ext:abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
			expected: false,
		},
		{
			name:     "wrong chars",
			src:      "ext:<script>alert('gief me your CUP!')</script>",
			expected: false,
		},
	} {
		td := td
		t.Run(td.name, func(t *testing.T) {
			if td.expected != IsExtendedToken(td.src) {
				t.Errorf("want: %t, have: %t\n", td.expected, IsExtendedToken(td.src))
			}
		})
	}
}

func TestExtLen(t *testing.T) {
	if expectedLen := len(extPrefix + ":"); expectedLen != extPrefixLenWithColon {
		t.Errorf("want: %d, have: %d", expectedLen, extPrefixLenWithColon)
	}
}
