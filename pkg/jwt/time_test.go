package jwt

import (
	"testing"
	"time"
)

func TestTime_MarshalJSON(t *testing.T) {
	mustParse := func(datetime string) time.Time {
		t, err := time.Parse(time.RFC3339, datetime)
		if err != nil {
			panic(err)
		}
		return t
	}

	for _, tc := range [...]struct {
		name     string
		t        time.Time
		expected string
		hasError bool
	}{
		{
			"encodes Epoch",
			mustParse("1970-01-01T00:00:00Z"),
			"0",
			false,
		},
		{
			"encodes time zero",
			time.Time{},
			"-62135596800",
			false,
		},
	} {
		tc := tc // Pin Variable
		t.Run(tc.name, func(t *testing.T) {
			b, err := Time(tc.t).MarshalJSON()
			if tc.hasError {
				if err == nil {
					t.Error("expected error, found nil")
				}
				return
			}

			if have, want := string(b), tc.expected; have != want {
				t.Errorf("expected %q, found %q", want, have)
			}
		})
	}
}

func TestTime_UnmarshalJSON(t *testing.T) {
	for _, tc := range []struct {
		name        string
		src         []byte
		expected    Time
		errExpected bool
	}{
		{
			name:     "unmarshals successfully",
			src:      []byte("0"),
			expected: Time(time.Unix(0, 0)),
		},
		{
			name:        "unmarshaling fails",
			src:         []byte("a"),
			errExpected: true,
		},
	} {
		tc := tc // Pin Variable
		t.Run(tc.name, func(t *testing.T) {
			var got Time
			err := got.UnmarshalJSON(tc.src)
			if got := err != nil; got != tc.errExpected {
				t.Errorf("expected err %v, got %v, err %v", tc.errExpected, got, err)
			}

			if got != tc.expected {
				t.Errorf("got %s, expected %s", time.Time(got), time.Time(tc.expected))
			}
		})
	}
}
