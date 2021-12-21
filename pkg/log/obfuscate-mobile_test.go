package log_test

import (
	"regexp"
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/log"
)

func TestObfuscateMobile(t *testing.T) {
	t.Run("Test valid numbers with plus prefix obfuscation", func(t *testing.T) {
		tests := []struct {
			name   string
			mobile string
			regex  string
			want   string
		}{
			{"12 digits", "+491731234567", `\+.*`, "+491xxxxxx567"},
			{"Uneven length", "+1234567890111", `\+.*`, "+1234xxxxxx111"},
			{"8 digits", "+49173123", `\+.*`, "+49xxxx23"},
			{"Number in text", "Your registered phone number is +4911111100", `\+.*`, "Your registered phone number is +491xxxxx00"},
			{"Real use case 1", "http://sinchmock.com?msisdn=4911111100", "msisdn=.*", "http://sinchmock.com?msisdn=491xxxxx00"},
			{"Real use case 2", `"to":["4911111100"]`, `"to":\[".*?"\]`, `"to":["491xxxxx00"]`},
			{"Real use case 3", `"to":["4911111100", "4911111100"]`, `"to":\[".*?"\]`, `"to":["491xxxxx00", "491xxxxx00"]`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if have := log.ObfuscateMobile(tt.mobile, regexp.MustCompile(tt.regex)); have != tt.want {
					t.Errorf("expected %q to be %q, found %q", tt.mobile, tt.want, have)
				}
			})
		}
	})
}
