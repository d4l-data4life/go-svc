package jwt

import (
	"testing"
	"time"
)

func TestClaimValid(t *testing.T) {
	var (
		now    = time.Now()
		past   = now.Add(-5 * time.Minute)
		future = now.Add(5 * time.Minute)
	)

	for i, tc := range []struct {
		name        string
		claim       Claims
		errExpected bool
	}{
		{
			name:        "invalid zero value",
			claim:       Claims{},
			errExpected: true,
		},
		{
			name: "incomplete",
			claim: Claims{
				Issuer: "invalid",
			},
			errExpected: true,
		},
		{
			name: "incomplete, missing exp nbf",
			claim: Claims{
				Issuer: IssuerGesundheitscloud,
			},
			errExpected: true,
		},
		{
			name: "incomplete, missing nbf",
			claim: Claims{
				Issuer:     IssuerGesundheitscloud,
				Expiration: Time(past),
			},
			errExpected: true,
		},
		{
			name: "incomplete, missing iss",
			claim: Claims{
				Expiration: Time(future),
				NotBefore:  Time(past),
			},
			errExpected: true,
		},
		{
			name: "incomplete, missing iss nbf",
			claim: Claims{
				Issuer:     IssuerGesundheitscloud,
				Expiration: Time(future),
			},
			errExpected: true,
		},
		{
			name: "invalid not before",
			claim: Claims{
				Issuer:     IssuerGesundheitscloud,
				Expiration: Time(future),
				NotBefore:  Time(future),
			},
			errExpected: true,
		},
		{
			name: "complete, but expired",
			claim: Claims{
				Issuer:     IssuerGesundheitscloud,
				Expiration: Time(past),
				NotBefore:  Time(past),
			},
			errExpected: true,
		},
		{
			name: "complete, but not before",
			claim: Claims{
				Issuer:     IssuerGesundheitscloud,
				Expiration: Time(past),
				NotBefore:  Time(future),
			},
			errExpected: true,
		},
		{
			name: "complete, but invalid issuer",
			claim: Claims{
				Issuer:     "invalid",
				Expiration: Time(future),
				NotBefore:  Time(past),
			},
			errExpected: true,
		},
		{
			name: "valid",
			claim: Claims{
				Issuer:     IssuerGesundheitscloud,
				Expiration: Time(future),
				NotBefore:  Time(past),
			},
		},
		{
			name: "valid with iat in the past",
			claim: Claims{
				Issuer:     IssuerGesundheitscloud,
				Expiration: Time(future),
				NotBefore:  Time(past),
				IssuedAt:   Time(past),
			},
		},
		{
			name: "valid with iat in the future",
			claim: Claims{
				Issuer:     IssuerGesundheitscloud,
				Expiration: Time(future),
				NotBefore:  Time(past),
				IssuedAt:   Time(future),
			},
		},
	} {
		// Pin Variables
		i := i
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.claim.valid(now)
			if got := err != nil; got != tc.errExpected {
				t.Errorf("%d: expected err %v, got %v, err=%v", i, tc.errExpected, got, err)
			}
			t.Log(err)
		})
	}
}
