package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	jwtgo "github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestCreateAccessToken(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	skew := time.Minute

	someUserID := uuid.Must(uuid.NewV4())
	someAppID := uuid.Must(uuid.NewV4())

	tests := []struct {
		name    string
		opt     []TokenOption
		want    Claims
		wantErr bool
	}{
		{
			name: "success",
			opt: []TokenOption{
				WithUserID(someUserID),
				WithAppID(someAppID),
				WithClientID("someClientID"),
				WithTenantID("d4l"),
				WithEMail("we@data4life.care"),
				WithScope(Scope{Tokens: []string{TokenUserRead, TokenUserKeysRead}}),
				WithExpirationTime(now.Add(time.Minute * 5)),
			},
			want: Claims{
				Issuer:     IssuerGesundheitscloud,
				Subject:    Owner{someUserID},
				UserID:     someUserID,
				AppID:      someAppID,
				TenantID:   "d4l",
				ClientID:   "someClientID",
				Email:      "we@data4life.care",
				Scope:      Scope{Tokens: []string{TokenUserRead, TokenUserKeysRead}},
				IssuedAt:   Time(now),
				NotBefore:  Time(now.Add(-skew)),
				Expiration: Time(now.Add(time.Minute * 5).Add(skew)),
			},
			wantErr: false,
		},
		{
			name: "success - only expiration duration set",
			opt: []TokenOption{
				WithExpirationDuration(time.Minute * 5),
			},
			want: Claims{
				Issuer:     IssuerGesundheitscloud,
				IssuedAt:   Time(now),
				NotBefore:  Time(now.Add(-skew)),
				Expiration: Time(now.Add(time.Minute * 5).Add(skew)),
			},
			wantErr: false,
		},
		{
			name: "success - with scope strings",
			opt: []TokenOption{
				WithExpirationDuration(time.Minute * 5),
				WithScopeStrings(TokenAttachmentsRead, TokenDeviceRead),
			},
			want: Claims{
				Issuer:     IssuerGesundheitscloud,
				Scope:      Scope{Tokens: []string{TokenAttachmentsRead, TokenDeviceRead}},
				IssuedAt:   Time(now),
				NotBefore:  Time(now.Add(-skew)),
				Expiration: Time(now.Add(time.Minute * 5).Add(skew)),
			},
			wantErr: false,
		},
		{
			name: "missing expiration",
			opt:  []TokenOption{},
			want: Claims{
				Issuer:    IssuerGesundheitscloud,
				IssuedAt:  Time(now),
				NotBefore: Time(now.Add(-skew)),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateAccessToken(key, tt.opt...)
			if err != nil {
				t.Fatal(err)
			}

			parsed, err := jwtgo.ParseWithClaims(got.AccessToken, &Claims{}, func(t *jwtgo.Token) (interface{}, error) {
				return &key.PublicKey, nil
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAccessToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			claims := parsed.Claims.(*Claims)
			equalClaims(t, tt.want, *claims)
		})
	}
}

func equalClaims(t *testing.T, want Claims, have Claims) {
	compareTimes(t, "issued-at", want.IssuedAt, have.IssuedAt)
	compareTimes(t, "not-before", want.NotBefore, have.NotBefore)
	compareTimes(t, "expiration", want.Expiration, have.Expiration)

	want.IssuedAt = Time{}
	have.IssuedAt = Time{}
	want.NotBefore = Time{}
	have.NotBefore = Time{}
	want.Expiration = Time{}
	have.Expiration = Time{}

	if have.JWTID == uuid.Nil {
		t.Errorf("JWTID must not be uuid.Nil")
	}
	have.JWTID = uuid.Nil

	assert.Equal(t, want, have)
}

func compareTimes(t *testing.T, name string, want Time, have Time) {
	tolerance := 1 * time.Second

	h := time.Time(have)
	w := time.Time(want)

	if h.Before(w.Add(-tolerance)) {
		t.Errorf("%s time have %s is before tolerated want %s", name, h.String(), w.Add(-tolerance).String())
	}
	if h.After(w.Add(tolerance)) {
		t.Errorf("%s time have %s is after tolerated want %s", name, h.String(), w.Add(tolerance).String())
	}
}
