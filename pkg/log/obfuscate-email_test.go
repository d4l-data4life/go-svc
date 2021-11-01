package log_test

import (
	"testing"

	"github.com/gesundheitscloud/go-svc/pkg/log"
)

func TestObfuscateEmail(t *testing.T) {
	t.Run("Test local part obfuscation", func(t *testing.T) {
		tests := []struct {
			name  string
			email string
			want  string
		}{
			{"Short name", "name@firstname.lastname.com", "naxx@firstxxxxxxxxxname.com"},
			{"Short name 2", "name2@firstname.lastname.com", "namxx@firstxxxxxxxxxname.com"},
			{"Very short name", "a@firstname.lastname.com", "a@firstxxxxxxxxxname.com"},
			{"Very short name 2", "ab@firstname.lastname.com", "ax@firstxxxxxxxxxname.com"},
			{"Special chars", ".!#$%&'*+-/=?^_`{|}~%@firstname.lastname.com", ".!#$%&'*+-/xxxxxxxxxx@firstxxxxxxxxxname.com"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if have := log.ObfuscateEmail(tt.email); have != tt.want {
					t.Errorf("expected %q to be %q, found %q", tt.email, tt.want, have)
				}
			})
		}
	})

	t.Run("Test domain part obfuscation", func(t *testing.T) {
		tests := []struct {
			name  string
			email string
			want  string
		}{
			{"Short domain 2 chars", "firstname.lastname@me.com", "firstnamexxxxxxxxx@mx.com"},
			{"Short domain 3 chars", "firstname.lastname@mee.com", "firstnamexxxxxxxxx@mxe.com"},
			{"Short domain 4 chars", "firstname.lastname@meee.com", "firstnamexxxxxxxxx@mxxe.com"},
			{"Short domain 5 chars", "firstname.lastname@meeee.com", "firstnamexxxxxxxxx@mexxe.com"},
			{"Numbers", "firstname.lastname@1234567890.com", "firstnamexxxxxxxxx@123xxxxx90.com"},
			{"Unusual domain", "firstname.lastname@a.b.c.d.e.f.com", "firstnamexxxxxxxxx@a.bxxxxxe.f.com"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if have := log.ObfuscateEmail(tt.email); have != tt.want {
					t.Errorf("expected %q to be %q, found %q", tt.email, tt.want, have)
				}
			})
		}
	})

	t.Run("Test obfuscation", func(t *testing.T) {
		tests := []struct {
			name  string
			email string
			want  string
		}{
			{"Confluence example", "firstname.lastname@firstname.lastname.com", "firstnamexxxxxxxxx@firstxxxxxxxxxname.com"},
			{
				"Text with multiple mails",
				"Email 1 is firstname.lastname@firstname.lastname.com and Email 2 is lastname.firstname@firstname.lastname.com",
				"Email 1 is firstnamexxxxxxxxx@firstxxxxxxxxxname.com and Email 2 is lastname.xxxxxxxxx@firstxxxxxxxxxname.com",
			},
			{"Not an email", "firstname.lastname@", "firstname.lastname@"},
			{"Not an email 2", "firstname.lastname@.com", "firstname.lastname@.com"},
			{"Not an emai 3l", "@firstname.lastname.com", "@firstname.lastname.com"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if have := log.ObfuscateEmail(tt.email); have != tt.want {
					t.Errorf("expected %q to be %q, found %q", tt.email, tt.want, have)
				}
			})
		}
	})
}
