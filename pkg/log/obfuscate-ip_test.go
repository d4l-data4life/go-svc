package log_test

import (
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/log"
)

func TestObfuscateIP(t *testing.T) {
	t.Run("Test valid ip addresses obfuscation", func(t *testing.T) {
		tests := []struct {
			name string
			ip   string
			want string
		}{
			{"Valid 1", "127.0.0.1", "127.0.xxx.xxx"},
			{"Valid 2", "192.168.10.14", "192.168.xxx.xxx"},
			{"Invalid 1", "123.456.789.123", "123.456.789.123"},
			{"Invalid 2", "123.456.123", "123.456.123"},
			{
				"Multiple addresses",
				"Adr. 1 at: 123.5.7.123, adr.2 at 89.135.24.1",
				"Adr. 1 at: 123.5.xxx.xxx, adr.2 at 89.135.xxx.xxx"},
			{
				"Both valid and invalid inputs",
				"Valid address at 12.233.5.66 and an invalid address at 456.12.12.12",
				"Valid address at 12.233.xxx.xxx and an invalid address at 456.12.12.12"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if have := log.ObfuscateIP(tt.ip); have != tt.want {
					t.Errorf("expected %q to be %q, found %q", tt.ip, tt.want, have)
				}
			})
		}
	})
}
