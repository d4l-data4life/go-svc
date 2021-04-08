package bievents

import (
	"testing"
)

func TestGetEmailType(t *testing.T) {
	for _, tc := range [...]struct {
		name      string
		email     string
		result    EmailType
		wantError bool
	}{
		{
			name:      "should fail for empty email string",
			email:     "",
			wantError: true,
		},
		{
			name:      "should fail for invalid email string with 2 @",
			email:     "some@invalid@email.com",
			wantError: true,
		},
		{
			name:      "should return internal for data4life email",
			email:     "abc@data4life.care",
			result:    Internal,
			wantError: false,
		},
		{
			name:      "should return internal for gesundheitscloud email",
			email:     "abc@gesundheitscloud.de",
			result:    Internal,
			wantError: false,
		},
		{
			name:      "should return internal for qamadness email",
			email:     "abc@qamadness.com",
			result:    Internal,
			wantError: false,
		},
		{
			name:      "should return internal for wearehackerone email",
			email:     "abc@wearehackerone.com",
			result:    Internal,
			wantError: false,
		},
		{
			name:      "should return internal for ghostinspector email",
			email:     "abc@ghostinspector.com",
			result:    Internal,
			wantError: false,
		},
		{
			name:      "should return external for gmail email",
			email:     "abc@gmail.com",
			result:    External,
			wantError: false,
		},
		{
			name:      "should return external for gmail email",
			email:     "abc@gmail.com",
			result:    External,
			wantError: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			have, err := GetEmailType(tc.email)

			if hasError := (err != nil); tc.wantError != hasError {
				t.Errorf("expected error to be %t found %t", tc.wantError, hasError)
			}

			if want := have; want != have {
				t.Errorf("expected email to be %s found %s", want, have)
			}
		})
	}
}
